// Package uf 提供用户反馈服务 API 的 Go 语言客户端
//
// 该包封装了对 https://uf.yigechengzi.com/ 的 API 调用，
// 采用模块化设计，支持扩展对接该域名下的多个 API。
package uf

import (
	"net/http"
)

// ActivationService 激活检查服务
//
// 提供软件激活状态检查功能。
// 用于验证软件是否已激活及获取激活到期时间。
//
// # 示例
//
//	client := uf.NewClient()
//	activation := client.Activation()
//	resp, err := activation.Check(1, "ABC-123-XYZ")
type ActivationService struct {
	client *Client
}

// NewActivationService 创建激活检查服务
//
// 参数 client 为 UF 服务客户端。
// 返回激活检查服务实例。
func NewActivationService(client *Client) *ActivationService {
	return &ActivationService{
		client: client,
	}
}

// Check 检查软件激活状态
//
// 参数 softwareId 为软件 ID，machineCode 为机器码。
// 返回激活检查响应和错误。
//
// # 示例
//
//	resp, err := activation.Check(1, "ABC-123-XYZ")
//	if err != nil {
//	    // 处理错误
//	}
//	if resp.Activated {
//	    fmt.Printf("软件已激活，到期时间: %s\n", resp.ExpireAt)
//	}
func (s *ActivationService) Check(softwareId uint, machineCode string) (*ActivationCheckResponse, error) {
	req := &ActivationCheckRequest{
		SoftwareID:  softwareId,
		MachineCode: machineCode,
	}
	resp := &ActivationCheckResponse{}
	err := s.client.DoJSONRequest(http.MethodPost, "/api/activation/check", req, resp)
	return resp, err
}
