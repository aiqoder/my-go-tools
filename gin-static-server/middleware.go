package ginstatic

import (
	"fmt"
	"hash/fnv"
	"io"
	fs2 "io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// ============================================================
// 静态文件中间件 - 解决路由冲突问题
// ============================================================

// StaticExtsMiddlewareConfig 静态文件扩展名中间件配置
type StaticExtsMiddlewareConfig struct {
	Root            string            // 静态文件根目录
	Prefix          string            // URL 路径前缀
	EmbedFS         any               // embed.FS
	EmbedRoot       string            // embed.FS 根目录
	StaticExts      []string          // 需要拦截的静态资源扩展名
	EnableCache     bool              // 是否启用缓存
	MaxCacheSize    int64             // 最大缓存大小
	MaxCacheFiles   int               // 最大缓存文件数
	EnableGzip      bool              // 是否启用 Gzip
	GzipLevel       int               // Gzip 压缩级别
	UseETag         bool              // 是否使用 ETag
	CacheControl    string            // 缓存控制头
	HideDotFiles    bool              // 是否隐藏点文件
	EnableIndex     bool              // 是否启用 index.html 回退（访问 / 自动返回 index.html）
	IndexFile       string            // 默认索引文件，默认 "index.html"
	OnCacheEvict    func(string)      // 缓存淘汰回调
	OnRequest       func(string) bool // 请求前回调
}

// MiddlewareOption 中间件配置选项函数
type MiddlewareOption func(*StaticExtsMiddlewareConfig)

// 默认静态资源扩展名列表
var defaultStaticExts = []string{
	".html", ".htm",
	".js", ".mjs",
	".css", ".scss", ".less",
	".json",
	".map",
	".woff", ".woff2",
	".ttf", ".eot", ".otf",
	".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp", ".avif",
	".wasm",
	".txt", ".md",
	".xml",
	".pdf", ".zip", ".gz", ".tar",
}

// defaultStaticExtsMiddlewareConfig 返回默认中间件配置
func defaultStaticExtsMiddlewareConfig(root string) *StaticExtsMiddlewareConfig {
	return &StaticExtsMiddlewareConfig{
		Root:           root,
		Prefix:         "",
		StaticExts:     defaultStaticExts,
		EnableCache:    true,
		MaxCacheSize:   100 * 1024 * 1024, // 100MB
		MaxCacheFiles:  500,
		EnableGzip:     true,
		GzipLevel:      1, // BestSpeed
		UseETag:        true,
		CacheControl:   "public, max-age=60",
		HideDotFiles:   true,
		EnableIndex:    true,
		IndexFile:      "index.html",
	}
}

// WithMiddlewarePrefix 设置 URL 前缀
func WithMiddlewarePrefix(prefix string) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.Prefix = strings.TrimSuffix(prefix, "/")
	}
}

// WithMiddlewareStaticExts 自定义静态资源扩展名
func WithMiddlewareStaticExts(exts []string) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.StaticExts = exts
	}
}

// WithMiddlewareEmbedFS 使用 embed.FS
func WithMiddlewareEmbedFS(fs any, root string) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EmbedFS = fs
		c.EmbedRoot = root
	}
}

// WithMiddlewareCache 启用缓存
func WithMiddlewareCache(maxSize int64, maxFiles int) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableCache = true
		c.MaxCacheSize = maxSize
		c.MaxCacheFiles = maxFiles
	}
}

// DisableMiddlewareCache 禁用缓存
func DisableMiddlewareCache() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableCache = false
	}
}

// WithMiddlewareGzip 启用 Gzip
func WithMiddlewareGzip(level int) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableGzip = true
		if level < 1 {
			level = 1
		}
		if level > 9 {
			level = 9
		}
		c.GzipLevel = level
	}
}

// DisableMiddlewareGzip 禁用 Gzip
func DisableMiddlewareGzip() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableGzip = false
	}
}

// WithMiddlewareETag 启用 ETag
func WithMiddlewareETag() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.UseETag = true
	}
}

