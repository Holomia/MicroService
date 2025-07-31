package register

// api/heartbeat 更新LastHeartbeat
import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"MicroService/pkg/model"
)

// HeartbeatHandler 处理心跳请求
func (r *Register) HeartbeatHandler(c *gin.Context) {
	var req model.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// 验证请求
	if req.ServiceId == "" || req.IpAddress == "" || req.Port <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Missing required fields",
		})
		return
	}

	// 检查服务是否存在
	stored, ok := r.LoadService(req.ServiceId)
	if !ok {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:  http.StatusNotFound,
			Error: "Service not found",
		})
		return
	}

	// 验证 IP 和端口
	if stored.IpAddress != req.IpAddress || stored.Port != req.Port {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Service information does not match",
		})
		return
	}

	// 更新心跳时间
	stored.LastHeartbeat = time.Now()
	r.StoreService(stored)

	// 返回成功响应
	c.JSON(http.StatusOK, model.HeartbeatResponse{
		Message:   "Heartbeat received",
		ServiceId: req.ServiceId,
	})
}
