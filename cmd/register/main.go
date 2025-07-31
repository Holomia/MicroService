package main

import (
	config2 "MicroService/internal/register/config"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"MicroService/internal/register"
)

func main() {
	// 加载配置
	config := config2.LoadConfig()

	// 初始化注册中心
	regConfig := register.Config{
		HeartbeatTTL:  config.HeartbeatTTL,
		CleanupPeriod: config.CleanupPeriod,
	}
	reg := register.NewRegister(regConfig)

	// 初始化 Gin 路由
	r := gin.Default()

	// 注册 API 端点
	r.POST("/api/register", reg.RegisterHandler)
	r.POST("/api/unregister", reg.UnregisterHandler)
	r.POST("/api/heartbeat", reg.HeartbeatHandler)
	r.GET("/api/discovery", reg.DiscoveryHandler)

	// 启动服务
	addr := fmt.Sprintf(":%d", config.Port)
	logrus.Infof("Starting registry service on %s", addr)
	if err := r.Run(addr); err != nil {
		logrus.Fatalf("Failed to start registry service: %v", err)
	}
}