// WithoutMiddlewareETag 禁用 ETag
func WithoutMiddlewareETag() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.UseETag = false
	}
}

// WithMiddlewareCacheControl 设置缓存控制头
func WithMiddlewareCacheControl(control string) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.CacheControl = control
	}
}

// WithMiddlewareHideDotFiles 隐藏点文件
func WithMiddlewareHideDotFiles() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.HideDotFiles = true
	}
}

// WithMiddlewareShowDotFiles 显示点文件
func WithMiddlewareShowDotFiles() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.HideDotFiles = false
	}
}

// WithMiddlewareIndex 启用 index.html 回退（访问 / 自动返回 index.html）
func WithMiddlewareIndex() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableIndex = true
	}
}

// DisableMiddlewareIndex 禁用 index.html 回退
func DisableMiddlewareIndex() MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.EnableIndex = false
	}
}

// WithMiddlewareIndexFile 设置默认索引文件
func WithMiddlewareIndexFile(filename string) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.IndexFile = filename
	}
}

// WithMiddlewareOnCacheEvict 设置缓存淘汰回调
func WithMiddlewareOnCacheEvict(fn func(string)) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.OnCacheEvict = fn
	}
}

// WithMiddlewareOnRequest 设置请求前回调
func WithMiddlewareOnRequest(fn func(string) bool) MiddlewareOption {
	return func(c *StaticExtsMiddlewareConfig) {
		c.OnRequest = fn
	}
}

