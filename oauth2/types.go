// Package oauth2 提供 OAuth2 授权码模式登录的后端实现
//
// 该模块完全依赖外部 OAuth2 服务器颁发的访问令牌进行身份验证，
// 不使用 JWT。包含配置管理、授权码换令牌、用户信息获取、令牌刷新等功能。
package oauth2

import "time"

// TokenResponse OAuth2 令牌响应
//
// 包含访问令牌、刷新令牌及相关过期信息
type TokenResponse struct {
	AccessToken      string `json:"access_token"`       // 访问令牌
	TokenType        string `json:"token_type"`         // 令牌类型（通常为 Bearer）
	ExpiresIn        int64  `json:"expires_in"`         // 访问令牌有效期（秒）
	RefreshToken     string `json:"refresh_token"`      // 刷新令牌
	RefreshExpiresIn int64  `json:"refresh_expires_in"` // 刷新令牌有效期（秒）
	Scope            string `json:"scope,omitempty"`    // 权限范围
}

// ExpiresAt 返回访问令牌的过期时间
//
// 返回 time.Time 类型的过期时间，便于前端计算倒计时
func (t *TokenResponse) ExpiresAt() time.Time {
	return time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
}

// RefreshExpiresAt 返回刷新令牌的过期时间
func (t *TokenResponse) RefreshExpiresAt() time.Time {
	return time.Now().Add(time.Duration(t.RefreshExpiresIn) * time.Second)
}

// UserInfo 用户信息
//
// 从 OAuth2 服务器获取的用户基本信息
type UserInfo struct {
	Sub       string  `json:"sub"`        // 用户唯一标识
	Username  string  `json:"username"`   // 用户名
	Status    int     `json:"status"`     // 用户状态
	ClientID  string  `json:"client_id"` // 客户端ID
	ExpireAt  *string `json:"expireAt,omitempty"`  // 过期时间
	IsExpired *bool   `json:"isExpired,omitempty"` // 是否已过期
}

// OAuth2Error OAuth2 错误响应
//
// 统一的错误响应格式，符合 OAuth2 RFC 规范
type OAuth2Error struct {
	Code            string `json:"error"`             // 错误码
	ErrorDescription string `json:"error_description"` // 错误描述
}

// Error 实现 error 接口
func (e *OAuth2Error) Error() string {
	if e.ErrorDescription != "" {
		return e.Code + ": " + e.ErrorDescription
	}
	return e.Code
}

// Config OAuth2 配置
//
// 用于在系统中传递 OAuth2 配置信息
type Config struct {
	Server       string // OAuth2 服务器地址
	ClientID     string // OAuth2 客户端 ID
	ClientSecret string // OAuth2 客户端密钥
	RedirectURI  string // OAuth2 重定向 URI
}

// PublicConfig 公开的 OAuth2 配置（不含密钥）
//
// 供前端使用的配置信息，不包含敏感信息
type PublicConfig struct {
	Server      string `json:"oauth_server"` // OAuth2 服务器地址
	ClientID    string `json:"client_id"`    // 客户端 ID
	RedirectURI string `json:"redirect_uri"` // 重定向 URI
}

// ToPublic 将完整配置转换为公开配置
//
// 返回不含 client_secret 的配置信息，供前端使用
func (c *Config) ToPublic() PublicConfig {
	return PublicConfig{
		Server:      c.Server,
		ClientID:    c.ClientID,
		RedirectURI: c.RedirectURI,
	}
}

// CallbackRequest 授权码回调请求
//
// 前端发送授权码的请求体
type CallbackRequest struct {
	Code string `json:"code" binding:"required"` // 授权码
}

// RefreshRequest 刷新令牌请求
//
// 前端发送刷新令牌的请求体
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

// TokenResponseBody OAuth2 token 端点返回的原始响应体
//
// 用于解析 OAuth2 服务器返回的令牌响应
type TokenResponseBody struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

// ToTokenResponse 转换为 TokenResponse
func (b *TokenResponseBody) ToTokenResponse() *TokenResponse {
	return &TokenResponse{
		AccessToken:      b.AccessToken,
		TokenType:        b.TokenType,
		ExpiresIn:        b.ExpiresIn,
		RefreshToken:     b.RefreshToken,
		RefreshExpiresIn: b.RefreshExpiresIn,
		Scope:            b.Scope,
	}
}
