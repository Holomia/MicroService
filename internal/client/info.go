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
func InfoHandler(c *gin.Context) {
	// 获取客户端自身的 ServiceId（从 Gin 上下文获取，main.go 中设置）
	clientID, exists := c.Get("serviceId")
	if !exists {
		// 理论上不会发生，因为 main.go 中会设置。作为安全措施，提供一个默认值或直接返回错误。
		// 这里选择返回错误，因为客户端 ID 是预期存在的。
		errMsg := "Client Service ID not found in context."
		logrus.Error(errMsg)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg, // 传入指向错误信息的指针
			Result: nil,
		})
		return
	}
	currentClientID := clientID.(string) // 类型断言为 string

	// 从 Gin 上下文获取注册中心地址和 http 客户端
	registryAddr, registryAddrExists := c.Get("registryAddr")
	httpClient, httpClientExists := c.Get("httpClient")

	if !registryAddrExists || !httpClientExists {
		errMsg := "Missing registry address or HTTP client in Gin context."
		logrus.Errorf(errMsg)
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg, // 传入指向错误信息的指针
			Result: nil,
		})
		return
	}

	regAddr := registryAddr.(string)
	client := httpClient.(*httpclient.Client) // 类型断言

	// 1. 服务发现：从注册中心获取一个 time-service 实例
	timeService, err := DiscoverTimeService(regAddr, client)
	if err != nil {
		logrus.Errorf("Failed to discover time-service: %v", err)
		errMsg := fmt.Sprintf("Failed to get time information: %s", err.Error()) // 将错误信息转为字符串
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg, // 传入指向错误信息的指针
			Result: nil,
		})
		return
	}

	//添加调试信息，确认获取的IP和端口
	logrus.Infof("Discovered time-service instance: ID=%s, IP=%s, Port=%d",
		timeService.ServiceId, timeService.IpAddress, timeService.Port)

	// 2. 调用 time-service 的 API
	timeServiceURL := fmt.Sprintf("http://%s:%d/api/getDateTime?style=full", timeService.IpAddress, timeService.Port)
	var timeResp model.GetDateTimeResponse

	// 在这里添加打印信息，确认将要调用的完整 URL
	logrus.Infof("Attempting to call time-service at URL: %s", timeServiceURL)

	// 使用 httpclient 的默认配置进行调用，可以根据需要调整
	httpClientConfig := httpclient.DefaultConfig()
	err = client.Get(timeServiceURL, &timeResp, httpClientConfig)
	if err != nil {
		logrus.Errorf("Failed to call time-service %s (ID: %s) at %s:%d: %v",
			timeService.ServiceName, timeService.ServiceId, timeService.IpAddress, timeService.Port, err)
		errMsg := fmt.Sprintf("Time service unavailable or returned error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg, // 传入指向错误信息的指针
			Result: nil,
		})
		return
	}

	// --- 时区转换逻辑 ---
	// /api/getDateTime 返回 GMT 时间，/api/getInfo 要求返回北京时间
	const timeLayout = "2006-01-02 15:04:05" // time-service "full" 格式
	gmtTimeStr := timeResp.Result            // 获取到的 GMT 时间字符串

	// 解析 GMT 时间字符串
	// time.Parse 默认将时间字符串视为 UTC 时间（如果未指定时区信息）
	gmtTime, err := time.Parse(timeLayout, gmtTimeStr)
	if err != nil {
		logrus.Errorf("Failed to parse GMT time '%s' from time-service: %v", gmtTimeStr, err)
		errMsg := "Internal error parsing time from time-service."
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}

	// 加载中国标准时间 (CST) 时区
	// "Asia/Shanghai" 是一个常用的时区名，代表 CST (UTC+8)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Errorf("Failed to load 'Asia/Shanghai' timezone: %v", err)
		// 如果时区加载失败，无法进行转换，应返回错误
		errMsg := "Failed to convert time to Beijing time due to timezone loading error."
		c.JSON(http.StatusInternalServerError, model.GetInfoResponse{
			Error:  &errMsg,
			Result: nil,
		})
		return
	}

	// 将解析出的 GMT 时间转换为北京时间
	beijingTime := gmtTime.In(loc)

	// 格式化为所需的字符串形式
	formattedBeijingTime := beijingTime.Format(timeLayout)
	// --- 时区转换逻辑结束 ---

	// 3. 拼接结果
	// 根据项目要求，这里应拼接客户端自身的 Service ID
	resultString := fmt.Sprintf("Hello Kingsoft Cloud Star Camp - %s - %s", currentClientID, formattedBeijingTime)

	// 4. 响应客户端
	var successError *string = nil // 成功时，错误指针为 nil
	c.JSON(http.StatusOK, model.GetInfoResponse{
		Error:  successError, // 赋值为 nil
		Result: resultString,
	})
}
