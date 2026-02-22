package uf

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewClient 测试创建客户端
func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		opts []ClientOption
	}{
		{
			name: "默认配置",
			opts: nil,
		},
		{
			name: "自定义超时",
			opts: []ClientOption{
				WithTimeout(10 * time.Second),
			},
		},
		{
			name: "自定义 BaseURL",
			opts: []ClientOption{
				WithBaseURL("https://custom.example.com/"),
			},
		},
		{
			name: "自定义 HTTP 客户端",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.opts...)
			if client == nil {
				t.Error("NewClient() 不应返回 nil")
			}
		})
	}
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

// ============================================================================
// 客户端功能测试
// ============================================================================

// TestClient_RecordActivity 测试记录活跃度
func TestClient_RecordActivity(t *testing.T) {
	tests := []struct {
		name       string
		softwareID uint
		wantID     uint
		wantOK     bool
	}{
		{
			name:       "成功响应",
			softwareID: 1,
			wantID:     1,
			wantOK:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("期望 POST 方法，实际 %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("期望 Content-Type = application/json，实际 %s", r.Header.Get("Content-Type"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ok": true, "id": 1}`))
			}))
			defer server.Close()

			client := NewClient(WithBaseURL(server.URL))
			resp, err := client.RecordActivity(tt.softwareID)

			if err != nil {
				t.Errorf("RecordActivity() 错误 = %v", err)
			}

			if resp.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v", resp.OK, tt.wantOK)
			}

			if resp.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", resp.ID, tt.wantID)
			}
		})
	}
}

// TestClient_CheckActivation 测试检查激活状态
func TestClient_CheckActivation(t *testing.T) {
	tests := []struct {
		name           string
		softwareID     uint
		machineCode    string
		wantOK         bool
		wantActivated  bool
		wantExpireAt   string
	}{
		{
			name:          "已激活",
			softwareID:    1,
			machineCode:   "ABC-123-XYZ",
			wantOK:        true,
			wantActivated: true,
			wantExpireAt:  "2026-12-31 23:59:59",
		},
		{
			name:          "未激活",
			softwareID:    1,
			machineCode:   "ABC-123-XYZ",
			wantOK:        true,
			wantActivated: false,
			wantExpireAt:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("期望 POST 方法，实际 %s", r.Method)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("期望 Content-Type = application/json，实际 %s", r.Header.Get("Content-Type"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if tt.wantActivated {
					w.Write([]byte(`{"ok": true, "activated": true, "expireAt": "2026-12-31 23:59:59"}`))
				} else {
					w.Write([]byte(`{"ok": true, "activated": false}`))
				}
			}))
			defer server.Close()

			client := NewClient(WithBaseURL(server.URL))
			resp, err := client.CheckActivation(tt.softwareID, tt.machineCode)

			if err != nil {
				t.Errorf("CheckActivation() 错误 = %v", err)
			}

			if resp.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v", resp.OK, tt.wantOK)
			}

			if resp.Activated != tt.wantActivated {
				t.Errorf("Activated = %v, want %v", resp.Activated, tt.wantActivated)
			}

			if resp.ExpireAt != tt.wantExpireAt {
				t.Errorf("ExpireAt = %v, want %v", resp.ExpireAt, tt.wantExpireAt)
			}
		})
	}
}

// TestActivityResponse_IsOK 测试 ActivityResponse.IsOK
func TestActivityResponse_IsOK(t *testing.T) {
	tests := []struct {
		name string
		resp ActivityResponse
		want bool
	}{
		{"成功响应", ActivityResponse{OK: true}, true},
		{"失败响应", ActivityResponse{OK: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsOK(); got != tt.want {
				t.Errorf("IsOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestActivityResponse_HasError 测试 ActivityResponse.HasError
func TestActivityResponse_HasError(t *testing.T) {
	tests := []struct {
		name string
		resp ActivityResponse
		want bool
	}{
		{"无错误", ActivityResponse{OK: true, Error: ""}, false},
		{"有错误", ActivityResponse{OK: false, Error: "some error"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.HasError(); got != tt.want {
				t.Errorf("HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestActivationCheckResponse_HasError 测试 ActivationCheckResponse.HasError
func TestActivationCheckResponse_IsOK(t *testing.T) {
	tests := []struct {
		name string
		resp ActivationCheckResponse
		want bool
	}{
		{"成功响应", ActivationCheckResponse{OK: true}, true},
		{"失败响应", ActivationCheckResponse{OK: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsOK(); got != tt.want {
				t.Errorf("IsOK() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestActivationCheckResponse_HasError 测试 ActivationCheckResponse.HasError
func TestActivationCheckResponse_HasError(t *testing.T) {
	tests := []struct {
		name string
		resp ActivationCheckResponse
		want bool
	}{
		{"无错误", ActivationCheckResponse{OK: true, Error: ""}, false},
		{"有错误", ActivationCheckResponse{OK: false, Error: "some error"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.HasError(); got != tt.want {
				t.Errorf("HasError() = %v, want %v", got, tt.want)
			}
		})
	}
}
