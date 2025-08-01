package main

import (
	"flag" // 导入 flag 包
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"MicroService/internal/register"
	config2 "MicroService/internal/register/config" // 你的配置包
)

func main() {
	// 1. 定义命令行参数
	// 定义一个整型变量来存储命令行参数 "port" 的值，默认值为 8180
	portFlag := flag.Int("port", 0, "The port for the registry service to listen on. Overrides environment variable.")
	flag.Parse() // 解析命令行参数

	// 2. 加载配置（首先从环境变量和默认值）
	// LoadConfig 函数会从环境变量加载配置，如果环境变量不存在，则使用默认值
	config := config2.LoadConfig()

	// 3. 使用命令行参数覆盖配置
	// 如果命令行参数 --port 被指定（即不为 0），则使用它的值
	if *portFlag != 0 {
		config.Port = *portFlag
	}

	// 4. 初始化注册中心
	regConfig := register.Config{
		HeartbeatTTL:  config.HeartbeatTTL,
		CleanupPeriod: config.CleanupPeriod,
	}
	reg := register.NewRegister(regConfig)

	// 5. 初始化 Gin 路由
	r := gin.Default()

	// 6. 注册 API 端点
	r.POST("/api/register", reg.RegisterHandler)
	r.POST("/api/unregister", reg.UnregisterHandler)
	r.POST("/api/heartbeat", reg.HeartbeatHandler)
	r.GET("/api/discovery", reg.DiscoveryHandler)

	// 7. 启动服务
	addr := fmt.Sprintf(":%d", config.Port)
	logrus.Infof("Starting registry service on %s", addr)
	if err := r.Run(addr); err != nil {
		logrus.Fatalf("Failed to start registry service: %v", err)
	}
}
