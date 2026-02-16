package oauth2

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OAuth2Handler OAuth2 HTTP 处理器
//
// 处理 OAuth2 相关的 HTTP 请求
type OAuth2Handler struct {
	oauth2Service *OAuth2Service
}

// NewOAuth2Handler 创建 OAuth2 处理器实例
//
// 参数 oauth2Service 为 OAuth2 服务层实例
func NewOAuth2Handler(oauth2Service *OAuth2Service) *OAuth2Handler {
	return &OAuth2Handler{
		oauth2Service: oauth2Service,
	}
}

// GetConfig 获取 OAuth2 配置
//
// GET /api/oauth2/config
// 返回不含 client_secret 的配置信息，供前端使用
func (h *OAuth2Handler) GetConfig(c *gin.Context) {
	config := h.oauth2Service.GetConfig()
	c.JSON(http.StatusOK, config)
}

// Callback 处理 OAuth2 授权码回调
//
// POST /api/oauth2/callback
// 接收前端传来的授权码，使用授权码向 OAuth2 服务器换取令牌
func (h *OAuth2Handler) Callback(c *gin.Context) {
	var req CallbackRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "缺少授权码: " + err.Error(),
		})
		return
	}

	tokenResp, err := h.oauth2Service.ExchangeCodeForToken(req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "token_exchange_failed",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// GetUserInfo 获取用户信息
//
// GET /api/oauth2/userinfo
// 从 Authorization Header 中提取 Bearer Token，向 OAuth2 服务器获取用户信息
func (h *OAuth2Handler) GetUserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	token := extractBearerToken(authHeader)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "unauthorized",
			"error_description": "缺少访问令牌",
		})
		return
	}

	userInfo, err := h.oauth2Service.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// RefreshToken 刷新访问令牌
//
// POST /api/oauth2/refresh
// 接收前端传来的刷新令牌，向 OAuth2 服务器换取新的访问令牌
func (h *OAuth2Handler) RefreshToken(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "缺少刷新令牌: " + err.Error(),
		})
		return
	}

	tokenResp, err := h.oauth2Service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "token_refresh_failed",
			"error_description": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// BuildAuthorizeURL 构建授权 URL
//
// GET /api/oauth2/authorize
// 返回 OAuth2 授权页面 URL，供前端跳转使用
func (h *OAuth2Handler) BuildAuthorizeURL(c *gin.Context) {
	state := c.Query("state")
	scope := c.DefaultQuery("scope", "read")

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "缺少 state 参数",
		})
		return
	}

	authURL := h.oauth2Service.BuildAuthorizeURL(state, scope)
	c.JSON(http.StatusOK, gin.H{
		"authorize_url": authURL,
	})
}

// extractBearerToken 从 Authorization Header 中提取 Bearer Token
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return authHeader
}

// RegisterRoutes 注册 OAuth2 路由
//
// 在 Gin 引擎中注册 OAuth2 相关的路由
func (h *OAuth2Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/oauth2/config", h.GetConfig)
	r.GET("/oauth2/authorize", h.BuildAuthorizeURL)
	r.POST("/oauth2/callback", h.Callback)
	r.GET("/oauth2/userinfo", h.GetUserInfo)
	r.POST("/oauth2/refresh", h.RefreshToken)
}

// Middleware 认证中间件
//
// 可选的认证中间件，用于验证请求中的访问令牌
func (h *OAuth2Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		token := extractBearerToken(authHeader)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":             "unauthorized",
				"error_description": "缺少访问令牌",
			})
			return
		}

		// 验证令牌
		_, err := h.oauth2Service.GetUserInfo(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":             "invalid_token",
				"error_description": err.Error(),
			})
			return
		}

		c.Next()
	}
}
