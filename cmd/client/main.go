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
	// 新增一个支持逗号分隔地址列表的参数
	registryAddrsFlag := flag.String("registry-addrs", "", "Comma-separated list of registry addresses (e.g., 'http://127.0.0.1:8180,http://127.0.0.1:8181'). Overrides config.")
	// 保留旧参数
	registryAddrFlag := flag.String("registry-addr", "", "The address of the registry. Deprecated. Use --registry-addrs instead.")
	serviceIPFlag := flag.String("ip", "", "The service's IP address. Overrides config.")
	flag.Parse()

	// 2. 加载配置
	cfg := config.LoadConfig()

	// 3. 使用命令行参数覆盖配置
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	if *serviceIPFlag != "" {
		cfg.IPAddress = *serviceIPFlag
	}

	// 4. 解析注册中心地址列表，优先级：新参数 > 旧参数 > 配置文件
	var registryAddrs []string
	if *registryAddrsFlag != "" {
		registryAddrs = strings.Split(*registryAddrsFlag, ",")
	} else if *registryAddrFlag != "" {
		registryAddrs = []string{*registryAddrFlag}
	} else if cfg.RegistryAddress != "" {
		registryAddrs = []string{cfg.RegistryAddress}
	}

	if len(registryAddrs) == 0 {
		logrus.Fatalf("No registry addresses provided.")
	}
	logrus.Infof("Registry addresses: %v", registryAddrs)

	// 5. 初始化日志
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("Client service starting with config: %+v", cfg)

	// 6. 初始化 HTTP 客户端
	httpClientConfig := httpclient.Config{
		Timeout:    cfg.HTTPClientTimeout,
		MaxRetries: cfg.HTTPClientMaxRetries,
		RetryDelay: cfg.HTTPClientRetryDelay,
	}
	httpClient := httpclient.NewClient(httpClientConfig)

	// 7. 服务注册
	clientIP := cfg.IPAddress
	if clientIP == "" {
		localIP, err := util.GetLocalIP()
		if err != nil {
			logrus.Fatalf("Failed to get local IP address for client service: %v", err)
		}
		clientIP = localIP
	}

	// 传入地址列表
	serviceID, err := client.RegisterService(
		registryAddrs,
		cfg.ServiceName,
		clientIP,
		cfg.Port,
	)
	if err != nil {
		logrus.Fatalf("Failed to register client service to registry: %v", err)
	}
	logrus.Infof("Client service registered with ID: %s", serviceID)

	// 8. 启动心跳
	// 传入地址列表
	heartbeatStopChan := client.StartHeartbeat(
		registryAddrs,
		serviceID,
		clientIP,
		cfg.Port,
		cfg.HeartbeatInterval,
	)

	// 9. 配置 Gin 路由
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceID)
		// 将地址列表传递给上下文，InfoHandler需要用它来做服务发现
		c.Set("registryAddrs", registryAddrs)
		c.Set("httpClient", httpClient)
		c.Next()
	})

	router.GET("/api/getInfo", client.InfoHandler)

	// 10. 启动 HTTP 服务器
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

	// 11. 优雅关闭
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

	// 12. 服务注销
	// 传入地址列表
	err = client.UnregisterService(
		registryAddrs,
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
