package main

import (
	"context"
	"flag" // 导入 flag 包
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MicroService/internal/time-service"
	"MicroService/internal/time-service/config"
	"MicroService/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1. 定义命令行参数
	// 为每个需要从命令行配置的字段定义一个 flag
	portFlag := flag.Int("port", 0, "The port for the time-service to listen on. Overrides config.")
	registryAddrFlag := flag.String("registry-addr", "", "The address of the registry. Overrides config.")
	serviceIPFlag := flag.String("ip", "", "The service's IP address. Overrides config.")
	flag.Parse() // 解析命令行参数

	// 2. 加载配置（首先从环境变量和默认值）
	// LoadTimeServiceConfig 函数会从环境变量加载配置
	cfg := config.LoadTimeServiceConfig()

	// 3. 使用命令行参数覆盖配置
	// 如果命令行参数被指定，则使用它的值
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *registryAddrFlag != "" {
		cfg.RegistryAddr = *registryAddrFlag
	}
	if *serviceIPFlag != "" {
		cfg.ServiceHostIP = *serviceIPFlag
	}

	logrus.Infof("Starting Time-Service on port %d...", cfg.Port)

	// 获取服务 IP 地址
	var currentIPAddress string
	if cfg.ServiceHostIP != "" {
		currentIPAddress = cfg.ServiceHostIP
		logrus.Infof("Using configured IP address: %s", currentIPAddress)
	} else {
		var err error
		currentIPAddress, err = util.GetLocalIP()
		if err != nil {
			logrus.Fatalf("Failed to get local IP address: %v", err)
		}
		logrus.Infof("Auto-detected IP address: %s", currentIPAddress)
	}

	// 注册服务
	serviceName := "time-service"
	serviceId, err := timeservice.RegisterService(cfg.RegistryAddr, serviceName, currentIPAddress, cfg.Port)
	if err != nil {
		logrus.Fatalf("Failed to register service: %v", err)
	}
	logrus.Infof("Service '%s' (ID: %s) registered successfully.", serviceName, serviceId)

	// 启动心跳
	stopHeartbeatChan := timeservice.StartHeartbeat(cfg.RegistryAddr, serviceId, currentIPAddress, cfg.Port, cfg.HeartbeatInterval)

	// 初始化 Gin 路由
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceId)
		c.Next()
	})

	router.GET("/api/getDateTime", timeservice.DateTimeHandler)

	// 启动 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Gin server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down Time-Service...")
	close(stopHeartbeatChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Time-Service forced to shutdown: %v", err)
	}

	err = timeservice.UnregisterService(cfg.RegistryAddr, serviceName, serviceId, currentIPAddress, cfg.Port)
	if err != nil {
		logrus.Errorf("Failed to unregister service during shutdown: %v", err)
	}

	logrus.Info("Time-Service stopped gracefully.")
}
