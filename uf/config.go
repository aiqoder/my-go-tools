// Package uf 提供用户反馈服务 API 的 Go 语言客户端
//
// 该包封装了对 https://uf.yigechengzi.com/ 的 API 调用，
// 采用模块化设计，支持扩展对接该域名下的多个 API。
package uf

import (
	"net/http"
	"strings"
	"time"
)

// 默认配置常量
const (
	// DefaultBaseURL 是 UF 服务的默认基础地址
	DefaultBaseURL = "https://uf.yigechengzi.com/"

	// DefaultTimeout 是默认请求超时时间
	DefaultTimeout = 30 * time.Second
)

// Config 客户端配置
//
// 用于配置 UF 服务客户端的连接参数。所有字段均为可选，
// 未配置时将使用默认值。
type Config struct {
	// BaseURL API 基础地址
	//
	// 默认为 https://uf.yigechengzi.com/
	BaseURL string

	// Timeout 请求超时时间
	//
	// 默认为 30 秒
	Timeout time.Duration

	// HTTPClient 自定义 HTTP 客户端
	//
	// 默认为 nil，将使用默认的 http.Client
	HTTPClient *http.Client
}

// DefaultConfig 返回默认配置
//
// 返回包含默认值的 Config 实例。
// 默认 BaseURL 为 https://uf.yigechengzi.com/，
// 默认超时时间为 30 秒。
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		HTTPClient: nil,
	}
}

// WithBaseURL 设置自定义 BaseURL 的选项函数
//
// 参数 baseURL 为自定义的 API 基础地址。
// 通常不需要修改，默认值已适配生产环境。
func WithBaseURL(baseURL string) func(*Client) {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithTimeout 设置自定义超时时间的选项函数
//
// 参数 timeout 为请求超时时长。
// 对于慢速网络或复杂请求，可适当增加超时时间。
func WithTimeout(timeout time.Duration) func(*Client) {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

// WithHTTPClient 设置自定义 HTTP 客户端的选项函数
//
// 参数 client 为自定义的 *http.Client。
// 用于配置代理、 TLS 证书等高级选项。
func WithHTTPClient(client *http.Client) func(*Client) {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}
