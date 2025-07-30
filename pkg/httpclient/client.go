package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"MicroService/pkg/model" // 假设你的 model 包路径
)

// Client 是一个通用的 HTTP 客户端，用于服务之间的通信
type Client struct {
	httpClient *http.Client
}

// Config 定义 HTTP 客户端的配置
type Config struct {
	Timeout    time.Duration // 请求超时时间
	MaxRetries int           // 最大重试次数
	RetryDelay time.Duration // 重试间隔
}

// DefaultConfig 返回默认的 HTTP 客户端配置
func DefaultConfig() Config {
	return Config{
		Timeout:    5 * time.Second,        // 默认超时 5 秒
		MaxRetries: 2,                      // 默认重试 2 次
		RetryDelay: 500 * time.Millisecond, // 默认重试间隔 500 毫秒
	}
}

// NewClient 创建一个新的 HTTP 客户端
func NewClient(config Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Post 发送 POST 请求，请求体和响应体为 JSON 格式
func (c *Client) Post(url string, request interface{}, response interface{}, config Config) error {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorResp model.ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
				lastErr = fmt.Errorf("server returned error: %s (status: %d)", errorResp.Error, resp.StatusCode)
			} else {
				lastErr = fmt.Errorf("request failed with status: %d", resp.StatusCode)
			}
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return lastErr
		}

		if response != nil {
			if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
				return fmt.Errorf("failed to decode response: %v", err)
			}
		}
		return nil
	}
	return lastErr
}

// Get 发送 GET 请求，响应体为 JSON 格式
func (c *Client) Get(url string, response interface{}, config Config) error {
	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return lastErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errorResp model.ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Error != "" {
				lastErr = fmt.Errorf("server returned error: %s (status: %d)", errorResp.Error, resp.StatusCode)
			} else {
				lastErr = fmt.Errorf("request failed with status: %d", resp.StatusCode)
			}
			if attempt < config.MaxRetries {
				time.Sleep(config.RetryDelay)
				continue
			}
			return lastErr
		}

		if response != nil {
			if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
				return fmt.Errorf("failed to decode response: %v", err)
			}
		}
		return nil
	}
	return lastErr
}
