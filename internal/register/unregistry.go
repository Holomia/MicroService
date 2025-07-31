package register

// api/unregister实现
import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MicroService/pkg/model"
)

// UnregisterHandler 处理服务注销请求
func (r *Register) UnregisterHandler(c *gin.Context) {
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

	// 检查服务是否存在且信息匹配
	stored, ok := r.LoadService(service.ServiceId)
	if !ok {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:  http.StatusNotFound,
			Error: "Service not found",
		})
		return
	}
	if stored.ServiceName != service.ServiceName ||
		stored.IpAddress != service.IpAddress ||
		stored.Port != service.Port {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Service information does not match",
		})
		return
	}

	// 删除服务
	r.DeleteService(service.ServiceId)

	// 返回成功响应
	c.JSON(http.StatusOK, model.RegisterServiceResponse{
		Message: "Service unregistered successfully",
		Service: service,
	})
}
