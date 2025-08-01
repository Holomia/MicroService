package client

//调用api/unregister
import (
	"fmt"

	"github.com/sirupsen/logrus"

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
)

func UnregisterService(registryAddrs []string, serviceName, serviceId, ipAddress string, port int) error {
	unregisterReq := model.RegisterServiceRequest{
		ServiceName: serviceName,
		ServiceId:   serviceId,
		IpAddress:   ipAddress,
		Port:        port,
	}

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	// 遍历所有注册中心地址，向每个地址发送注销请求
	for _, registryAddr := range registryAddrs {
		unregisterURL := fmt.Sprintf("%s/api/unregister", registryAddr)
		var unregisterResp model.RegisterServiceResponse

		err := httpClient.Post(unregisterURL, unregisterReq, &unregisterResp, clientConfig)
		if err != nil {
			// 注意：这里可以选择返回第一个遇到的错误，或者记录所有错误并继续
			logrus.Errorf("Failed to unregister client service %s-%s from registry %s: %v", serviceName, serviceId, registryAddr, err)
			// 这里为了简单，我们选择记录错误并继续，确保尝试注销所有节点
			continue
		}
		logrus.Infof("Client Service unregistered successfully from %s: %s", registryAddr, unregisterResp.Message)
	}

	return nil // 返回 nil 表示注销流程已完成
}
