package client

//定时调用api/heartbeat
import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
)

// StartHeartbeat 定期向注册中心发送客户端心跳
func StartHeartbeat(registryAddr, serviceId, ipAddress string, port int, interval time.Duration) chan struct{} {
	ticker := time.NewTicker(interval)
	stopChan := make(chan struct{})

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
					logrus.Errorf("Failed to send heartbeat for client service %s: %v", serviceId, err)
				} else {
					logrus.Debugf("Heartbeat sent successfully for client service: %s", heartbeatResp.ServiceId)
				}
			case <-stopChan:
				logrus.Infof("Heartbeat stopped for client service: %s", serviceId)
				return
			}
		}
	}()
	return stopChan
}
