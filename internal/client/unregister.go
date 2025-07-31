package client

//调用api/unregister
import (
	"fmt"

	"github.com/sirupsen/logrus"

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
)

// UnregisterService 向注册中心注销客户端服务
func UnregisterService(registryAddr, serviceName, serviceId, ipAddress string, port int) error {
	unregisterReq := model.RegisterServiceRequest{
		ServiceName: serviceName,
		ServiceId:   serviceId,
		IpAddress:   ipAddress,
		Port:        port,
	}

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	unregisterURL := fmt.Sprintf("%s/api/unregister", registryAddr)
	var unregisterResp model.RegisterServiceResponse

	err := httpClient.Post(unregisterURL, unregisterReq, &unregisterResp, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to unregister client service %s-%s: %v", serviceName, serviceId, err)
	}

	logrus.Infof("Client Service unregistered successfully: %s", unregisterResp.Message)
	return nil
}
