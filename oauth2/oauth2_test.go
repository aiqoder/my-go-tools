package oauth2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTokenResponse_ExpiresAt(t *testing.T) {
	resp := &TokenResponse{
		ExpiresIn: 3600,
	}

	expiresAt := resp.ExpiresAt()
	if expiresAt.IsZero() {
		t.Error("ExpiresAt 不应返回零值")
	}
}

func TestTokenResponse_RefreshExpiresAt(t *testing.T) {
	resp := &TokenResponse{
		RefreshExpiresIn: 86400,
	}

	expiresAt := resp.RefreshExpiresAt()
	if expiresAt.IsZero() {
		t.Error("RefreshExpiresAt 不应返回零值")
	}
}

func TestOAuth2Error_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *OAuth2Error
		expected string
	}{
		{
			name:     "有错误描述",
			err:      &OAuth2Error{Code: "invalid_request", ErrorDescription: "缺少参数"},
			expected: "invalid_request: 缺少参数",
		},
		{
			name:     "无错误描述",
			err:      &OAuth2Error{Code: "invalid_request"},
			expected: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("OAuth2Error.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_ToPublic(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	public := cfg.ToPublic()

	if public.Server != cfg.Server {
		t.Errorf("Server 不匹配: got %v, want %v", public.Server, cfg.Server)
	}
	if public.ClientID != cfg.ClientID {
		t.Errorf("ClientID 不匹配: got %v, want %v", public.ClientID, cfg.ClientID)
	}
	if public.RedirectURI != cfg.RedirectURI {
		t.Errorf("RedirectURI 不匹配: got %v, want %v", public.RedirectURI, cfg.RedirectURI)
	}
	if public.ClientID == cfg.ClientSecret {
		t.Error("公开配置不应包含 client_secret")
	}
}

func TestTokenResponseBody_ToTokenResponse(t *testing.T) {
	body := &TokenResponseBody{
		AccessToken:      "access-token",
		TokenType:        "Bearer",
		ExpiresIn:        3600,
		RefreshToken:     "refresh-token",
		RefreshExpiresIn: 86400,
		Scope:            "read",
	}

	resp := body.ToTokenResponse()

	if resp.AccessToken != body.AccessToken {
		t.Errorf("AccessToken 不匹配: got %v, want %v", resp.AccessToken, body.AccessToken)
	}
	if resp.TokenType != body.TokenType {
		t.Errorf("TokenType 不匹配: got %v, want %v", resp.TokenType, body.TokenType)
	}
	if resp.ExpiresIn != body.ExpiresIn {
		t.Errorf("ExpiresIn 不匹配: got %v, want %v", resp.ExpiresIn, body.ExpiresIn)
	}
	if resp.RefreshToken != body.RefreshToken {
		t.Errorf("RefreshToken 不匹配: got %v, want %v", resp.RefreshToken, body.RefreshToken)
	}
	if resp.RefreshExpiresIn != body.RefreshExpiresIn {
		t.Errorf("RefreshExpiresIn 不匹配: got %v, want %v", resp.RefreshExpiresIn, body.RefreshExpiresIn)
	}
	if resp.Scope != body.Scope {
		t.Errorf("Scope 不匹配: got %v, want %v", resp.Scope, body.Scope)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "标准 Bearer 格式",
			header:   "Bearer token123",
			expected: "token123",
		},
		{
			name:     "无 Bearer 前缀",
			header:   "token123",
			expected: "token123",
		},
		{
			name:     "空 Header",
			header:   "",
			expected: "",
		},
		{
			name:     "Bearer 后面无空格",
			header:   "Bearertoken",
			expected: "Bearertoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBearerToken(tt.header)
			if got != tt.expected {
				t.Errorf("extractBearerToken(%q) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}
}

// MockServer 创建模拟的 OAuth2 服务器
type MockServer struct {
	server      *httptest.Server
	tokenResp   TokenResponse
	userInfo    UserInfo
	tokenActive bool
}

func NewMockServer() *MockServer {
	ms := &MockServer{
		tokenResp: TokenResponse{
			AccessToken:      "mock-access-token",
			TokenType:        "Bearer",
			ExpiresIn:        3600,
			RefreshToken:     "mock-refresh-token",
			RefreshExpiresIn: 86400,
			Scope:            "read",
		},
		userInfo: UserInfo{
			Sub:      "user123",
			Username: "testuser",
			Status:   1,
			ClientID: "app123",
		},
		tokenActive: true,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", ms.handleToken)
	mux.HandleFunc("/oauth2/userinfo", ms.handleUserInfo)
	mux.HandleFunc("/oauth2/introspect", ms.handleIntrospect)

	ms.server = httptest.NewServer(mux)
	return ms
}

func (ms *MockServer) handleToken(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	grantType := r.Form.Get("grant_type")

	if grantType == "authorization_code" {
		json.NewEncoder(w).Encode(ms.tokenResp)
	} else if grantType == "refresh_token" {
		json.NewEncoder(w).Encode(ms.tokenResp)
	} else {
		http.Error(w, "invalid grant_type", http.StatusBadRequest)
	}
}

func (ms *MockServer) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(ms.userInfo)
}

func (ms *MockServer) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": ms.tokenActive})
}

func (ms *MockServer) Close() {
	ms.server.Close()
}

func (ms *MockServer) URL() string {
	return ms.server.URL
}

// TestOAuth2Service_ExchangeCodeForToken 测试授权码换令牌
func TestOAuth2Service_ExchangeCodeForToken(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	resp, err := svc.ExchangeCodeForToken("test-code")
	if err != nil {
		t.Errorf("ExchangeCodeForToken 失败: %v", err)
	}

	if resp.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken 不匹配: got %v, want %v", resp.AccessToken, "mock-access-token")
	}
}

func TestOAuth2Service_ExchangeCodeForToken_EmptyCode(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	_, err := svc.ExchangeCodeForToken("")
	if err == nil {
		t.Error("空授权码应返回错误")
	}
}

func TestOAuth2Service_GetUserInfo(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	info, err := svc.GetUserInfo("mock-access-token")
	if err != nil {
		t.Errorf("GetUserInfo 失败: %v", err)
	}

	if info.Sub != "user123" {
		t.Errorf("Sub 不匹配: got %v, want %v", info.Sub, "user123")
	}
	if info.Username != "testuser" {
		t.Errorf("Username 不匹配: got %v, want %v", info.Username, "testuser")
	}
}

func TestOAuth2Service_GetUserInfo_EmptyToken(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	_, err := svc.GetUserInfo("")
	if err == nil {
		t.Error("空令牌应返回错误")
	}
}

func TestOAuth2Service_RefreshToken(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	resp, err := svc.RefreshToken("mock-refresh-token")
	if err != nil {
		t.Errorf("RefreshToken 失败: %v", err)
	}

	if resp.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken 不匹配: got %v, want %v", resp.AccessToken, "mock-access-token")
	}
}

func TestOAuth2Service_GetConfig(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	public := svc.GetConfig()

	if public.Server != cfg.Server {
		t.Errorf("Server 不匹配: got %v, want %v", public.Server, cfg.Server)
	}
	if public.ClientID != cfg.ClientID {
		t.Errorf("ClientID 不匹配: got %v, want %v", public.ClientID, cfg.ClientID)
	}
	if public.RedirectURI != cfg.RedirectURI {
		t.Errorf("RedirectURI 不匹配: got %v, want %v", public.RedirectURI, cfg.RedirectURI)
	}
}

func TestOAuth2Service_BuildAuthorizeURL(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	url := svc.BuildAuthorizeURL("test-state", "read")

	if !strings.Contains(url, "client_id=test-client") {
		t.Errorf("授权 URL 应包含 client_id: %v", url)
	}
	if !strings.Contains(url, "redirect_uri=") {
		t.Errorf("授权 URL 应包含 redirect_uri: %v", url)
	}
	if !strings.Contains(url, "response_type=code") {
		t.Errorf("授权 URL 应包含 response_type=code: %v", url)
	}
	if !strings.Contains(url, "state=test-state") {
		t.Errorf("授权 URL 应包含 state: %v", url)
	}
}

func TestOAuth2Handler_GetConfig(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/oauth2/config", handler.GetConfig)

	// 发送测试请求
	req := httptest.NewRequest("GET", "/api/oauth2/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusOK)
	}

	var resp PublicConfig
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if resp.ClientID != "test-client" {
		t.Errorf("ClientID 不匹配: got %v, want %v", resp.ClientID, "test-client")
	}
}

func TestOAuth2Handler_Callback(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/oauth2/callback", handler.Callback)

	// 发送测试请求
	body := `{"code":"test-code"}`
	req := httptest.NewRequest("POST", "/api/oauth2/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusOK)
	}

	var resp TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if resp.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken 不匹配: got %v, want %v", resp.AccessToken, "mock-access-token")
	}
}

func TestOAuth2Handler_Callback_MissingCode(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/oauth2/callback", handler.Callback)

	// 发送测试请求（无 code）
	body := `{}`
	req := httptest.NewRequest("POST", "/api/oauth2/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestOAuth2Handler_GetUserInfo(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/oauth2/userinfo", handler.GetUserInfo)

	// 发送测试请求
	req := httptest.NewRequest("GET", "/api/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer mock-access-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusOK)
	}

	var resp UserInfo
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if resp.Sub != "user123" {
		t.Errorf("Sub 不匹配: got %v, want %v", resp.Sub, "user123")
	}
}

func TestOAuth2Handler_GetUserInfo_NoToken(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/oauth2/userinfo", handler.GetUserInfo)

	// 发送测试请求（无 token）
	req := httptest.NewRequest("GET", "/api/oauth2/userinfo", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestOAuth2Handler_RefreshToken(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/oauth2/refresh", handler.RefreshToken)

	// 发送测试请求
	body := `{"refresh_token":"mock-refresh-token"}`
	req := httptest.NewRequest("POST", "/api/oauth2/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusOK)
	}

	var resp TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("解析响应失败: %v", err)
	}

	if resp.AccessToken != "mock-access-token" {
		t.Errorf("AccessToken 不匹配: got %v, want %v", resp.AccessToken, "mock-access-token")
	}
}

func TestOAuth2Handler_RefreshToken_MissingToken(t *testing.T) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/oauth2/refresh", handler.RefreshToken)

	// 发送测试请求（无 refresh_token）
	body := `{}`
	req := httptest.NewRequest("POST", "/api/oauth2/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("状态码不匹配: got %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// TestOAuth2Service_DefaultServer 测试默认 Server
func TestOAuth2Service_DefaultServer(t *testing.T) {
	// 不设置 Server，使用默认值
	cfg := &Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	// 验证使用了默认 Server
	if svc.GetServer() != defaultOAuth2Server {
		t.Errorf("默认 Server 不匹配: got %v, want %v", svc.GetServer(), defaultOAuth2Server)
	}

	// 验证自定义 Server 仍然有效
	cfg2 := &Config{
		Server:       "http://custom-server.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc2 := NewOAuth2Service(cfg2)

	if svc2.GetServer() != "http://custom-server.com" {
		t.Errorf("自定义 Server 不匹配: got %v, want %v", svc2.GetServer(), "http://custom-server.com")
	}
}
