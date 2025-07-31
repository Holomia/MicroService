package register

import (
	"MicroService/pkg/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

// api/register
// Config 定义注册中心的配置
type Config struct {
	HeartbeatTTL  time.Duration
	CleanupPeriod time.Duration
}

// Register 注册中心核心结构
type Register struct {
	services      sync.Map           // 存储服务实例，键为 serviceId，值为 model.Service
	roundRobin    map[string]*uint64 // 服务名到轮询计数器的映射
	roundRobinMu  sync.RWMutex       // 保护 roundRobin 映射
	heartbeatTTL  time.Duration      // 心跳超时时间
	cleanupPeriod time.Duration      // 清理周期
}

// NewRegister 创建并初始化注册中心
// NewRegister 创建并初始化注册中心
func NewRegister(config Config) *Register {
	r := &Register{
		services:      sync.Map{},
		roundRobin:    make(map[string]*uint64),
		roundRobinMu:  sync.RWMutex{},
		heartbeatTTL:  config.HeartbeatTTL,
		cleanupPeriod: config.CleanupPeriod,
	}
	go r.startCleanup()
	return r
}

// StoreService 存储服务实例
func (r *Register) StoreService(service model.Service) {
	r.services.Store(service.ServiceId, service)
}

// LoadService 加载服务实例
func (r *Register) LoadService(serviceId string) (model.Service, bool) {
	if s, ok := r.services.Load(serviceId); ok {
		return s.(model.Service), true
	}
	return model.Service{}, false
}

// DeleteService 删除服务实例
func (r *Register) DeleteService(serviceId string) {
	r.services.Delete(serviceId)
}

// RegisterHandler 处理服务注册请求
func (r *Register) RegisterHandler(c *gin.Context) {
	var req model.RegisterServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// 转换为 Service 并验证
	service := req.ToService()
	if err := service.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: err.Error(),
		})
		return
	}

	// 设置初始心跳时间
	service.LastHeartbeat = time.Now()
	r.StoreService(service)

	// 返回成功响应
	c.JSON(http.StatusOK, model.RegisterServiceResponse{
		Message: "Service registered successfully",
		Service: service,
	})
}
