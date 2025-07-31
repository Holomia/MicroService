package register

//加分项预留
import (
	"MicroService/pkg/httpclient"
	"MicroService/pkg/model"
	"time"

	"github.com/sirupsen/logrus"
)

// SyncConfig 同步配置
type SyncConfig struct {
	Peers        []string      // 其他注册中心的地址列表
	SyncInterval time.Duration // 同步间隔
}

// StartSync 启动与其他注册中心的数据同步
func (r *Register) StartSync(config SyncConfig) {
	if len(config.Peers) == 0 {
		logrus.Warn("No peers configured for synchronization")
		return
	}

	ticker := time.NewTicker(config.SyncInterval)
	defer ticker.Stop()

	client := httpclient.NewClient(httpclient.DefaultConfig())

	for range ticker.C {
		r.syncWithPeers(client, config.Peers)
	}
}

// syncWithPeers 与其他注册中心同步数据
func (r *Register) syncWithPeers(client *httpclient.Client, peers []string) {
	services := r.GetAllServices()
	for _, peer := range peers {
		url := peer + "/api/sync" // 假设同步端点为 /api/sync
		var resp model.DiscoveryListResponse
		err := client.Post(url, services, &resp, httpclient.DefaultConfig())
		if err != nil {
			logrus.Errorf("Failed to sync with peer %s: %v", peer, err)
			continue
		}
		// 更新本地服务列表（实现最终一致性）
		for _, s := range resp.Services {
			if s.LastHeartbeat.IsZero() {
				s.LastHeartbeat = time.Now()
			}
			r.StoreService(s)
		}
		logrus.Infof("Successfully synced with peer %s", peer)
	}
}
