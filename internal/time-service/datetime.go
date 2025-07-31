package timeservice

import (
	"MicroService/pkg/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// api/getDateTime

// DateTimeHandler 处理获取日期时间请求
func DateTimeHandler(c *gin.Context) {
	style := c.Query("style")
	if style == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "style parameter is required",
		})
		return
	}

	// 获取当前 UTC 时间
	now := time.Now().UTC()
	var result string
	switch style {
	case "full":
		result = now.Format("2006-01-02 15:04:05")
	case "date":
		result = now.Format("2006-01-02")
	case "time":
		result = now.Format("15:04:05")
	case "unix":
		result = fmt.Sprintf("%d", now.UnixMilli()) // 返回 13 位毫秒时间戳
	default:
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "Invalid style parameter. Accepted values: full, date, time, unix",
		})
		return
	}

	// 从 Gin 上下文中获取服务ID，这需要在 main.go 中设置
	// 例如：c.Set("serviceId", yourServiceId)
	serviceId, exists := c.Get("serviceId")
	if !exists {
		// 如果 serviceId 不存在，给一个默认值或返回错误，取决于你的设计
		serviceId = "unknown-service-id"
	}

	c.JSON(http.StatusOK, model.GetDateTimeResponse{
		Result:    result,
		ServiceId: serviceId.(string), // 类型断言为 string
	})
}
