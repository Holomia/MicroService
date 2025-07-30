package model

import (
	"errors"
	"time"
)

type Service struct {
	ServiceName   string    `json:"serviceName"` // 服务名称，例如 "time-service"
	ServiceId     string    `json:"serviceId"`   // 服务实例唯一标识，建议使用 UUID
	IpAddress     string    `json:"ipAddress"`   // 服务实例的 IP 地址
	Port          int       `json:"port"`        // 服务实例的端口号
	LastHeartbeat time.Time `json:"-"`           // 最后一次心跳时间，仅用于内部管理
}

func (s *Service) Validate() error {
	if s.ServiceName == "" {
		return errors.New("serviceName is required")
	}
	if s.ServiceId == "" {
		return errors.New("serviceId is required")
	}
	if s.IpAddress == "" {
		return errors.New("ipAddress is required")
	}
	if s.Port <= 0 {
		return errors.New("port must be greater than 0")
	}
	return nil
}
func (r *RegisterServiceRequest) ToService() Service {
	return Service{
		ServiceName: r.ServiceName,
		ServiceId:   r.ServiceId,
		IpAddress:   r.IpAddress,
		Port:        r.Port,
	}
}

// 通用的错误响应
type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

// 注册服务请求
type RegisterServiceRequest struct {
	ServiceName string `json:"serviceName" binding:"required"`
	ServiceId   string `json:"serviceId" binding:"required"`
	IpAddress   string `json:"ipAddress" binding:"required"`
	Port        int    `json:"port" binding:"required"`
}

// 注册服务响应
type RegisterServiceResponse struct {
	Message string  `json:"message"`
	Service Service `json:"service"`
}

// 心跳请求
type HeartbeatRequest struct {
	ServiceId string `json:"serviceId" binding:"required"`
	IpAddress string `json:"ipAddress" binding:"required"`
	Port      int    `json:"port" binding:"required"`
}

// 心跳响应
type HeartbeatResponse struct {
	Message   string `json:"message"`
	ServiceId string `json:"serviceId"`
}

// 服务发现响应（单个实例）
type DiscoveryResponse struct {
	ServiceName string `json:"serviceName"`
	ServiceId   string `json:"serviceId"`
	IpAddress   string `json:"ipAddress"`
	Port        int    `json:"port"`
}

// 服务发现响应（所有实例）
type DiscoveryListResponse struct {
	Services []Service `json:"services"`
}

// 时间服务响应
type GetDateTimeResponse struct {
	Result    string `json:"result"`
	ServiceId string `json:"serviceId"`
}

// 客户端信息响应
type GetInfoResponse struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
