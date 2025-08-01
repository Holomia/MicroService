package timeservice

//调用api/register
import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"MicroService/pkg/util"
	"fmt"
)

// RegisterService 向注册中心注册服务
// registryAddr: 注册中心的地址，例如 "http://localhost:8180"
// serviceName: 服务名称，例如 "time-service"
// ipAddress: 本服务实例的IP地址，如果为空则内部尝试自动检测
// port: 本服务实例运行的端口
func RegisterService(registryAddrs []string, serviceName, ipAddress string, port int) (string, error) {
	finalIPAddr := ipAddress
	if finalIPAddr == "" {
		// 如果未手动指定，则尝试自动获取
		var err error
		finalIPAddr, err = util.GetLocalIP()
		if err != nil {
			return "", fmt.Errorf("failed to get local IP address and no explicit IP was provided: %v", err)
		}
	}

	// 生成唯一的 ServiceId
	serviceId := util.GenerateUUID()

	registerReq := model.RegisterServiceRequest{
		ServiceName: serviceName,
		ServiceId:   serviceId,
		IpAddress:   finalIPAddr,
		Port:        port,
	}

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	// 遍历所有注册中心地址进行注册
	for _, registryAddr := range registryAddrs {
		registerURL := fmt.Sprintf("%s/api/register", registryAddr)
		var registerResp model.RegisterServiceResponse
		err := httpClient.Post(registerURL, registerReq, &registerResp, clientConfig)
		if err != nil {
			// 如果注册失败，这里可以根据需要决定是继续尝试其他注册中心还是直接返回错误
			return "", fmt.Errorf("failed to register service %s-%s at %s:%d to registry %s: %v", serviceName, serviceId, finalIPAddr, port, registryAddr, err)
		}
		fmt.Printf("Service registered successfully to %s: %s\n", registryAddr, registerResp.Message)
	}

	return serviceId, nil
}
