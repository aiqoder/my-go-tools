package oauth2

import (
	"testing"
)

// BenchmarkOAuth2Service_ExchangeCodeForToken 授权码换令牌基准测试
func BenchmarkOAuth2Service_ExchangeCodeForToken(b *testing.B) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ExchangeCodeForToken("test-code")
	}
}

// BenchmarkOAuth2Service_GetUserInfo 获取用户信息基准测试
func BenchmarkOAuth2Service_GetUserInfo(b *testing.B) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetUserInfo("mock-access-token")
	}
}

// BenchmarkOAuth2Service_RefreshToken 刷新令牌基准测试
func BenchmarkOAuth2Service_RefreshToken(b *testing.B) {
	mock := NewMockServer()
	defer mock.Close()

	cfg := &Config{
		Server:       mock.URL(),
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.RefreshToken("mock-refresh-token")
	}
}

// BenchmarkOAuth2Service_GetConfig 获取配置基准测试
func BenchmarkOAuth2Service_GetConfig(b *testing.B) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.GetConfig()
	}
}

// BenchmarkOAuth2Service_BuildAuthorizeURL 构建授权 URL 基准测试
func BenchmarkOAuth2Service_BuildAuthorizeURL(b *testing.B) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.BuildAuthorizeURL("test-state", "read")
	}
}

// BenchmarkExtractBearerToken 提取 Bearer Token 基准测试
func BenchmarkExtractBearerToken(b *testing.B) {
	header := "Bearer mock-access-token-123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractBearerToken(header)
	}
}

// BenchmarkConfig_ToPublic 转换公开配置基准测试
func BenchmarkConfig_ToPublic(b *testing.B) {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.ToPublic()
	}
}
