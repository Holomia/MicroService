package timeservice

//退出信号时调用api/unregister
import (
	"fmt"

	"github.com/sirupsen/logrus"

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
)

// UnregisterService 向注册中心注销服务
// registryAddr: 注册中心的地址
// serviceName: 服务名称
// serviceId: 本服务实例的唯一ID
// ipAddress: 本服务实例的IP地址
// port: 本服务实例的端口
func UnregisterService(registryAddr, serviceName, serviceId, ipAddress string, port int) error {
	unregisterReq := model.RegisterServiceRequest{ // 复用注册请求结构
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
		return fmt.Errorf("failed to unregister service %s-%s: %v", serviceName, serviceId, err)
	}

	logrus.Infof("Service unregistered successfully: %s", unregisterResp.Message)
	return nil
}
