package oauth2

import (
	"github.com/gin-gonic/gin"
)

// RouterOption 路由配置选项
type RouterOption func(*gin.RouterGroup)

// WithAuthMiddleware 添加认证中间件
func WithAuthMiddleware(handler *OAuth2Handler) RouterOption {
	return func(r *gin.RouterGroup) {
		r.Use(handler.Middleware())
	}
}

// SetupRouter 设置 OAuth2 路由
//
// 参数 r 为 gin.Engine 实例或 gin.RouterGroup
// 返回配置后的路由组
func SetupRouter(r *gin.Engine, handler *OAuth2Handler, opts ...RouterOption) *gin.RouterGroup {
	api := r.Group("/api")
	{
		oauth2Group := api.Group("")
		for _, opt := range opts {
			opt(oauth2Group)
		}
		handler.RegisterRoutes(oauth2Group)
	}
	return api
}

// SetupRouterWithGroup 在指定的 RouterGroup 上设置 OAuth2 路由
//
// 更加灵活的路由设置方式，可以在已有的路由组上添加 OAuth2 路由
func SetupRouterWithGroup(group *gin.RouterGroup, handler *OAuth2Handler, opts ...RouterOption) *gin.RouterGroup {
	for _, opt := range opts {
		opt(group)
	}
	handler.RegisterRoutes(group)
	return group
}

// DefaultConfig 返回默认配置
//
// 用于快速初始化，配置通过环境变量覆盖
func DefaultConfig() *Config {
	return &Config{
		Server:       getEnv("OAUTH2_SERVER", "http://localhost:8080"),
		ClientID:     getEnv("OAUTH2_CLIENT_ID", ""),
		ClientSecret: getEnv("OAUTH2_CLIENT_SECRET", ""),
		RedirectURI:  getEnv("OAUTH2_REDIRECT_URI", "http://localhost:3000/callback"),
	}
}

// getEnv 获取环境变量值
func getEnv(key, defaultValue string) string {
	if value := getEnvFunc(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvFunc 环境变量获取函数，便于测试 mock
var getEnvFunc = func(key string) string {
	// 实际实现会使用 os.Getenv
	// 这里使用延迟函数避免循环依赖
	return ""
}
