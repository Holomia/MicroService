package config

import (
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type TimeServiceConfig struct {
	Port              int
	RegistryAddr      string
	ServiceHostIP     string
	HeartbeatInterval time.Duration
}

func LoadTimeServiceConfig() TimeServiceConfig {
	config := DefaultTimeServiceConfig()

	if portStr := os.Getenv("SERVICE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			config.Port = port
		} else {
			logrus.Warnf("Invalid SERVICE_PORT: %s, using default: %d", portStr, config.Port)
		}
	}

	if addr := os.Getenv("REGISTRY_ADDR"); addr != "" {
		config.RegistryAddr = addr
	}

	if ip := os.Getenv("SERVICE_HOST_IP"); ip != "" {
		config.ServiceHostIP = ip
	}

	if intervalStr := os.Getenv("HEARTBEAT_INTERVAL_SECONDS"); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			config.HeartbeatInterval = time.Duration(interval) * time.Second
		} else {
			logrus.Warnf("Invalid HEARTBEAT_INTERVAL_SECONDS: %s, using default: %v", intervalStr, config.HeartbeatInterval)
		}
	}

	return config
}

func DefaultTimeServiceConfig() TimeServiceConfig {
	return TimeServiceConfig{
		Port:              8280,
		RegistryAddr:      "http://localhost:8180",
		ServiceHostIP:     "",
		HeartbeatInterval: 60 * time.Second,
	}
}
