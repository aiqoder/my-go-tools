// Package uf 提供用户反馈服务 API 的 Go 语言客户端
//
// 该包封装了对 https://uf.yigechengzi.com/ 的 API 调用。
//
// # 快速开始
//
//	import "github.com/aiqoder/my-go-tools/uf"
//
//	client := uf.NewClient()
//
//	// 记录活跃度
//	resp, err := client.RecordActivity(1)
//
//	// 检查激活状态
//	resp, err := client.CheckActivation(1, "ABC-123-XYZ")
package uf

import (
	"bytes"
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
	client := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// RecordActivity 记录软件活跃度
//
// 参数 softwareId 为软件 ID。
// 返回活跃度记录响应和错误。
func (c *Client) RecordActivity(softwareId uint) (*ActivityResponse, error) {
	req := &ActivityRequest{SoftwareID: softwareId}
	resp := &ActivityResponse{}
	err := c.doJSONRequest(http.MethodPost, "/api/activity", req, resp)
	return resp, err
}

// CheckActivation 检查软件激活状态
//
// 参数 softwareId 为软件 ID，machineCode 为机器码。
// 返回激活检查响应和错误。
func (c *Client) CheckActivation(softwareId uint, machineCode string) (*ActivationCheckResponse, error) {
	req := &ActivationCheckRequest{
		SoftwareID:  softwareId,
		MachineCode: machineCode,
	}
	resp := &ActivationCheckResponse{}
	err := c.doJSONRequest(http.MethodPost, "/api/activation/check", req, resp)
	return resp, err
}

// buildURL 构建完整请求 URL
func (c *Client) buildURL(path string) string {
	path = strings.TrimLeft(path, "/")
	return c.baseURL + "/" + path
}

// doRequest 发起 HTTP 请求
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.buildURL(path), body)
	if err != nil {
		return nil, NewRequestError(fmt.Sprintf("创建请求失败: %v", err), err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return nil, NewTimeoutError(fmt.Sprintf("请求超时: %v", err))
		}
		return nil, NewNetworkError(fmt.Sprintf("网络请求失败: %v", err), err)
	}

	return resp, nil
}

// doJSONRequest 发起 JSON 请求并解析响应
func (c *Client) doJSONRequest(method, path string, reqBody, respBody interface{}) error {
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if json.Unmarshal(respBytes, &errResp) == nil && errResp.Error != "" {
			return NewServerError(errResp.Error)
		}
		return NewServerError(fmt.Sprintf("HTTP 状态码: %d, 响应: %s", resp.StatusCode, string(respBytes)))
	}

	if respBody != nil {
		if err := json.Unmarshal(respBytes, respBody); err != nil {
			return NewResponseError(fmt.Sprintf("解析响应失败: %v", err), err)
		}
	}

	return nil
}
