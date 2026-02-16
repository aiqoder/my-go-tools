package uf

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewClient 测试创建客户端
func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		opts     []ClientOption
		wantURL  string
		wantTime time.Duration
	}{
		{
			name:     "默认配置",
			opts:     nil,
			wantURL:  DefaultBaseURL,
			wantTime: DefaultTimeout,
		},
		{
			name: "自定义超时",
			opts: []ClientOption{
				WithTimeout(60 * time.Second),
			},
			wantURL:  DefaultBaseURL,
			wantTime: 60 * time.Second,
		},
		{
			name: "自定义 BaseURL",
			opts: []ClientOption{
				WithBaseURL("https://custom.example.com/"),
			},
			wantURL:  "https://custom.example.com",
			wantTime: DefaultTimeout,
		},
		{
			name: "自定义 HTTP 客户端",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
			},
			wantURL:  DefaultBaseURL,
			wantTime: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.opts...)

			if client.GetBaseURL() != tt.wantURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.GetBaseURL(), tt.wantURL)
			}

			if client.GetHTTPClient().Timeout != tt.wantTime {
				t.Errorf("NewClient() timeout = %v, want %v", client.GetHTTPClient().Timeout, tt.wantTime)
			}
		})
	}
}

// TestClient_GetBaseURL 测试获取 BaseURL
func TestClient_GetBaseURL(t *testing.T) {
	client := NewClient()
	want := DefaultBaseURL

	if got := client.GetBaseURL(); got != want {
		t.Errorf("GetBaseURL() = %v, want %v", got, want)
	}
}

// TestClient_SetHTTPClient 测试设置自定义 HTTP 客户端
func TestClient_SetHTTPClient(t *testing.T) {
	client := NewClient()
	customClient := &http.Client{Timeout: 15 * time.Second}

	client.SetHTTPClient(customClient)

	if client.GetHTTPClient().Timeout != 15*time.Second {
		t.Errorf("SetHTTPClient() failed, timeout = %v", client.GetHTTPClient().Timeout)
	}
}

