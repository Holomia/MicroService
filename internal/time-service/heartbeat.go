package timeservice

// 定时调用api/heartbeat
import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
)

// StartHeartbeat 定期向注册中心发送心跳
// registryAddr: 注册中心的地址，例如 "http://localhost:8180"
// serviceId: 本服务实例的唯一ID
// ipAddress: 本服务实例的IP地址
// port: 本服务实例的端口
// interval: 心跳间隔
func StartHeartbeat(registryAddrs []string, serviceId, ipAddress string, port int, interval time.Duration) chan struct{} {
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{})

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	heartbeatReq := model.HeartbeatRequest{
		ServiceId: serviceId,
		IpAddress: ipAddress,
		Port:      port,
	}

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for _, registryAddr := range registryAddrs {
					heartbeatURL := fmt.Sprintf("%s/api/heartbeat", registryAddr)
					var heartbeatResp model.HeartbeatResponse
					err := httpClient.Post(heartbeatURL, heartbeatReq, &heartbeatResp, clientConfig)
					if err != nil {
						logrus.Errorf("Failed to send heartbeat for service %s to %s: %v", serviceId, registryAddr, err)
					} else {
						logrus.Debugf("Heartbeat sent successfully for service: %s to %s", heartbeatResp.ServiceId, registryAddr)
					}
				}
			case <-stopChan:
				logrus.Infof("Heartbeat stopped for service: %s", serviceId)
				return
			}
		}
	}()
	return stopChan
}
