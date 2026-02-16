package uf

import (
	"fmt"
)

// ErrorCode 错误码常量
//
// 定义 UF 服务可能返回的错误码。
const (
	// ErrCodeRequestFailed 表示请求失败
	ErrCodeRequestFailed = "REQUEST_FAILED"

	// ErrCodeInvalidResponse 表示响应解析失败
	ErrCodeInvalidResponse = "INVALID_RESPONSE"

	// ErrCodeTimeout 表示请求超时
	ErrCodeTimeout = "TIMEOUT"

	// ErrCodeNetworkError 表示网络错误
	ErrCodeNetworkError = "NETWORK_ERROR"

	// ErrCodeServerError 表示服务器错误
	ErrCodeServerError = "SERVER_ERROR"

	// ErrCodeInvalidParams 表示参数错误
	ErrCodeInvalidParams = "INVALID_PARAMS"
)

// Error UF 服务错误结构
//
// 封装了 UF API 调用过程中可能发生的各种错误。
type Error struct {
	// Code 错误码
	Code string

	// Message 错误信息
	Message string

	// Err 原始错误
	Err error
}

// Error 实现 error 接口
//
// 返回格式化的错误信息，包含错误码和错误消息。
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
//
// 用于错误链解包，支持 errors.As 等错误处理方式。
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError 创建新的 UF 错误
//
// 参数 code 为错误码，message 为错误描述，err 为原始错误（可传 nil）。
func NewError(code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewRequestError 创建请求错误
//
// 用于创建网络请求相关的错误。
func NewRequestError(message string, err error) *Error {
	return NewError(ErrCodeRequestFailed, message, err)
}

// NewResponseError 创建响应解析错误
//
// 用于创建响应体解析失败相关的错误。
func NewResponseError(message string, err error) *Error {
	return NewError(ErrCodeInvalidResponse, message, err)
}

// NewTimeoutError 创建超时错误
//
// 用于创建请求超时相关的错误。
func NewTimeoutError(message string) *Error {
	return NewError(ErrCodeTimeout, message, nil)
}

// NewNetworkError 创建网络错误
//
// 用于创建网络连接相关的错误。
func NewNetworkError(message string, err error) *Error {
	return NewError(ErrCodeNetworkError, message, err)
}

// NewServerError 创建服务器错误
//
// 用于创建服务器返回错误响应相关的错误。
func NewServerError(message string) *Error {
	return NewError(ErrCodeServerError, message, nil)
}

// NewParamsError 创建参数错误
//
// 用于创建请求参数相关的错误。
func NewParamsError(message string) *Error {
	return NewError(ErrCodeInvalidParams, message, nil)
}
