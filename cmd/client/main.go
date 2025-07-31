package main

import (
	"MicroService/pkg/util"
	"context"
	"flag" // 依然保留，虽然此处不用于文件路径，但可以在未来用于其他命令行参数
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MicroService/internal/client"        // 导入 client 业务逻辑包
	"MicroService/internal/client/config" // 导入你的配置包
	"MicroService/pkg/httpclient"         // 导入 HTTP 客户端

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus" // 确保你的日志库是 logrus
)

func main() {
	flag.Parse() // 如果没有命令行参数，这行代码仍然是无害的

	// 1. 加载配置 (直接从环境变量)
	cfg := config.LoadConfig()

	// 2. 初始化日志
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("Client service starting with config: %+v", cfg)

	// 3. 初始化 HTTP 客户端
	httpClientConfig := httpclient.Config{
		Timeout:    cfg.HTTPClientTimeout,
		MaxRetries: cfg.HTTPClientMaxRetries,
		RetryDelay: cfg.HTTPClientRetryDelay,
	}
	httpClient := httpclient.NewClient(httpClientConfig)

	// 4. 服务注册
	// 使用配置文件中的 IP 地址，如果为空则尝试自动获取
	clientIP := cfg.IPAddress
	if clientIP == "" {
		localIP, err := util.GetLocalIP() // 假设 client 包提供这个函数
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

	// 5. 启动心跳
	heartbeatStopChan := client.StartHeartbeat(
		cfg.RegistryAddress,
		serviceID,
		clientIP,
		cfg.Port,
		cfg.HeartbeatInterval,
	)

	// 6. 配置 Gin 路由
	router := gin.Default()
	// 添加中间件，将服务 ID、注册中心地址和 HTTP 客户端注入到 Gin 上下文
	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceID)
		c.Set("registryAddr", cfg.RegistryAddress)
		c.Set("httpClient", httpClient) // 传递 HTTP 客户端实例
		c.Next()
	})

	router.GET("/api/getInfo", client.InfoHandler) // 客户端核心 API

	// 7. 启动 HTTP 服务器
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

	// 8. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // 阻塞直到接收到信号

	logrus.Info("Shutting down client service...")

	// 停止心跳 Goroutine
	close(heartbeatStopChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Client service forced to shutdown: %v", err)
	} else {
		logrus.Info("Client service stopped gracefully.")
	}

	// 9. 服务注销
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
