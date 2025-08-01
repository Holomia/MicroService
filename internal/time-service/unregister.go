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
// registryAddrs: 注册中心的地址列表
func UnregisterService(registryAddrs []string, serviceName, serviceId, ipAddress string, port int) error {
	unregisterReq := model.RegisterServiceRequest{ // 复用注册请求结构
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
			logrus.Errorf("Failed to unregister service %s-%s from registry %s: %v", serviceName, serviceId, registryAddr, err)
			continue
		}
		logrus.Infof("Service unregistered successfully from %s: %s", registryAddr, unregisterResp.Message)
	}

	return nil // 返回 nil 表示注销流程已完成
}
