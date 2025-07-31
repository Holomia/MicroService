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
func StartHeartbeat(registryAddr, serviceId, ipAddress string, port int, interval time.Duration) chan struct{} {
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{}) // 用于停止心跳的通道

	clientConfig := httpclient.DefaultConfig()
	httpClient := httpclient.NewClient(clientConfig)

	heartbeatURL := fmt.Sprintf("%s/api/heartbeat", registryAddr)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				heartbeatReq := model.HeartbeatRequest{
					ServiceId: serviceId,
					IpAddress: ipAddress,
					Port:      port,
				}
				var heartbeatResp model.HeartbeatResponse

				err := httpClient.Post(heartbeatURL, heartbeatReq, &heartbeatResp, clientConfig)
				if err != nil {
					logrus.Errorf("Failed to send heartbeat for service %s: %v", serviceId, err)
				} else {
					logrus.Debugf("Heartbeat sent successfully for service: %s", heartbeatResp.ServiceId)
				}
			case <-stopChan:
				logrus.Infof("Heartbeat stopped for service: %s", serviceId)
				return
			}
		}
	}()
	return stopChan
}
