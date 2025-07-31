package client

//调用api/register
import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"MicroService/pkg/util"
	"fmt"
)

// RegisterService 向注册中心注册客户端服务
// registryAddr: 注册中心的地址
// serviceName: 服务名称，例如 "client"
// ipAddress: 本客户端实例的IP地址
// port: 本客户端实例运行的端口
func RegisterService(registryAddr, serviceName, ipAddress string, port int) (string, error) {
	finalIPAddr := ipAddress
	if finalIPAddr == "" {
		var err error
		finalIPAddr, err = util.GetLocalIP()
		if err != nil {
			return "", fmt.Errorf("failed to get local IP address and no explicit IP was provided for client: %v", err)
		}
	}

	serviceId := util.GenerateUUID() // 为客户端生成唯一的 ServiceId

	registerReq := model.RegisterServiceRequest{
		ServiceName: serviceName,
		ServiceId:   serviceId,
		IpAddress:   finalIPAddr,
		Port:        port,
	}

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	registerURL := fmt.Sprintf("%s/api/register", registryAddr)
	var registerResp model.RegisterServiceResponse

	err := httpClient.Post(registerURL, registerReq, &registerResp, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to register client service %s-%s at %s:%d: %v", serviceName, serviceId, finalIPAddr, port, err)
	}

	fmt.Printf("Client Service registered successfully: %s\n", registerResp.Message)
	return serviceId, nil
}

//// getLocalIP 获取本机非环回 IPv4 地址 (作为回退/自动检测使用)
//func getLocalIP() (string, error) {
//	addrs, err := net.InterfaceAddrs()
//	if err != nil {
//		return "", err
//	}
//	for _, address := range addrs {
//		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
//			if ipnet.IP.To4() != nil {
//				return ipnet.IP.String(), nil
//			}
//		}
//	}
//	return "", fmt.Errorf("no non-loopback IPv4 address found")
//}
