package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings" // 新增导入
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
	portFlag := flag.Int("port", 0, "The port for the time-service to listen on. Overrides config.")
	// 新增一个支持逗号分隔地址列表的参数
	registryAddrsFlag := flag.String("registry-addrs", "", "Comma-separated list of registry addresses (e.g., 'http://127.0.0.1:8180,http://127.0.0.1:8181'). Overrides config.")
	// 保留旧参数，但明确说明它将被覆盖
	registryAddrFlag := flag.String("registry-addr", "", "The address of the registry. Deprecated. Use --registry-addrs instead.")
	serviceIPFlag := flag.String("ip", "", "The service's IP address. Overrides config.")
	flag.Parse()

	// 2. 加载配置
	cfg := config.LoadTimeServiceConfig()

	// 3. 使用命令行参数覆盖配置
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *serviceIPFlag != "" {
		cfg.ServiceHostIP = *serviceIPFlag
	}

	// 4. 解析注册中心地址列表，优先级：新参数 > 旧参数 > 配置文件
	var registryAddrs []string
	if *registryAddrsFlag != "" {
		registryAddrs = strings.Split(*registryAddrsFlag, ",")
	} else if *registryAddrFlag != "" {
		registryAddrs = []string{*registryAddrFlag}
	} else if cfg.RegistryAddr != "" {
		registryAddrs = []string{cfg.RegistryAddr}
	}

	if len(registryAddrs) == 0 {
		logrus.Fatalf("No registry addresses provided.")
	}
	logrus.Infof("Registry addresses: %v", registryAddrs)

	logrus.Infof("Starting Time-Service on port %d...", cfg.Port)

	// 5. 获取服务 IP 地址
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

	// 6. 服务注册，传入地址列表
	serviceName := "time-service"
	serviceId, err := timeservice.RegisterService(registryAddrs, serviceName, currentIPAddress, cfg.Port)
	if err != nil {
		logrus.Fatalf("Failed to register service: %v", err)
	}
	logrus.Infof("Service '%s' (ID: %s) registered successfully.", serviceName, serviceId)

	// 7. 启动心跳，传入地址列表
	stopHeartbeatChan := timeservice.StartHeartbeat(registryAddrs, serviceId, currentIPAddress, cfg.Port, cfg.HeartbeatInterval)

	// 8. 初始化 Gin 路由
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceId)
		c.Next()
	})

	router.GET("/api/getDateTime", timeservice.DateTimeHandler)

	// 9. 启动 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Gin server error: %v", err)
		}
	}()

	// 10. 优雅关闭
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

	// 11. 服务注销，传入地址列表
	err = timeservice.UnregisterService(registryAddrs, serviceName, serviceId, currentIPAddress, cfg.Port)
	if err != nil {
		logrus.Errorf("Failed to unregister service during shutdown: %v", err)
	}

	logrus.Info("Time-Service stopped gracefully.")
}
