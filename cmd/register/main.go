package main

import (
	"context"
	"flag" // 导入 flag 包
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"MicroService/internal/register"
	config2 "MicroService/internal/register/config" // 你的配置包
)

type RegistryConfig struct {
	Port  int
	Peers []string
}

func main() {
	// 1. 定义命令行参数
	portFlag := flag.Int("port", 0, "The port for the registry service to listen on. Overrides environment variable.")
	peersFlag := flag.String("peers", "", "Comma-separated list of peer registry addresses (e.g., 'http://127.0.0.1:8181,http://127.0.0.1:8182')")
	flag.Parse()

	// 2. 加载配置
	config := config2.LoadConfig()

	// 3. 使用命令行参数覆盖配置
	if *portFlag != 0 {
		config.Port = *portFlag
	}

	// 4. 解析对等节点地址
	var peers []string
	if *peersFlag != "" {
		peers = strings.Split(*peersFlag, ",")
	}

	// 5. 初始化注册中心
	regConfig := register.Config{
		HeartbeatTTL:  config.HeartbeatTTL,
		CleanupPeriod: config.CleanupPeriod,
	}
	// 将解析出的对等节点地址传递给NewRegister
	reg := register.NewRegister(regConfig, peers)

	// 6. 初始化 Gin 路由
	r := gin.Default()

	// 7. 注册 API 端点
	r.POST("/api/register", reg.RegisterHandler)
	r.POST("/api/unregister", reg.UnregisterHandler)
	r.POST("/api/heartbeat", reg.HeartbeatHandler)
	r.GET("/api/discovery", reg.DiscoveryHandler)

	// 新增：注册内部同步端点
	r.POST("/api/internal/sync", reg.SyncHandler)

	// 8. 启动服务
	addr := fmt.Sprintf(":%d", config.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logrus.Infof("Starting registry service on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start registry service: %v", err)
		}
	}()

	// 9. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down registry service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Registry service forced to shutdown: %v", err)
	}

	logrus.Info("Registry service stopped gracefully.")
}
