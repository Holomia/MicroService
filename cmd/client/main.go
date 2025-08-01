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

	"MicroService/internal/client"
	"MicroService/internal/client/config"
	"MicroService/pkg/httpclient"
	"MicroService/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1. 定义命令行参数
	portFlag := flag.Int("port", 0, "The port for the client service to listen on. Overrides config.")
	registryAddrFlag := flag.String("registry-addr", "", "The address of the registry. Overrides config.")
	serviceIPFlag := flag.String("ip", "", "The service's IP address. Overrides config.")
	flag.Parse() // 解析命令行参数

	// 2. 加载配置（首先从环境变量和默认值）
	cfg := config.LoadConfig()

	// 3. 使用命令行参数覆盖配置
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *registryAddrFlag != "" {
		cfg.RegistryAddress = *registryAddrFlag
	}
	if *serviceIPFlag != "" {
		cfg.IPAddress = *serviceIPFlag
	}

	// 4. 初始化日志
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("Client service starting with config: %+v", cfg)

	// 5. 初始化 HTTP 客户端
	httpClientConfig := httpclient.Config{
		Timeout:    cfg.HTTPClientTimeout,
		MaxRetries: cfg.HTTPClientMaxRetries,
		RetryDelay: cfg.HTTPClientRetryDelay,
	}
	httpClient := httpclient.NewClient(httpClientConfig)

	// 6. 服务注册
	clientIP := cfg.IPAddress
	if clientIP == "" {
		localIP, err := util.GetLocalIP()
		if err != nil {
			logrus.Fatalf("Failed to get local IP address for client service: %v", err)
		}
		clientIP = localIP
	}

	serviceID, err := client.RegisterService(
		cfg.RegistryAddress,
		cfg.ServiceName,
		clientIP,
		cfg.Port,
	)
	if err != nil {
		logrus.Fatalf("Failed to register client service to registry: %v", err)
	}
	logrus.Infof("Client service registered with ID: %s", serviceID)

	// 7. 启动心跳
	heartbeatStopChan := client.StartHeartbeat(
		cfg.RegistryAddress,
		serviceID,
		clientIP,
		cfg.Port,
		cfg.HeartbeatInterval,
	)

	// 8. 配置 Gin 路由
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceID)
		c.Set("registryAddr", cfg.RegistryAddress)
		c.Set("httpClient", httpClient)
		c.Next()
	})

	router.GET("/api/getInfo", client.InfoHandler)

	// 9. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logrus.Infof("Client service listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Client service listen error: %s", err)
		}
	}()

	// 10. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down client service...")
	close(heartbeatStopChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Client service forced to shutdown: %v", err)
	} else {
		logrus.Info("Client service stopped gracefully.")
	}

	// 11. 服务注销
	err = client.UnregisterService(
		cfg.RegistryAddress,
		cfg.ServiceName,
		serviceID,
		clientIP,
		cfg.Port,
	)
	if err != nil {
		logrus.Errorf("Failed to unregister client service: %v", err)
	} else {
		logrus.Infof("Client service %s unregistered successfully.", serviceID)
	}

	logrus.Info("Client service exited.")
}
