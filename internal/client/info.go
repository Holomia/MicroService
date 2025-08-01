package client

//实现api/getInfo 调用api/getDateTime？style
import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"time"

	//"MicroService/internal/client"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// InfoHandler 处理获取客户端信息请求
// 此函数将负责：
// 1. 从注册中心发现一个可用的 time-service 实例。
// 2. 调用 time-service 的 /api/getDateTime?style=full 接口。
// 3. 拼接返回结果并响应。
// 4. 处理 time-service 不可用的情况。
// InfoHandler handles the client's /api/getInfo request
func InfoHandler(c *gin.Context) {
	// 从 Gin 上下文获取客户端 ID
	clientID, exists := c.Get("serviceId")
	if !exists {
		errMsg := "Client Service ID not found in context."
		logrus.Error(errMsg)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}
	currentClientID := clientID.(string)

	// 从 Gin 上下文获取注册中心地址列表和 HTTP 客户端
	registryAddrs, registryAddrsExists := c.Get("registryAddrs")
	httpClient, httpClientExists := c.Get("httpClient")

	// 检查是否获取到注册中心地址列表和 HTTP 客户端
	if !registryAddrsExists || !httpClientExists {
		errMsg := "Missing registry address or HTTP client in Gin context."
		logrus.Error(errMsg)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}

	// 类型断言，确保获取到的是我们期望的类型
	registryAddrsSlice := registryAddrs.([]string)
	client := httpClient.(*httpclient.Client)

	// 遍历注册中心，尝试发现 time-service
	var timeServiceInstance *model.Service
	var discoveryErr error

	// 循环，直到找到一个可用的服务实例，或遍历完所有注册中心
	// 这种方法提供了简单的故障转移
	for _, addr := range registryAddrsSlice {
		discoveryURL := fmt.Sprintf("%s/api/discovery?name=time-service", addr)
		var discoveryResp model.DiscoveryResponse

		err := client.Get(discoveryURL, &discoveryResp, httpclient.DefaultConfig())
		if err != nil {
			logrus.Warnf("Failed to discover 'time-service' from registry %s: %v", addr, err)
			discoveryErr = err
			continue // 尝试下一个注册中心
		}

		if discoveryResp.ServiceId != "" {
			// 将发现的实例信息直接赋值给 timeServiceInstance
			timeServiceInstance = &model.Service{
				ServiceName: discoveryResp.ServiceName,
				ServiceId:   discoveryResp.ServiceId,
				IpAddress:   discoveryResp.IpAddress,
				Port:        discoveryResp.Port,
			}
			break
		}
	}

	// 如果遍历完所有注册中心都找不到服务，则返回错误
	if timeServiceInstance == nil {
		var errMsg string
		if discoveryErr != nil {
			errMsg = fmt.Sprintf("Time service unavailable. All registries failed to discover: %v", discoveryErr)
		} else {
			errMsg = "Time service unavailable. Could not find a healthy instance from any registry."
		}
		logrus.Error(errMsg)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}

	// 构造调用时间服务的 URL
	timeServiceURL := fmt.Sprintf("http://%s:%d/api/getDateTime?style=full", timeServiceInstance.IpAddress, timeServiceInstance.Port)

	// 调用时间服务获取 GMT 时间
	var timeServiceResp model.GetDateTimeResponse
	err := client.Get(timeServiceURL, &timeServiceResp, httpclient.DefaultConfig())
	if err != nil {
		errMsg := "Failed to call time-service."
		logrus.Errorf("%s: %v", errMsg, err)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}

	// 将 GMT 时间转换为北京时间
	gmtTime, err := time.Parse("2006-01-02 15:04:05", timeServiceResp.Result)
	if err != nil {
		errMsg := "Failed to parse GMT time from time-service."
		logrus.Errorf("%s: %v", errMsg, err)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}
	beijingLocation, _ := time.LoadLocation("Asia/Shanghai")
	beijingTime := gmtTime.In(beijingLocation)

	// 构建最终响应
	result := fmt.Sprintf("Hello Kingsoft Cloud Star Camp - %s - %s", currentClientID, beijingTime.Format("2006-01-02 15:04:05"))
	c.JSON(http.StatusOK, model.GetInfoResponse{
		Error:  nil,
		Result: &result,
	})
}
