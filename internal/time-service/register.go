package timeservice

//调用api/register
import (
	"fmt"
	"net" // 仍然保留 getLocalIP 以备自动检测回退

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"MicroService/pkg/util"
)

// RegisterService 向注册中心注册服务
// registryAddr: 注册中心的地址，例如 "http://localhost:8180"
// serviceName: 服务名称，例如 "time-service"
// ipAddress: 本服务实例的IP地址，如果为空则内部尝试自动检测
// port: 本服务实例运行的端口
func RegisterService(registryAddr, serviceName, ipAddress string, port int) (string, error) {
	finalIPAddr := ipAddress
	if finalIPAddr == "" {
		// 如果未手动指定，则尝试自动获取
		var err error
		finalIPAddr, err = GetLocalIP()
		if err != nil {
			return "", fmt.Errorf("failed to get local IP address and no explicit IP was provided: %v", err)
		}
	}

	// 生成唯一的 ServiceId
	serviceId := util.GenerateUUID()

	registerReq := model.RegisterServiceRequest{
		ServiceName: serviceName,
		ServiceId:   serviceId,
		IpAddress:   finalIPAddr, // 使用最终确定的IP地址
		Port:        port,
	}

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	registerURL := fmt.Sprintf("%s/api/register", registryAddr)
	var registerResp model.RegisterServiceResponse

	err := httpClient.Post(registerURL, registerReq, &registerResp, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to register service %s-%s at %s:%d: %v", serviceName, serviceId, finalIPAddr, port, err)
	}

	fmt.Printf("Service registered successfully: %s\n", registerResp.Message)
	return serviceId, nil
}

// getLocalIP 获取本机非环回 IPv4 地址 (作为回退/自动检测使用)
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no non-loopback IPv4 address found")
}
