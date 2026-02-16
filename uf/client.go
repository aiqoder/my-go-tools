// Package uf 提供用户反馈服务 API 的 Go 语言客户端
//
// 该包封装了对 https://uf.yigechengzi.com/ 的 API 调用，
// 采用模块化设计，支持扩展对接该域名下的多个 API。
//
// # 快速开始
//
//	import "github.com/aiqoder/go-tools/uf"
//
//	// 创建默认配置的客户端
//	client := uf.NewClient()
//
//	// 或者使用选项自定义配置
//	client := uf.NewClient(
//	    uf.WithTimeout(60 * time.Second),
//	)
package uf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client UF API 客户端
//
// 提供对用户反馈服务的统一访问入口。
// 客户端是线程安全的，可在多个 goroutine 中并发使用。
//
// # 示例
//
//	client := uf.NewClient()
//	baseURL := client.GetBaseURL()
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientOption 客户端配置选项函数
//
// 用于在创建客户端时自定义配置。
type ClientOption func(*Client)

// NewClient 创建 UF 服务客户端
//
// 默认 BaseURL 为 https://uf.yigechengzi.com/，
// 默认超时时间为 30 秒。
//
// # 选项支持
//
//	uf.NewClient(
//	    uf.WithTimeout(60 * time.Second),     // 自定义超时
//	    uf.WithHTTPClient(customClient),        // 自定义 HTTP 客户端
//	)
func NewClient(opts ...ClientOption) *Client {
	// 创建默认配置的客户端
	client := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	// 应用选项
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// GetBaseURL 返回客户端配置的 BaseURL
//
// 返回当前客户端配置的 API 基础地址。
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// SetHTTPClient 设置自定义 HTTP 客户端
//
// 参数 client 为自定义的 *http.Client。
// 用于配置代理、 TLS 证书等高级选项。
func (c *Client) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

// GetHTTPClient 返回当前的 HTTP 客户端
//
// 返回客户端内部使用的 *http.Client，可用于进一步配置。
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// buildURL 构建完整请求 URL
//
// 参数 path 为 API 路径（相对于 BaseURL）。
// 返回完整的请求 URL 字符串。
func (c *Client) buildURL(path string) string {
	path = strings.TrimLeft(path, "/")
	return c.baseURL + "/" + path
}

// doRequest 发起 HTTP 请求
//
// 参数 method 为 HTTP 方法（如 GET、POST），
// path 为 API 路径，body 为请求体（可为 nil）。
//
// 返回 *http.Response 和错误。
// 调用方必须自行关闭响应体。
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.buildURL(path), body)
	if err != nil {
		return nil, NewRequestError(fmt.Sprintf("创建请求失败: %v", err), err)
	}

	// 设置默认请求头
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// 判断是否为超时错误
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return nil, NewTimeoutError(fmt.Sprintf("请求超时: %v", err))
		}
		return nil, NewNetworkError(fmt.Sprintf("网络请求失败: %v", err), err)
	}

	return resp, nil
}

// doRequestWithContext 发起带上下文的 HTTP 请求
//
// 参数 ctx 为请求上下文，method 为 HTTP 方法，
// path 为 API 路径，body 为请求体（可为 nil）。
//
// 返回 *http.Response 和错误。
// 调用方必须自行关闭响应体。
func (c *Client) doRequestWithContext(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path), body)
	if err != nil {
		return nil, NewRequestError(fmt.Sprintf("创建请求失败: %v", err), err)
	}

	// 设置默认请求头
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// 判断是否为超时错误
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return nil, NewTimeoutError(fmt.Sprintf("请求超时: %v", err))
		}
		return nil, NewNetworkError(fmt.Sprintf("网络请求失败: %v", err), err)
	}

	return resp, nil
}

// Get 发起 GET 请求
//
// 参数 path 为 API 路径。
// 返回响应体和错误。
func (c *Client) Get(path string) ([]byte, error) {
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// Post 发起 POST 请求
//
// 参数 path 为 API 路径，data 为请求体数据。
// 返回响应体和错误。
func (c *Client) Post(path string, data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, NewParamsError(fmt.Sprintf("序列化请求体失败: %v", err))
	}

	resp, err := c.doRequest(http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// PostForm 发起 Form 表单 POST 请求
//
// 参数 path 为 API 路径，data 为表单数据。
// 返回响应体和错误。
func (c *Client) PostForm(path string, data url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, c.buildURL(path), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, NewRequestError(fmt.Sprintf("创建请求失败: %v", err), err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewNetworkError(fmt.Sprintf("网络请求失败: %v", err), err)
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// Put 发起 PUT 请求
//
// 参数 path 为 API 路径，data 为请求体数据。
// 返回响应体和错误。
func (c *Client) Put(path string, data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, NewParamsError(fmt.Sprintf("序列化请求体失败: %v", err))
	}

	resp, err := c.doRequest(http.MethodPut, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// Delete 发起 DELETE 请求
//
// 参数 path 为 API 路径。
// 返回响应体和错误。
func (c *Client) Delete(path string) ([]byte, error) {
	resp, err := c.doRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// readResponseBody 读取响应体
//
// 参数 resp 为 HTTP 响应。
// 返回响应体字节数据和错误。
func readResponseBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewResponseError(fmt.Sprintf("读取响应失败: %v", err), err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// 尝试解析错误响应
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, NewServerError(errResp.Error)
		}
		return nil, NewServerError(fmt.Sprintf("HTTP 状态码: %d, 响应: %s", resp.StatusCode, string(body)))
	}

	return body, nil
}

// DoJSONRequest 发起 JSON 请求并解析响应
//
// 参数 method 为 HTTP 方法，path 为 API 路径，
// reqBody 为请求体（传 nil 表示无请求体），respBody 为响应体目标结构。
// 返回错误。
func (c *Client) DoJSONRequest(method, path string, reqBody, respBody interface{}) error {
	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return NewParamsError(fmt.Sprintf("序列化请求体失败: %v", err))
		}
		body = bytes.NewReader(data)
	}

	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewResponseError(fmt.Sprintf("读取响应失败: %v", err), err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != "" {
			return NewServerError(errResp.Error)
		}
		return NewServerError(fmt.Sprintf("HTTP 状态码: %d, 响应: %s", resp.StatusCode, string(respBytes)))
	}

	// 解析响应体
	if respBody != nil {
		if err := json.Unmarshal(respBytes, respBody); err != nil {
			return NewResponseError(fmt.Sprintf("解析响应失败: %v", err), err)
		}
	}

	return nil
}

// GetWithContext 带上下文发起 GET 请求
//
// 参数 ctx 为请求上下文，path 为 API 路径。
// 返回响应体和错误。
func (c *Client) GetWithContext(ctx context.Context, path string) ([]byte, error) {
	resp, err := c.doRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}

// PostWithContext 带上下文发起 POST 请求
//
// 参数 ctx 为请求上下文，path 为 API 路径，data 为请求体数据。
// 返回响应体和错误。
func (c *Client) PostWithContext(ctx context.Context, path string, data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, NewParamsError(fmt.Sprintf("序列化请求体失败: %v", err))
	}

	resp, err := c.doRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readResponseBody(resp)
}
