package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Config 包含了客户端服务的所有配置
type Config struct {
	ServiceName          string
	Port                 int
	IPAddress            string // 可以为空，表示自动检测
	HeartbeatInterval    time.Duration
	RegistryAddress      string
	HTTPClientTimeout    time.Duration
	HTTPClientMaxRetries int
	HTTPClientRetryDelay time.Duration
	Debug                bool
}

// LoadConfig 从环境变量加载客户端配置
func LoadConfig() Config {
	config := Config{
		ServiceName:          "client",                // 默认服务名称
		Port:                 8300,                    // 默认端口
		IPAddress:            "",                      // 默认不指定IP，自动检测
		HeartbeatInterval:    60 * time.Second,        // 默认心跳间隔
		RegistryAddress:      "http://localhost:8180", // 默认注册中心地址
		HTTPClientTimeout:    10 * time.Second,        // 默认 HTTP 客户端超时
		HTTPClientMaxRetries: 3,                       // 默认 HTTP 客户端重试次数
		HTTPClientRetryDelay: 1 * time.Second,         // 默认 HTTP 客户端重试间隔
		Debug:                false,                   // 默认关闭调试模式
	}

	// 加载 SERVICE_NAME
	if name := os.Getenv("CLIENT_SERVICE_NAME"); name != "" {
		config.ServiceName = name
	}

	// 加载 CLIENT_PORT
	if portStr := os.Getenv("CLIENT_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			config.Port = port
		} else {
			logrus.Warnf("Invalid CLIENT_PORT: %s, using default: %d", portStr, config.Port)
		}
	}

	// 加载 CLIENT_IP_ADDRESS
	if ipAddr := os.Getenv("CLIENT_IP_ADDRESS"); ipAddr != "" {
		config.IPAddress = ipAddr
	}

	// 加载 HEARTBEAT_INTERVAL_SECONDS
	if intervalStr := os.Getenv("HEARTBEAT_INTERVAL_SECONDS"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			config.HeartbeatInterval = time.Duration(interval) * time.Second
		} else {
			logrus.Warnf("Invalid HEARTBEAT_INTERVAL_SECONDS: %s, using default: %v", intervalStr, config.HeartbeatInterval)
		}
	}

	// 加载 REGISTRY_ADDRESS
	if regAddr := os.Getenv("REGISTRY_ADDRESS"); regAddr != "" {
		config.RegistryAddress = regAddr
	}

	// 加载 HTTP_CLIENT_TIMEOUT_SECONDS
	if timeoutStr := os.Getenv("HTTP_CLIENT_TIMEOUT_SECONDS"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			config.HTTPClientTimeout = time.Duration(timeout) * time.Second
		} else {
			logrus.Warnf("Invalid HTTP_CLIENT_TIMEOUT_SECONDS: %s, using default: %v", timeoutStr, config.HTTPClientTimeout)
		}
	}

	// 加载 HTTP_CLIENT_MAX_RETRIES
	if retriesStr := os.Getenv("HTTP_CLIENT_MAX_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil && retries >= 0 { // 重试次数可以为0
			config.HTTPClientMaxRetries = retries
		} else {
			logrus.Warnf("Invalid HTTP_CLIENT_MAX_RETRIES: %s, using default: %d", retriesStr, config.HTTPClientMaxRetries)
		}
	}

	// 加载 HTTP_CLIENT_RETRY_DELAY_SECONDS
	if delayStr := os.Getenv("HTTP_CLIENT_RETRY_DELAY_SECONDS"); delayStr != "" {
		if delay, err := strconv.Atoi(delayStr); err == nil && delay >= 0 {
			config.HTTPClientRetryDelay = time.Duration(delay) * time.Second
		} else {
			logrus.Warnf("Invalid HTTP_CLIENT_RETRY_DELAY_SECONDS: %s, using default: %v", delayStr, config.HTTPClientRetryDelay)
		}
	}

	// 加载 DEBUG
	if debugStr := os.Getenv("CLIENT_DEBUG"); debugStr != "" {
		config.Debug = strings.ToLower(debugStr) == "true"
	}

	return config
}