// TestClient_Get 测试 GET 请求
func TestClient_Get(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("期望 GET 方法，实际 %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	data, err := client.Get("/test")

	if err != nil {
		t.Errorf("Get() 错误 = %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Errorf("解析响应失败 = %v", err)
	}

	if !resp.OK {
		t.Error("期望响应 OK = true")
	}
}

// TestClient_Post 测试 POST 请求
func TestClient_Post(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("期望 Content-Type = application/json，实际 %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	data, err := client.Post("/test", map[string]interface{}{"key": "value"})

	if err != nil {
		t.Errorf("Post() 错误 = %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Errorf("解析响应失败 = %v", err)
	}

	if !resp.OK {
		t.Error("期望响应 OK = true")
	}
}

// TestClient_PostForm 测试表单 POST 请求
func TestClient_PostForm(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法，实际 %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("期望 Content-Type = application/x-www-form-urlencoded，实际 %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	data, err := client.PostForm("/test", map[string][]string{
		"key": {"value"},
	})

	if err != nil {
		t.Errorf("PostForm() 错误 = %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Errorf("解析响应失败 = %v", err)
	}

	if !resp.OK {
		t.Error("期望响应 OK = true")
	}
}

// TestClient_ErrorHandling 测试错误处理
func TestClient_ErrorHandling(t *testing.T) {
	// 测试无效 URL
	t.Run("无效 URL", func(t *testing.T) {
		client := NewClient(WithBaseURL("://invalid"))
		_, err := client.Get("/test")

		if err == nil {
			t.Error("期望错误，实际无错误")
		}
	})

	// 测试服务器返回错误
	t.Run("服务器错误", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"ok": false, "error": "参数错误"}`))
		}))
		defer server.Close()

		client := NewClient(WithBaseURL(server.URL))
		_, err := client.Get("/test")

		if err == nil {
			t.Error("期望错误，实际无错误")
		}
	})
}

// TestError_Error 测试错误类型
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantPart string
	}{
		{
			name: "带原始错误",
			err: &Error{
				Code:    ErrCodeRequestFailed,
				Message: "请求失败",
				Err:     assertErr("original error"),
			},
			wantPart: "请求失败",
		},
		{
			name: "无原始错误",
			err: &Error{
				Code:    ErrCodeTimeout,
				Message: "超时",
				Err:     nil,
			},
			wantPart: "超时",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			if tt.wantPart != "" && !contains(errStr, tt.wantPart) {
				t.Errorf("Error() = %v, want contain %v", errStr, tt.wantPart)
			}
		})
	}
}

// TestError_Unwrap 测试错误解包
func TestError_Unwrap(t *testing.T) {
	originalErr := assertErr("original error")
	err := &Error{
		Code:    ErrCodeRequestFailed,
		Message: "请求失败",
		Err:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

// TestNewErrorFunctions 测试错误创建函数
func TestNewErrorFunctions(t *testing.T) {
	tests := []struct {
		name string
		fn   func() *Error
		want string
	}{
		{
			name: "NewRequestError",
			fn:   func() *Error { return NewRequestError("test", nil) },
			want: ErrCodeRequestFailed,
		},
		{
			name: "NewResponseError",
			fn:   func() *Error { return NewResponseError("test", nil) },
			want: ErrCodeInvalidResponse,
		},
		{
			name: "NewTimeoutError",
			fn:   func() *Error { return NewTimeoutError("test") },
			want: ErrCodeTimeout,
		},
		{
			name: "NewNetworkError",
			fn:   func() *Error { return NewNetworkError("test", nil) },
			want: ErrCodeNetworkError,
		},
		{
			name: "NewServerError",
			fn:   func() *Error { return NewServerError("test") },
			want: ErrCodeServerError,
		},
		{
			name: "NewParamsError",
			fn:   func() *Error { return NewParamsError("test") },
			want: ErrCodeInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err.Code != tt.want {
				t.Errorf("Code = %v, want %v", err.Code, tt.want)
			}
		})
	}
}

// TestResponse_IsOK 测试 Response.IsOK
func TestResponse_IsOK(t *testing.T) {
	tests := []struct {
		name string
		resp Response
		want bool
	}{
		{"成功响应", Response{OK: true}, true},
		{"失败响应", Response{OK: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsOK(); got != tt.want {
				t.Errorf("IsOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResponse_HasError 测试 Response.HasError
func TestResponse_HasError(t *testing.T) {
	tests := []struct {
		name string
		resp Response
		want bool
	}{
		{"无错误", Response{OK: true, Error: ""}, false},
		{"有错误", Response{OK: false, Error: "some error"}, true},
		{"OK 但有错误信息", Response{OK: true, Error: "warning"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.HasError(); got != tt.want {
				t.Errorf("HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConfig_DefaultConfig 测试默认配置
func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL = %v, want %v", cfg.BaseURL, DefaultBaseURL)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}

	if cfg.HTTPClient != nil {
		t.Error("HTTPClient should be nil by default")
	}
}

// TestWithBaseURL 测试 WithBaseURL 选项函数
func TestWithBaseURL(t *testing.T) {
	opt := WithBaseURL("https://test.com/")
	client := NewClient()
	opt(client)

	if client.baseURL != "https://test.com" {
		t.Errorf("baseURL = %v, want https://test.com", client.baseURL)
	}

	// 测试空值不覆盖
	optEmpty := WithBaseURL("")
	client2 := &Client{baseURL: "original"}
	optEmpty(client2)

	if client2.baseURL != "original" {
		t.Errorf("baseURL should not be changed, got %v", client2.baseURL)
	}
}

// TestWithTimeout 测试 WithTimeout 选项函数
func TestWithTimeout(t *testing.T) {
	opt := WithTimeout(60 * time.Second)
	client := NewClient()
	opt(client)

	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("Timeout = %v, want 60s", client.httpClient.Timeout)
	}

	// 测试零值不覆盖
	optZero := WithTimeout(0)
	client2 := &Client{httpClient: &http.Client{Timeout: 30 * time.Second}}
	optZero(client2)

	if client2.httpClient.Timeout != 30*time.Second {
		t.Errorf("Timeout should not be changed, got %v", client2.httpClient.Timeout)
	}
}

// TestWithHTTPClient 测试 WithHTTPClient 选项函数
func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{}
	opt := WithHTTPClient(customClient)
	client := NewClient()
	opt(client)

	if client.httpClient != customClient {
		t.Error("httpClient should be set to custom client")
	}

	// 测试 nil 不覆盖
	optNil := WithHTTPClient(nil)
	client2 := &Client{httpClient: &http.Client{}}
	optNil(client2)

	if client2.httpClient == nil {
		t.Error("httpClient should not be changed to nil")
	}
}

// 辅助函数

// assertErr 返回一个错误用于测试
func assertErr(msg string) error {
	return &testError{msg: msg}
}

// testError 测试用错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// contains 检查 s 是否包含 substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
