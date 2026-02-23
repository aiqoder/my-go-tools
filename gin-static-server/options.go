package ginstatic

import (
	"compress/gzip"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// Config 静态文件服务器配置
type Config struct {
	// 基础配置
	Root   string // 静态文件根目录
	Prefix string // URL 前缀路径

	// Embed 配置
	EmbedFS   any    // 支持 embed.FS 或 http.FileSystem
	EmbedRoot string // embed.FS 的根目录（子目录）

	// 缓存配置
	EnableCache   bool  // 是否启用内存缓存
	MaxCacheSize  int64 // 最大缓存大小（字节），默认 100MB
	MaxCacheFiles int   // 最大缓存文件数，默认 500

	// 压缩配置
	EnableGzip      bool // 是否启用 Gzip 压缩，默认 true
	GzipLevel       int  // Gzip 压缩级别 (1-9)，默认 gzip.BestSpeed
	CompressMinSize int  // 最小压缩大小，默认 1024 字节

	// SPA 支持
	EnableSPA   bool   // 是否启用 SPA 回退，默认 false
	IndexFile   string // index.html 路径，默认 "index.html"
	SPAFallback bool   // 是否在文件不存在时回退到 index.html

	// 安全配置
	HideDotFiles bool // 是否隐藏点文件，默认 true

	// 缓存控制
	CacheControl string // 缓存控制头，默认 "public, max-age=60"
	UseETag      bool   // 是否使用 ETag，默认 true

	// 性能配置
	PreloadOnStart bool // 启动时预加载文件到缓存，默认 false
	ReadTimeout    int  // 读取超时（毫秒），默认 0 表示不限制
	WriteTimeout   int  // 写入超时（毫秒），默认 0 表示不限制

	// 自定义
	Custom404    string            // 自定义 404 页面路径
	MimeTypes    map[string]string // 自定义 MIME 类型
	OnCacheEvict func(string)      // 缓存淘汰回调
	OnRequest    func(string) bool // 请求前回调，返回 false 拒绝请求
}

// Option 配置选项函数类型
type Option func(*Config)

// defaultConfig 返回默认配置
func defaultConfig(root string) *Config {
	return &Config{
		Root:            root,
		Prefix:          "",
		EnableCache:     true,
		MaxCacheSize:    100 * 1024 * 1024, // 100MB
		MaxCacheFiles:   500,
		EnableGzip:      true,
		GzipLevel:       gzip.BestSpeed,
		CompressMinSize: 1024,
		EnableSPA:       false,
		IndexFile:       "index.html",
		SPAFallback:     false,
		HideDotFiles:    true,
		CacheControl:    "public, max-age=60",
		UseETag:         true,
		PreloadOnStart:  false,
		ReadTimeout:     0,
		WriteTimeout:    0,
		Custom404:       "",
		MimeTypes:       make(map[string]string),
		OnCacheEvict:    nil,
		OnRequest:       nil,
	}
}

// WithPrefix 设置 URL 前缀
// 例如: WithPrefix("/static") 将使 /static/js/app.js -> ./public/js/app.js
func WithPrefix(prefix string) Option {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// WithCache 启用内存缓存
// maxSize: 最大缓存大小（字节）
// maxFiles: 最大缓存文件数
func WithCache(maxSize int64, maxFiles int) Option {
	return func(c *Config) {
		c.EnableCache = true
		c.MaxCacheSize = maxSize
		c.MaxCacheFiles = maxFiles
	}
}

// DisableCache 禁用内存缓存
func DisableCache() Option {
	return func(c *Config) {
		c.EnableCache = false
	}
}

// WithGzip 启用 Gzip 压缩
// level: 压缩级别 (1-9)，1 最快，9 最高压缩比
func WithGzip(level int) Option {
	if level < 1 {
		level = 1
	}
	if level > 9 {
		level = 9
	}
	return func(c *Config) {
		c.EnableGzip = true
		c.GzipLevel = level
	}
}

// DisableGzip 禁用 Gzip 压缩
func DisableGzip() Option {
	return func(c *Config) {
		c.EnableGzip = false
	}
}

// WithCompressMinSize 设置最小压缩大小
// 小于此大小的文件将不会被压缩
func WithCompressMinSize(size int) Option {
	return func(c *Config) {
		c.CompressMinSize = size
	}
}

// WithSPA 启用 SPA 回退支持
// 当请求的文件不存在时，返回 index.html 由前端路由接管
func WithSPA(indexFile string) Option {
	return func(c *Config) {
		c.EnableSPA = true
		c.IndexFile = indexFile
		c.SPAFallback = true
	}
}

// DisableSPA 禁用 SPA 回退
func DisableSPA() Option {
	return func(c *Config) {
		c.EnableSPA = false
		c.SPAFallback = false
	}
}

// WithCacheControl 设置缓存控制头
func WithCacheControl(header string) Option {
	return func(c *Config) {
		c.CacheControl = header
	}
}

// WithoutETag 禁用 ETag
func WithoutETag() Option {
	return func(c *Config) {
		c.UseETag = false
	}
}

// WithETag 启用 ETag
func WithETag() Option {
	return func(c *Config) {
		c.UseETag = true
	}
}

// HideDotFiles 隐藏点文件（默认行为）
func HideDotFiles() Option {
	return func(c *Config) {
		c.HideDotFiles = true
	}
}

// ShowDotFiles 显示点文件
func ShowDotFiles() Option {
	return func(c *Config) {
		c.HideDotFiles = false
	}
}

// WithPreloadOnStart 启动时预加载文件到缓存
func WithPreloadOnStart() Option {
	return func(c *Config) {
		c.PreloadOnStart = true
	}
}

// WithReadTimeout 设置读取超时
func WithReadTimeout(timeoutMs int) Option {
	return func(c *Config) {
		c.ReadTimeout = timeoutMs
	}
}

// WithWriteTimeout 设置写入超时
func WithWriteTimeout(timeoutMs int) Option {
	return func(c *Config) {
		c.WriteTimeout = timeoutMs
	}
}

// WithCustom404 设置自定义 404 页面
func WithCustom404(path string) Option {
	return func(c *Config) {
		c.Custom404 = path
	}
}

// WithMimeTypes 设置自定义 MIME 类型
func WithMimeTypes(types map[string]string) Option {
	return func(c *Config) {
		for k, v := range types {
			c.MimeTypes[k] = v
		}
	}
}

// WithOnCacheEvict 设置缓存淘汰回调
func WithOnCacheEvict(fn func(string)) Option {
	return func(c *Config) {
		c.OnCacheEvict = fn
	}
}

// WithOnRequest 设置请求前回调
func WithOnRequest(fn func(string) bool) Option {
	return func(c *Config) {
		c.OnRequest = fn
	}
}

// WithEmbedRoot 设置 embed.FS 的根目录
// 当 embed 的是子目录时使用，例如：
//
//	//go:embed dist
//	var assets embed.FS
//
//	ginstatic.NewEmbed(r, assets, ginstatic.WithEmbedRoot("dist"))
func WithEmbedRoot(root string) Option {
	return func(c *Config) {
		c.EmbedRoot = root
	}
}

// WithIndexFile 设置 index 文件路径
func WithIndexFile(path string) Option {
	return func(c *Config) {
		c.IndexFile = path
	}
}

// applyOptions 应用配置选项
func applyConfig(root string, opts []Option) *Config {
	cfg := defaultConfig(root)
	for _, opt := range opts {
		opt(cfg)
	}
	// 确保 Root 是绝对路径或正确的相对路径
	if cfg.Root != "" {
		cfg.Root = filepath.Clean(cfg.Root)
	}
	return cfg
}

// GetMimeType 获取文件的 MIME 类型
func GetMimeType(filename string, customTypes map[string]string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "application/octet-stream"
	}
	ext = ext[1:] // 去掉前导点

	// 检查自定义类型
	if mime, ok := customTypes[ext]; ok {
		return mime
	}

	// 内置常见类型
	mimeTypes := map[string]string{
		"html":  "text/html; charset=utf-8",
		"htm":   "text/html; charset=utf-8",
		"css":   "text/css; charset=utf-8",
		"js":    "application/javascript; charset=utf-8",
		"mjs":   "application/javascript; charset=utf-8",
		"json":  "application/json; charset=utf-8",
		"xml":   "application/xml; charset=utf-8",
		"txt":   "text/plain; charset=utf-8",
		"md":    "text/markdown; charset=utf-8",
		"png":   "image/png",
		"jpg":   "image/jpeg",
		"jpeg":  "image/jpeg",
		"gif":   "image/gif",
		"svg":   "image/svg+xml",
		"ico":   "image/x-icon",
		"webp":  "image/webp",
		"avif":  "image/avif",
		"woff":  "font/woff",
		"woff2": "font/woff2",
		"ttf":   "font/ttf",
		"eot":   "application/vnd.ms-fontobject",
		"otf":   "font/otf",
		"pdf":   "application/pdf",
		"zip":   "application/zip",
		"gz":    "application/gzip",
		"tar":   "application/x-tar",
		"wasm":  "application/wasm",
		"map":   "application/json", // source map
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}

// parseTimeDuration 解析时间字符串
func parseTimeDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// NewEmbed 使用 embed.FS 创建静态文件服务器
// embedFS: 使用 go:embed 嵌入的文件系统
// opts: 配置选项
//
// 示例:
//
//	//go:embed "dist"
//	var assets embed.FS
//
//	func main() {
//	    r := gin.Default()
//	    ginstatic.NewEmbed(r, assets)
//	}
func NewEmbed(router *gin.Engine, embedFS any, opts ...Option) *StaticEngine {
	cfg := &Config{
		Root:            "",
		Prefix:          "",
		EmbedFS:         embedFS,
		EnableCache:     true,
		MaxCacheSize:    100 * 1024 * 1024,
		MaxCacheFiles:   500,
		EnableGzip:      true,
		GzipLevel:       gzip.BestSpeed,
		CompressMinSize: 1024,
		EnableSPA:       false,
		IndexFile:       "index.html",
		SPAFallback:     false,
		HideDotFiles:    true,
		CacheControl:    "public, max-age=60",
		UseETag:         true,
		PreloadOnStart:  false,
		ReadTimeout:     0,
		WriteTimeout:    0,
		Custom404:       "",
		MimeTypes:       make(map[string]string),
		OnCacheEvict:    nil,
		OnRequest:       nil,
	}

	// 应用选项
	for _, opt := range opts {
		opt(cfg)
	}

	engine := &StaticEngine{
		config: cfg,
		cache:  NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict),
	}

	// 注册路由
	engine.registerRoutes(router)

	// 预加载（如启用）
	if cfg.PreloadOnStart {
		go func() {
			engine.preloadEmbed()
		}()
	}

	return engine
}

// NewEmbedWithConfig 使用配置和 embed.FS 创建静态文件服务器
func NewEmbedWithConfig(router *gin.Engine, embedFS any, cfg *Config) *StaticEngine {
	cfg.EmbedFS = embedFS
	engine := &StaticEngine{
		config: cfg,
		cache:  NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict),
	}

	engine.registerRoutes(router)

	if cfg.PreloadOnStart {
		go func() {
			engine.preloadEmbed()
		}()
	}

	return engine
}