// isStaticFile 检查路径是否匹配静态资源扩展名
func isStaticFile(path string, exts []string) bool {
	path = strings.ToLower(path)
	for _, ext := range exts {
		ext = strings.ToLower(ext)
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// StaticFileExtsMiddleware 创建静态文件扩展名中间件
// 该中间件只处理以特定扩展名结尾的请求，解决与 API 路由冲突问题
// 当文件不存在时，调用 c.Next() 继续传递请求
func StaticFileExtsMiddleware(root string, opts ...MiddlewareOption) gin.HandlerFunc {
	cfg := defaultStaticExtsMiddlewareConfig(root)
	for _, opt := range opts {
		opt(cfg)
	}

	// 确保 Root 是绝对路径
	if cfg.Root != "" {
		cfg.Root = filepath.Clean(cfg.Root)
	}

	// 创建缓存（如启用）
	var cache *Cache
	if cfg.EnableCache {
		cache = NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查是否为根路径，启用 index.html 回退
		if cfg.EnableIndex && (path == "/" || path == "") {
			path = "/" + cfg.IndexFile
		}

		// 检查路径是否匹配静态资源扩展名
		if !isStaticFile(path, cfg.StaticExts) {
			c.Next()
			return
		}

		// 移除前缀
		path = strings.TrimPrefix(path, cfg.Prefix)

		// 移除前导斜杠
		path = strings.TrimPrefix(path, "/")

		// 安全检查：目录遍历防护
		safe, cleanPath := IsPathTraversal(cfg.Root, path)
		if !safe {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 检查隐藏文件
		if cfg.HideDotFiles && IsHiddenPath(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 请求前回调
		if cfg.OnRequest != nil && !cfg.OnRequest(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 获取文件
		data, modTime, etag, err := getMiddlewareFile(cfg, cache, cleanPath)
		if err != nil {
			// 文件不存在，调用 Next 继续传递
			c.Next()
			return
		}

		// 条件请求检查
		if checkMiddlewareNotModified(c, modTime, etag, cfg.UseETag) {
			c.Abort()
			return
		}

		// 设置响应头
		mimeType := GetMimeType(cleanPath, nil)
		c.Header("Content-Type", mimeType)
		c.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

		if cfg.UseETag {
			c.Header("ETag", etag)
		}

		if cfg.CacheControl != "" {
			c.Header("Cache-Control", cfg.CacheControl)
		}

		// Gzip 压缩
		data = applyMiddlewareGzip(c, data, cfg)

		c.Status(http.StatusOK)
		c.Header("Content-Type", mimeType)
		c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
		c.Writer.Write(data)
		c.Abort()
	}
}

// getMiddlewareFile 获取文件内容（支持缓存）
func getMiddlewareFile(cfg *StaticExtsMiddlewareConfig, cache *Cache, path string) ([]byte, time.Time, string, error) {
	// 尝试从缓存获取
	if cfg.EnableCache && cache != nil {
		if entry, ok := cache.Get(path); ok {
			return entry.Data, entry.ModTime, entry.ETag, nil
		}
	}

	var data []byte
	var modTime time.Time
	var etag string
	var err error

	// 判断使用文件系统还是 embed
	if cfg.EmbedFS != nil {
		data, modTime, etag, err = getMiddlewareEmbedFile(cfg, path)
	} else {
		data, modTime, etag, err = getMiddlewareOSFile(cfg, path)
	}

	if err != nil {
		return nil, time.Time{}, "", err
	}

	// 缓存（如启用）
	if cfg.EnableCache && cache != nil {
		entry := &cacheEntry{
			Data:    data,
			ModTime: modTime,
			Size:    int64(len(data)),
			ETag:    etag,
		}

		// 压缩（如启用）
		if cfg.EnableGzip && len(data) >= 1024 {
			gzData, gzErr := GzipCompress(data, cfg.GzipLevel)
			if gzErr == nil {
				entry.Gzipped = gzData
			}
		}

		cache.Set(path, entry)
	}

	return data, modTime, etag, nil
}

// getMiddlewareOSFile 从文件系统读取文件
func getMiddlewareOSFile(cfg *StaticExtsMiddlewareConfig, path string) ([]byte, time.Time, string, error) {
	absPath := filepath.Join(cfg.Root, path)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, time.Time{}, "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, time.Time{}, "", err
	}

	etag := generateMiddlewareETag(absPath, info)
	return data, info.ModTime(), etag, nil
}

// getMiddlewareEmbedFile 从 embed.FS 读取文件
func getMiddlewareEmbedFile(cfg *StaticExtsMiddlewareConfig, path string) ([]byte, time.Time, string, error) {
	// 移除前导斜杠
	path = strings.TrimPrefix(path, "/")

	// 加上 EmbedRoot 前缀
	if cfg.EmbedRoot != "" {
		path = cfg.EmbedRoot + "/" + path
	}

	var data []byte
	var info os.FileInfo
	var err error

	// 尝试不同的 FS 类型
	type fsOpener interface {
		Open(name string) (fs2.File, error)
	}

	if f, ok := cfg.EmbedFS.(fsOpener); ok {
		file, openErr := f.Open(path)
		if openErr != nil {
			return nil, time.Time{}, "", openErr
		}
		defer file.Close()

		info, err = file.Stat()
		if err != nil {
			return nil, time.Time{}, "", err
		}

		data, err = io.ReadAll(file)
		if err != nil {
			return nil, time.Time{}, "", err
		}
	} else if httpFS, ok := cfg.EmbedFS.(http.FileSystem); ok {
		file, openErr := httpFS.Open(path)
		if openErr != nil {
			return nil, time.Time{}, "", openErr
		}
		defer file.Close()

		info, err = file.Stat()
		if err != nil {
			return nil, time.Time{}, "", err
		}

		data, err = io.ReadAll(file)
		if err != nil {
			return nil, time.Time{}, "", err
		}
	} else {
		return nil, time.Time{}, "", fmt.Errorf("unsupported file system type")
	}

	if info.IsDir() {
		return nil, time.Time{}, "", os.ErrNotExist
	}

	etag := generateMiddlewareETagEmbed(path, info)
	return data, info.ModTime(), etag, nil
}

// generateMiddlewareETag 为文件生成 ETag
func generateMiddlewareETag(path string, info os.FileInfo) string {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", info.Size(), info.ModTime().UnixNano(), path)))
	return fmt.Sprintf(`"%x"`, h.Sum(nil))
}

// generateMiddlewareETagEmbed 为 embed 文件生成 ETag
func generateMiddlewareETagEmbed(path string, info os.FileInfo) string {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", info.Size(), info.ModTime().UnixNano(), path)))
	return fmt.Sprintf(`"%x"`, h.Sum(nil))
}

// checkMiddlewareNotModified 检查条件请求
func checkMiddlewareNotModified(c *gin.Context, modTime time.Time, etag string, useETag bool) bool {
	if useETag {
		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			c.Status(http.StatusNotModified)
			return true
		}
	}

	ifModifiedSince := c.GetHeader("If-Modified-Since")
	if ifModifiedSince != "" {
		ifMod, err := time.Parse(http.TimeFormat, ifModifiedSince)
		if err == nil && !modTime.After(ifMod) {
			c.Status(http.StatusNotModified)
			return true
		}
	}

	return false
}

// applyMiddlewareGzip 应用 Gzip 压缩
func applyMiddlewareGzip(c *gin.Context, data []byte, cfg *StaticExtsMiddlewareConfig) []byte {
	if !cfg.EnableGzip || len(data) < 1024 {
		return data
	}

	acceptEncoding := c.GetHeader("Accept-Encoding")
	if !containsEncoding(acceptEncoding, "gzip") {
		return data
	}

	// 尝试从缓存获取压缩数据
	if cfg.EnableCache {
		// 简化处理，直接压缩
	}

	gzData, err := GzipCompress(data, cfg.GzipLevel)
	if err == nil && len(gzData) < len(data) {
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		return gzData
	}

	return data
}

// NewStaticFileExtsMiddlewareWithConfig 使用配置创建中间件
func NewStaticFileExtsMiddlewareWithConfig(cfg *StaticExtsMiddlewareConfig) gin.HandlerFunc {
	// 确保 Root 是绝对路径
	if cfg.Root != "" {
		cfg.Root = filepath.Clean(cfg.Root)
	}

	// 如果没有设置扩展名，使用默认值
	if len(cfg.StaticExts) == 0 {
		cfg.StaticExts = defaultStaticExts
	}

	// 创建缓存（如启用）
	var cache *Cache
	if cfg.EnableCache {
		cache = NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict)
	}

	// 如果没有设置 EnableIndex，默认启用
	if !cfg.EnableIndex {
		cfg.EnableIndex = true
	}
	if cfg.IndexFile == "" {
		cfg.IndexFile = "index.html"
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查是否为根路径，启用 index.html 回退
		if cfg.EnableIndex && (path == "/" || path == "") {
			path = "/" + cfg.IndexFile
		}

		// 检查路径是否匹配静态资源扩展名
		if !isStaticFile(path, cfg.StaticExts) {
			c.Next()
			return
		}

		// 移除前缀
		path = strings.TrimPrefix(path, cfg.Prefix)

		// 移除前导斜杠
		path = strings.TrimPrefix(path, "/")

		// 安全检查
		safe, cleanPath := IsPathTraversal(cfg.Root, path)
		if !safe {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 检查隐藏文件
		if cfg.HideDotFiles && IsHiddenPath(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 请求前回调
		if cfg.OnRequest != nil && !cfg.OnRequest(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// 获取文件
		data, modTime, etag, err := getMiddlewareFile(cfg, cache, cleanPath)
		if err != nil {
			c.Next()
			return
		}

		// 条件请求检查
		if checkMiddlewareNotModified(c, modTime, etag, cfg.UseETag) {
			return
		}

		// 设置响应头
		mimeType := GetMimeType(cleanPath, nil)
		c.Header("Content-Type", mimeType)
		c.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

		if cfg.UseETag {
			c.Header("ETag", etag)
		}

		if cfg.CacheControl != "" {
			c.Header("Cache-Control", cfg.CacheControl)
		}

		// Gzip 压缩
		data = applyMiddlewareGzip(c, data, cfg)

		c.Status(http.StatusOK)
		c.Header("Content-Type", mimeType)
		c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
		c.Writer.Write(data)
		c.Abort()
	}
}
