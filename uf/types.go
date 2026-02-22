package uf

// Response 通用响应结构
//
// 所有 API 调用返回的通用响应格式。
// 遵循 UF 服务的统一响应规范。
type Response struct {
	// OK 请求是否成功
	//
	// true 表示成功，false 表示失败
	OK bool `json:"ok"`

	// Error 错误信息
	//
	// 仅在 OK 为 false 时存在
	Error string `json:"error,omitempty"`
}

// IsOK 检查响应是否成功
//
// 返回 true 表示 API 调用成功，false 表示失败。
func (r *Response) IsOK() bool {
	return r.OK
}

// HasError 检查响应是否包含错误信息
//
// 返回 true 表示存在错误，false 表示无错误。
func (r *Response) HasError() bool {
	return !r.OK && r.Error != ""
}

// ErrorResponse 错误响应结构
//
// 当 API 返回错误时使用的响应格式。
type ErrorResponse struct {
	// OK 请求是否成功
	OK bool `json:"ok"`

	// Error 错误信息
	Error string `json:"error"`
}

// NewErrorResponse 创建错误响应
//
// 参数 message 为错误信息。
// 返回包含错误信息的 ErrorResponse 实例。
func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		OK:    false,
		Error: message,
	}
}

// ============================================================================
// 活跃度记录相关类型
// ============================================================================

// ActivityRequest 活跃度记录请求
//
// 用于创建或更新软件的活跃度记录。
type ActivityRequest struct {
	// SoftwareID 软件 ID
	SoftwareID uint `json:"softwareId"`
}

// ActivityResponse 活跃度记录响应
//
// 活跃度记录接口的响应格式。
type ActivityResponse struct {
	// OK 请求是否成功
	//
	// true 表示成功，false 表示失败
	OK bool `json:"ok"`

	// ID 活跃度记录 ID
	//
	// 仅在 OK 为 true 时存在
	ID uint `json:"id,omitempty"`

	// Error 错误信息
	//
	// 仅在 OK 为 false 时存在
	Error string `json:"error,omitempty"`
}

// IsOK 检查响应是否成功
//
// 返回 true 表示 API 调用成功，false 表示失败。
func (r *ActivityResponse) IsOK() bool {
	return r.OK
}

// HasError 检查响应是否包含错误信息
//
// 返回 true 表示存在错误，false 表示无错误。
func (r *ActivityResponse) HasError() bool {
	return !r.OK && r.Error != ""
}

// ============================================================================
// 激活检查相关类型
// ============================================================================

// ActivationCheckRequest 激活检查请求
//
// 用于检查软件是否已激活及激活状态。
type ActivationCheckRequest struct {
	// SoftwareID 软件 ID
	SoftwareID uint `json:"softwareId"`

	// MachineCode 机器码
	MachineCode string `json:"machineCode"`
}

// ActivationCheckResponse 激活检查响应
//
// 激活检查接口的响应格式。
type ActivationCheckResponse struct {
	// OK 请求是否成功
	//
	// true 表示成功，false 表示失败
	OK bool `json:"ok"`

	// Activated 是否已激活
	//
	// true 表示已激活，false 表示未激活
	Activated bool `json:"activated"`

	// ExpireAt 过期时间
	//
	// 格式为 "YYYY-MM-DD HH:MM:SS"，仅在 OK 为 true 时存在
	ExpireAt string `json:"expireAt,omitempty"`

	// Error 错误信息
	//
	// 仅在 OK 为 false 时存在
	Error string `json:"error,omitempty"`
}

// IsOK 检查响应是否成功
//
// 返回 true 表示 API 调用成功，false 表示失败。
func (r *ActivationCheckResponse) IsOK() bool {
	return r.OK
}

// HasError 检查响应是否包含错误信息
//
// 返回 true 表示存在错误，false 表示无错误。
func (r *ActivationCheckResponse) HasError() bool {
	return !r.OK && r.Error != ""
}
