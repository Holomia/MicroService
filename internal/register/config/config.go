package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Config 定义注册中心的配置
type Config struct {
	Port          int
	HeartbeatTTL  time.Duration
	CleanupPeriod time.Duration
	SyncAddresses []string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() Config {
	config := Config{
		Port:          8180,
		HeartbeatTTL:  180 * time.Second,
		CleanupPeriod: 60 * time.Second,
		SyncAddresses: []string{},
	}

	if portStr := os.Getenv("REGISTRY_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			config.Port = port
		} else {
			logrus.Warnf("Invalid REGISTRY_PORT: %s, using default: %d", portStr, config.Port)
		}
	}

	if ttlStr := os.Getenv("HEARTBEAT_TTL_SECONDS"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 {
			config.HeartbeatTTL = time.Duration(ttl) * time.Second
		} else {
			logrus.Warnf("Invalid HEARTBEAT_TTL_SECONDS: %s, using default: %v", ttlStr, config.HeartbeatTTL)
		}
	}
	if periodStr := os.Getenv("CLEANUP_PERIOD_SECONDS"); periodStr != "" {
		if period, err := strconv.Atoi(periodStr); err == nil && period > 0 {
			config.CleanupPeriod = time.Duration(period) * time.Second
		} else {
			logrus.Warnf("Invalid CLEANUP_PERIOD_SECONDS: %s, using default: %v", periodStr, config.CleanupPeriod)
		}
	}
	if syncStr := os.Getenv("SYNC_ADDRESSES"); syncStr != "" {
		config.SyncAddresses = strings.Split(syncStr, ",")
	}
	return config
}
