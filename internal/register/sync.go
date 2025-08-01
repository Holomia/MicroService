package register

import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SyncRequest 用于接收增量同步请求的结构体
type SyncRequest struct {
	Service model.Service `json:"service"`
	Action  string        `json:"action"` // "register" or "unregister"
}

// SyncHandler 处理来自其他注册中心的同步请求
func (r *Register) SyncHandler(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Invalid sync request body: " + err.Error(),
		})
		return
	}

	if req.Action == "register" {
		// 更新本地服务列表
		req.Service.LastHeartbeat = time.Now() // 收到同步时，更新心跳时间
		r.StoreService(req.Service)
		logrus.Infof("Synchronized new service registration: %s-%s", req.Service.ServiceName, req.Service.ServiceId)
	} else if req.Action == "unregister" {
		// 从本地服务列表移除
		r.DeleteService(req.Service.ServiceId)
		logrus.Infof("Synchronized service unregistration: %s-%s", req.Service.ServiceName, req.Service.ServiceId)
	} else {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Invalid sync action: " + req.Action,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Sync successful"})
}

// syncToPeers 异步地将变更推送到其他对等节点
func (r *Register) syncToPeers(service model.Service, action string) {
	// 使用 goroutine 异步发送，不阻塞主请求
	go func() {
		client := httpclient.NewClient(httpclient.DefaultConfig())
		syncReq := SyncRequest{
			Service: service,
			Action:  action,
		}

		for _, peer := range r.Peers {
			url := peer + "/api/internal/sync"
			var resp interface{} // 我们不需要解析响应
			err := client.Post(url, syncReq, &resp, httpclient.DefaultConfig())
			if err != nil {
				logrus.Errorf("Failed to sync %s action for service %s-%s to peer %s: %v",
					action, service.ServiceName, service.ServiceId, peer, err)
			} else {
				logrus.Debugf("Successfully synced %s for service %s to peer %s", action, service.ServiceId, peer)
			}
		}
	}()
}
