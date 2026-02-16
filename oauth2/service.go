package oauth2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 默认 OAuth2 服务器地址
const defaultOAuth2Server = "https://uf.yigechengzi.com/"

// OAuth2Service OAuth2 服务层
//
// 封装 OAuth2 核心业务逻辑，包括授权码换令牌、获取用户信息、刷新令牌等功能
type OAuth2Service struct {
	oauth2Server string
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// ServiceOption 服务配置选项
type ServiceOption func(*OAuth2Service)

// WithHTTPClient 自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) ServiceOption {
	return func(s *OAuth2Service) {
		s.httpClient = client
	}
}

// NewOAuth2Service 创建 OAuth2 服务实例
//
// 配置通过 Config 结构体传入，支持自定义 HTTP 客户端
// 如果未设置 Server，将使用默认值 https://uf.yigechengzi.com/
func NewOAuth2Service(cfg *Config, opts ...ServiceOption) *OAuth2Service {
	// 如果未设置 Server，则使用默认值
	server := cfg.Server
	if server == "" {
		server = defaultOAuth2Server
	}

	s := &OAuth2Service{
		oauth2Server: server,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectURI,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// ExchangeCodeForToken 使用授权码换取访问令牌
//
// 参数 code 为 OAuth2 授权服务器返回的授权码
// 返回 TokenResponse 包含访问令牌和刷新令牌
func (s *OAuth2Service) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("授权码不能为空")
	}

	tokenURL := s.oauth2Server + "/oauth2/token"

	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("client_id", s.clientID)
	formData.Set("client_secret", s.clientSecret)
	formData.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var oauthErr OAuth2Error
		if json.Unmarshal(body, &oauthErr) == nil && oauthErr.Code != "" {
			return nil, fmt.Errorf("OAuth2 错误: %s - %s", oauthErr.Code, oauthErr.ErrorDescription)
		}
		return nil, fmt.Errorf("令牌交换失败，HTTP 状态码: %d", resp.StatusCode)
	}

	var tokenResp TokenResponseBody
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("令牌响应中缺少 access_token")
	}

	return tokenResp.ToTokenResponse(), nil
}

// GetUserInfo 使用访问令牌获取用户信息
//
// 参数 accessToken 为有效的访问令牌
// 返回 UserInfo 包含用户基本信息
func (s *OAuth2Service) GetUserInfo(accessToken string) (*UserInfo, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("访问令牌不能为空")
	}

	userInfoURL := s.oauth2Server + "/oauth2/userinfo"

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var oauthErr OAuth2Error
		if json.Unmarshal(body, &oauthErr) == nil && oauthErr.Code != "" {
			return nil, fmt.Errorf("OAuth2 错误: %s - %s", oauthErr.Code, oauthErr.ErrorDescription)
		}
		return nil, fmt.Errorf("获取用户信息失败，HTTP 状态码: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	return &userInfo, nil
}

// RefreshToken 使用刷新令牌获取新的访问令牌
//
// 参数 refreshToken 为刷新令牌
// 返回 TokenResponse 包含新的访问令牌和刷新令牌
func (s *OAuth2Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("刷新令牌不能为空")
	}

	tokenURL := s.oauth2Server + "/oauth2/token"

	formData := url.Values{}
	formData.Set("grant_type", "refresh_token")
	formData.Set("refresh_token", refreshToken)
	formData.Set("client_id", s.clientID)
	formData.Set("client_secret", s.clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var oauthErr OAuth2Error
		if json.Unmarshal(body, &oauthErr) == nil && oauthErr.Code != "" {
			return nil, fmt.Errorf("OAuth2 错误: %s - %s", oauthErr.Code, oauthErr.ErrorDescription)
		}
		return nil, fmt.Errorf("令牌刷新失败，HTTP 状态码: %d", resp.StatusCode)
	}

	var tokenResp TokenResponseBody
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("令牌响应中缺少 access_token")
	}

	return tokenResp.ToTokenResponse(), nil
}

// GetConfig 获取 OAuth2 公开配置
//
// 返回不含 client_secret 的配置信息，供前端使用
func (s *OAuth2Service) GetConfig() PublicConfig {
	return PublicConfig{
		Server:      s.oauth2Server,
		ClientID:    s.clientID,
		RedirectURI: s.redirectURI,
	}
}

// GetServer 返回 OAuth2 服务器地址
func (s *OAuth2Service) GetServer() string {
	return s.oauth2Server
}

// IntrospectToken 验证令牌（可选功能）
//
// 部分 OAuth2 服务器支持 introspection 端点
func (s *OAuth2Service) IntrospectToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("令牌不能为空")
	}

	introspectURL := s.oauth2Server + "/oauth2/introspect"

	formData := url.Values{}
	formData.Set("token", token)
	formData.Set("client_id", s.clientID)
	formData.Set("client_secret", s.clientSecret)

	req, err := http.NewRequest("POST", introspectURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return false, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("令牌验证失败，HTTP 状态码: %d", resp.StatusCode)
	}

	var result struct {
		Active bool `json:"active"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("解析响应失败: %w", err)
	}

	return result.Active, nil
}

// 构建授权 URL
//
// 用于生成 OAuth2 授权页面的 URL，供前端跳转使用
func (s *OAuth2Service) BuildAuthorizeURL(state string, scope string) string {
	authURL := s.oauth2Server + "/oauth2/authorize"

	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)

	if scope != "" {
		params.Set("scope", scope)
	}

	return authURL + "?" + params.Encode()
}
