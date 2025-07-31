package register

import (
	"MicroService/pkg/model"
	"github.com/sirupsen/logrus"
	"time"
)

//实现心跳超时清理机制

// startCleanup 启动定时清理任务
func (r *Register) startCleanup() {
	ticker := time.NewTicker(r.cleanupPeriod)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanupExpiredServices()
	}
}

// cleanupExpiredServices 清理心跳超时的服务实例
func (r *Register) cleanupExpiredServices() {
	now := time.Now()
	var expired []string

	r.services.Range(func(key, value interface{}) bool {
		service := value.(model.Service)
		if now.Sub(service.LastHeartbeat) > r.heartbeatTTL {
			expired = append(expired, key.(string))
		}
		return true
	})

	for _, serviceId := range expired {
		r.services.Delete(serviceId)
		logrus.Infof("Removed expired service: %s", serviceId)
	}
}
