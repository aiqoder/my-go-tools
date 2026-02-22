// Package uf 提供用户反馈服务 API 的 Go 语言客户端
//
// 该包封装了对 https://uf.yigechengzi.com/ 的 API 调用，
// 采用模块化设计，支持扩展对接该域名下的多个 API。
package uf

import (
	"fmt"
	"net/http"
	"net/url"
)

// ActivityService 活跃度记录服务
//
// 提供软件活跃度记录的创建和更新功能。
// 通过该服务可以记录软件的活跃使用情况。
//
// # 示例
//
//	client := uf.NewClient()
//	activity := client.Activity()
//	resp, err := activity.CreateByGET(1)
type ActivityService struct {
	client *Client
}

// NewActivityService 创建活跃度记录服务
//
// 参数 client 为 UF 服务客户端。
// 返回活跃度记录服务实例。
func NewActivityService(client *Client) *ActivityService {
	return &ActivityService{
		client: client,
	}
}

// CreateByGET 通过 GET 请求创建活跃度记录
//
// 使用 Query 参数方式提交活跃度记录。
// 参数 softwareId 为软件 ID。
//
// # 示例
//
//	resp, err := activity.CreateByGET(1)
func (s *ActivityService) CreateByGET(softwareId uint) (*ActivityResponse, error) {
	path := fmt.Sprintf("/api/activity?%s", url.Values{"softwareId": []string{fmt.Sprintf("%d", softwareId)}}.Encode())
	resp := &ActivityResponse{}
	err := s.client.DoJSONRequest(http.MethodGet, path, nil, resp)
	return resp, err
}

// CreateByPOST 通过 POST 请求创建活跃度记录
//
// 使用 JSON Body 方式提交活跃度记录。
// 参数 softwareId 为软件 ID。
//
// # 示例
//
//	resp, err := activity.CreateByPOST(1)
func (s *ActivityService) CreateByPOST(softwareId uint) (*ActivityResponse, error) {
	req := &ActivityRequest{SoftwareID: softwareId}
	resp := &ActivityResponse{}
	err := s.client.DoJSONRequest(http.MethodPost, "/api/activity", req, resp)
	return resp, err
}
