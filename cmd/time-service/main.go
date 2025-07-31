package main // 这是正确的，表示这是一个可执行程序

import (
	"MicroService/pkg/util"
	"fmt"      // 导入 fmt 包以使用 fmt.Sprintf
	"net/http" // 导入 net/http 包，以便引用 http.ErrServerClosed

	"MicroService/internal/time-service"        // 导入 time-service 内部逻辑包，路径正确
	"MicroService/internal/time-service/config" // 导入配置包，路径正确
	"os"
	"os/signal"
	"syscall"
	//"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus" // 使用 logrus 进行日志输出
)

func main() {
	// 初始化日志
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel) // 根据需要设置日志级别

	// 加载配置
	cfg := config.DefaultTimeServiceConfig()
	// TODO: 可以从命令行参数、环境变量或配置文件加载覆盖默认配置
	// 例如：
	// if err := env.Parse(&cfg); err != nil {
	//     logrus.Fatalf("Failed to load config from env: %v", err)
	// }

	logrus.Infof("Starting Time-Service on port %d...", cfg.Port)

	// 获取服务 IP 地址
	var currentIPAddress string
	if cfg.ServiceHostIP != "" {
		currentIPAddress = cfg.ServiceHostIP
		logrus.Infof("Using configured IP address: %s", currentIPAddress)
	} else {
		var err error
		// 注意：getLocalIP() 在 internal/time-service 包中，需要正确引用
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

	// 使用中间件将 serviceId 注入到 Gin 上下文
	router.Use(func(c *gin.Context) {
		c.Set("serviceId", serviceId)
		c.Next()
	})

	// 注册时间服务 API
	router.GET("/api/getDateTime", timeservice.DateTimeHandler)

	// 启动 HTTP 服务器
	go func() {
		if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil && err != http.ErrServerClosed {
			// 如果不是服务器正常关闭的错误，则记录为致命错误
			logrus.Fatalf("Gin server error: %v", err)
		}
	}()

	// 监听操作系统信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 监听 Ctrl+C (SIGINT) 和 kill (SIGTERM) 命令
	<-quit                                               // 阻塞主 Goroutine，直到接收到信号

	logrus.Info("Shutting down Time-Service...")

	// 停止心跳 Goroutine
	// 关闭通道会向所有监听该通道的 Goroutine 发送一个零值，从而解除它们的阻塞
	close(stopHeartbeatChan)

	// 发送注销请求
	err = timeservice.UnregisterService(cfg.RegistryAddr, serviceName, serviceId, currentIPAddress, cfg.Port)
	if err != nil {
		logrus.Errorf("Failed to unregister service during shutdown: %v", err)
	}

	logrus.Info("Time-Service stopped gracefully.")
}
