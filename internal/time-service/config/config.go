package config

import "time"

type TimeServiceConfig struct {
	Port              int           // 服务运行端口
	RegistryAddr      string        // 注册中心地址
	ServiceHostIP     string        // 可选：服务注册时使用的IP地址，如果为空则自动检测
	HeartbeatInterval time.Duration // 心跳间隔
}

func DefaultTimeServiceConfig() TimeServiceConfig {
	return TimeServiceConfig{
		Port:              8280,                    // 默认端口
		RegistryAddr:      "http://localhost:8180", // 默认注册中心地址
		ServiceHostIP:     "",                      // 默认不指定，自动检测
		HeartbeatInterval: 60 * time.Second,        // 默认心跳间隔
	}
}
