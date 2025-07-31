package client

//调用/api/discovery？name 获取时间服务实例
import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"fmt"
)

// DiscoverTimeService 从注册中心发现一个 time-service 实例
// registryAddr: 注册中心的地址，例如 "http://localhost:8180"
// httpClient: 用于发送 HTTP 请求的客户端实例
func DiscoverTimeService(registryAddr string, httpClient *httpclient.Client) (model.Service, error) {
	discoveryURL := fmt.Sprintf("%s/api/discovery?name=time-service", registryAddr)

	var discoveryResp model.DiscoveryResponse // 预期返回单个服务实例 (负载均衡)

	clientConfig := httpclient.DefaultConfig() // 使用默认 HTTP 客户端配置进行服务发现
	err := httpClient.Get(discoveryURL, &discoveryResp, clientConfig)
	if err != nil {
		return model.Service{}, fmt.Errorf("failed to discover 'time-service': %v", err)
	}

	// 验证发现到的服务信息
	if discoveryResp.ServiceId == "" || discoveryResp.IpAddress == "" || discoveryResp.Port <= 0 {
		return model.Service{}, fmt.Errorf("discovered 'time-service' instance has invalid information: %+v", discoveryResp)
	}

	return model.Service{
		ServiceName: discoveryResp.ServiceName,
		ServiceId:   discoveryResp.ServiceId,
		IpAddress:   discoveryResp.IpAddress,
		Port:        discoveryResp.Port,
	}, nil
}
