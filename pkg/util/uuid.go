package util

import (
	"errors"
	"github.com/google/uuid"
	"net"
	"strings"
)

func GenerateUUID() string {
	return uuid.New().String()
}

// GetLocalIP 获取一个非环回且非 APIPA 的 IPv4 地址
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// 检查地址是否是 IP 地址且不是环回地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// 检查 IP 是否为 IPv4
			if ip := ipnet.IP.To4(); ip != nil {
				// 排除以 "169.254" 开头的 AP IP A 地址
				if !strings.HasPrefix(ip.String(), "169.254") {
					return ip.String(), nil
				}
			}
		}
	}
	return "", errors.New("cannot find a non-loopback and non-APIPA IPv4 address")
}
