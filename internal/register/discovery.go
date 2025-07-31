package register

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"MicroService/pkg/model"
)

// GetAllServices 获取所有服务实例
func (r *Register) GetAllServices() []model.Service {
	var services []model.Service
	r.services.Range(func(_, value interface{}) bool {
		services = append(services, value.(model.Service))
		return true
	})
	return services
}

// GetServiceByName 获取指定服务名的健康实例（轮询负载均衡）
func (r *Register) GetServiceByName(name string) (model.Service, bool) {
	var healthy []model.Service
	now := time.Now()
	r.services.Range(func(_, value interface{}) bool {
		s := value.(model.Service)
		if s.ServiceName == name && now.Sub(s.LastHeartbeat) <= r.heartbeatTTL {
			healthy = append(healthy, s)
		}
		return true
	})

	if len(healthy) == 0 {
		return model.Service{}, false
	}

	// 使用读锁检查计数器是否存在
	r.roundRobinMu.RLock()
	counter, ok := r.roundRobin[name]
	r.roundRobinMu.RUnlock()

	if !ok {
		// 仅在必要时使用写锁初始化
		r.roundRobinMu.Lock()
		if _, exists := r.roundRobin[name]; !exists {
			r.roundRobin[name] = new(uint64)
		}
		counter = r.roundRobin[name]
		r.roundRobinMu.Unlock()
	}

	// 轮询选择实例
	index := atomic.AddUint64(counter, 1) % uint64(len(healthy))
	return healthy[index], true
}

// DiscoveryHandler 处理服务发现请求
func (r *Register) DiscoveryHandler(c *gin.Context) {
	name := c.Query("name")

	if name == "" {
		// 返回所有服务实例
		services := r.GetAllServices()
		c.JSON(http.StatusOK, model.DiscoveryListResponse{
			Services: services,
		})
		return
	}

	// 返回单个服务实例（轮询负载均衡）
	if service, ok := r.GetServiceByName(name); ok {
		c.JSON(http.StatusOK, model.DiscoveryResponse{
			ServiceName: service.ServiceName,
			ServiceId:   service.ServiceId,
			IpAddress:   service.IpAddress,
			Port:        service.Port,
		})
		return
	}

	// 服务未找到
	c.JSON(http.StatusNotFound, model.ErrorResponse{
		Code:  http.StatusNotFound,
		Error: "No healthy service instances found for " + name,
	})
}
