package ginstatic

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MiddlewareOptions 中间件配置选项
type MiddlewareOptions struct {
	// 请求限流
	EnableRateLimit bool
	RateLimit       int           // 每秒请求数
	Burst           int           // 突发限制

	// CORS
	EnableCORS      bool
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
	ExposeHeaders   []string
	MaxAge          time.Duration

	// 安全头
	EnableSecureHeaders bool
	ContentTypeNosniff  bool
	XSSProtection       string
	FrameOptions        string

	// 日志
	EnableLogging bool
	LogFormat     string
}

// DefaultMiddlewareOptions 返回默认中间件选项
func DefaultMiddlewareOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		EnableRateLimit:     false,
		RateLimit:           100,
		Burst:               200,
		EnableCORS:          false,
		AllowedOrigins:      []string{"*"},
		AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:      []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:       []string{"Content-Length", "Content-Type"},
		MaxAge:              12 * time.Hour,
		EnableSecureHeaders: true,
		ContentTypeNosniff:  true,
		XSSProtection:       "1; mode=block",
		FrameOptions:        "SAMEORIGIN",
		EnableLogging:       false,
		LogFormat:           "",
	}
}

// RateLimitMiddleware 创建限流中间件
// 注意：这是一个简单的实现，生产环境建议使用更完善的限流方案
func RateLimitMiddleware(requestsPerSecond int, burst int) gin.HandlerFunc {
	// 简化实现：使用令牌桶算法
	type tokenBucket struct {
		tokens     int
		lastUpdate time.Time
	}

	var bucket tokenBucket
	bucket.tokens = burst
	bucket.lastUpdate = time.Now()

	ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))

	return func(c *gin.Context) {
		select {
		case <-ticker.C:
			if bucket.tokens < burst {
				bucket.tokens++
			}
		default:
		}

		if bucket.tokens > 0 {
			bucket.tokens--
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
		}
	}
}

// CORSMiddleware 创建 CORS 中间件
func CORSMiddleware(opts *MiddlewareOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查是否允许该来源
		allowed := false
		for _, o := range opts.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if len(opts.AllowedOrigins) == 1 && opts.AllowedOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Methods", strings.Join(opts.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(opts.AllowedHeaders, ", "))
			c.Header("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", int(opts.MaxAge.Seconds())))
		}

		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecureHeadersMiddleware 创建安全头中间件
func SecureHeadersMiddleware(opts *MiddlewareOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		if opts.ContentTypeNosniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		if opts.XSSProtection != "" {
			c.Header("X-XSS-Protection", opts.XSSProtection)
		}

		if opts.FrameOptions != "" {
			c.Header("X-Frame-Options", opts.FrameOptions)
		}

		// 常见安全头
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// RequestLoggerMiddleware 创建请求日志中间件
func RequestLoggerMiddleware(format string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if format == "" {
			format = "[GIN] %s | %3d | %13v | %15s | %-7s %s"
		}

		// 简单日志输出
		if status >= 400 {
			log.Printf(format,
				c.Request.Method,
				status,
				latency,
				c.ClientIP(),
				c.Request.UserAgent(),
				path,
			)
		}
	}
}

// CacheMiddleware 创建缓存控制中间件
func CacheMiddleware(maxAge time.Duration, immutable bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := "public"
		if maxAge > 0 {
			value += ", max-age=" + fmt.Sprintf("%d", int(maxAge.Seconds()))
		}
		if immutable {
			value += ", immutable"
		}
		c.Header("Cache-Control", value)
		c.Next()
	}
}

// NoCacheMiddleware 创建禁用缓存中间件
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// CompressMiddleware 创建压缩检测中间件
// 注：主要压缩逻辑在 handler 中实现，此处用于检测客户端支持
func CompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		c.Set("Accept-Encoding", acceptEncoding)
		c.Next()
	}
}

// WithMiddleware 添加中间件到静态服务器
func (e *StaticEngine) WithMiddleware(opts *MiddlewareOptions) *StaticEngine {
	// 这个方法可以用于后续扩展
	// 当前主要功能在 New 函数中已经集成
	return e
}

// AddMiddleware 添加自定义中间件
func (e *StaticEngine) AddMiddleware(middleware gin.HandlerFunc) *StaticEngine {
	// 可以在后续添加
	_ = middleware
	return e
}
