package ginstatic

import (
	"fmt"
	"hash/fnv"
	"io"
	fs2 "io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// StaticEngine 静态文件服务引擎
type StaticEngine struct {
	config *Config
	cache  *Cache
}

// New 创建新的静态文件服务引擎
func New(router *gin.Engine, root string, opts ...Option) *StaticEngine {
	cfg := applyConfig(root, opts)
	engine := &StaticEngine{
		config: cfg,
		cache:  NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict),
	}

	// 注册路由
	engine.registerRoutes(router)

	// 预加载（如启用）
	if cfg.PreloadOnStart {
		go func() {
			engine.cache.preloadDirectory(cfg.Root, cfg.EnableGzip, cfg.GzipLevel)
		}()
	}

	return engine
}

// NewWithConfig 使用配置创建静态文件服务引擎
func NewWithConfig(router *gin.Engine, cfg *Config) *StaticEngine {
	engine := &StaticEngine{
		config: cfg,
		cache:  NewCache(cfg.MaxCacheSize, cfg.MaxCacheFiles, cfg.OnCacheEvict),
	}

	// 注册路由
	engine.registerRoutes(router)

	// 预加载（如启用）
	if cfg.PreloadOnStart {
		go func() {
			engine.cache.preloadDirectory(cfg.Root, cfg.EnableGzip, cfg.GzipLevel)
		}()
	}

	return engine
}

// registerRoutes 注册路由
func (e *StaticEngine) registerRoutes(router *gin.Engine) {
	prefix := strings.TrimSuffix(e.config.Prefix, "/")

	// 如果启用了 SPA 回退，使用自定义处理
	if e.config.EnableSPA && e.config.SPAFallback {
		router.GET(prefix+"/*path", e.serveSPA())
	} else {
		router.GET(prefix+"/*path", e.serveStatic())
	}
}

// serveStatic 服务静态文件
func (e *StaticEngine) serveStatic() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		if path == "" || path == "/" {
			path = "/" + e.config.IndexFile
		}

		// 移除前导斜杠
		path = strings.TrimPrefix(path, "/")

		// 安全检查
		safe, cleanPath := IsPathTraversal(e.config.Root, path)
		if !safe {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 检查隐藏文件
		if e.config.HideDotFiles && IsHiddenPath(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 请求前回调
		if e.config.OnRequest != nil && !e.config.OnRequest(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 获取文件
		data, modTime, etag, err := e.getFile(cleanPath)
		if err != nil {
			// 尝试 index.html（SPA 模式）
			if e.config.EnableSPA && e.config.SPAFallback {
				data, modTime, etag, err = e.getFile(e.config.IndexFile)
				if err != nil {
					e.serveError(c, http.StatusNotFound)
					return
				}
			} else {
				e.serveError(c, http.StatusNotFound)
				return
			}
		}

		// 条件请求检查
		if e.checkNotModified(c, modTime, etag) {
			return
		}

		// 设置响应头
		mimeType := GetMimeType(cleanPath, e.config.MimeTypes)
		c.Header("Content-Type", mimeType)
		c.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

		if e.config.UseETag {
			c.Header("ETag", etag)
		}

		if e.config.CacheControl != "" {
			c.Header("Cache-Control", e.config.CacheControl)
		}

		// Gzip 压缩
		data, encoding := e.getCompressedData(c, data, cleanPath)
		if encoding != "" {
			c.Header("Content-Encoding", encoding)
			c.Header("Vary", "Accept-Encoding")
		}

		c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
		c.Data(http.StatusOK, mimeType, data)
	}
}

// serveSPA 服务 SPA 应用
func (e *StaticEngine) serveSPA() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")

		// 移除前导斜杠，但保留空路径
		path = strings.TrimPrefix(path, "/")

		// 安全检查
		safe, cleanPath := IsPathTraversal(e.config.Root, path)
		if !safe {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 请求前回调
		if e.config.OnRequest != nil && !e.config.OnRequest(cleanPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 尝试获取文件
		data, modTime, etag, err := e.getFile(cleanPath)

		// 如果文件不存在或请求的是目录，返回 index.html
		if err != nil || cleanPath == "" {
			cleanPath = e.config.IndexFile
			data, modTime, etag, err = e.getFile(cleanPath)
			if err != nil {
				e.serveError(c, http.StatusNotFound)
				return
			}
		}

		// 条件请求检查
		if e.checkNotModified(c, modTime, etag) {
			return
		}

		// 设置响应头
		mimeType := GetMimeType(cleanPath, e.config.MimeTypes)
		c.Header("Content-Type", mimeType)
		c.Header("Last-Modified", modTime.UTC().Format(http.TimeFormat))

		if e.config.UseETag {
			c.Header("ETag", etag)
		}

		if e.config.CacheControl != "" {
			c.Header("Cache-Control", e.config.CacheControl)
		}

		// Gzip 压缩
		data, encoding := e.getCompressedData(c, data, cleanPath)
		if encoding != "" {
			c.Header("Content-Encoding", encoding)
			c.Header("Vary", "Accept-Encoding")
		}

		c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
		c.Data(http.StatusOK, mimeType, data)
	}
}

// getFile 获取文件内容
func (e *StaticEngine) getFile(path string) ([]byte, time.Time, string, error) {
	// 尝试从缓存获取
	if e.config.EnableCache {
		if entry, ok := e.cache.Get(path); ok {
			return entry.Data, entry.ModTime, entry.ETag, nil
		}
	}

	var data []byte
	var modTime time.Time
	var etag string
	var err error

	// 判断使用文件系统还是 embed
	if e.config.EmbedFS != nil {
		// 使用 embed.FS
		data, modTime, etag, err = e.getEmbedFile(path)
	} else {
		// 使用文件系统
		data, modTime, etag, err = e.getOSFile(path)
	}

	if err != nil {
		return nil, time.Time{}, "", err
	}

	// 缓存（如启用）
	if e.config.EnableCache {
		entry := &cacheEntry{
			Data:    data,
			ModTime: modTime,
			Size:    int64(len(data)),
			ETag:    etag,
		}

		// 压缩（如启用）
		if e.config.EnableGzip && len(data) >= e.config.CompressMinSize {
			gzData, err := GzipCompress(data, e.config.GzipLevel)
			if err == nil {
				entry.Gzipped = gzData
			}
		}

		e.cache.Set(path, entry)
	}

	return data, modTime, etag, nil
}

// getOSFile 从文件系统读取文件
func (e *StaticEngine) getOSFile(path string) ([]byte, time.Time, string, error) {
	absPath := filepath.Join(e.config.Root, path)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, time.Time{}, "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, time.Time{}, "", err
	}

	etag := generateETag(absPath, info)
	return data, info.ModTime(), etag, nil
}

// getEmbedFile 从 embed.FS 读取文件
func (e *StaticEngine) getEmbedFile(path string) ([]byte, time.Time, string, error) {
	// 移除前导斜杠
	path = strings.TrimPrefix(path, "/")

	// 加上 EmbedRoot 前缀
	if e.config.EmbedRoot != "" {
		path = e.config.EmbedRoot + "/" + path
	}

	var data []byte
	var info os.FileInfo
	var err error

	// 尝试直接使用 embed.FS (io/fs.FS)
	if embedFS, ok := e.config.EmbedFS.(fs2.FS); ok {
		data, info, err = readFromFS(embedFS, path)
	} else if httpFS, ok := e.config.EmbedFS.(http.FileSystem); ok {
		data, info, err = readFromHTTPFS(httpFS, path)
	} else {
		// 尝试使用 embed.FS 的 Open 方法
		type opener interface {
			Open(name string) (fs2.File, error)
		}
		f, ok := e.config.EmbedFS.(opener)
		if !ok {
			return nil, time.Time{}, "", fmt.Errorf("unsupported file system type")
		}
		data, info, err = readFromFSOpener(f, path)
	}

	if err != nil {
		return nil, time.Time{}, "", err
	}

	if info.IsDir() {
		return nil, time.Time{}, "", os.ErrNotExist
	}

	etag := generateETagEmbed(path, info)
	return data, info.ModTime(), etag, nil
}

// readFromFS 从 io/fs.FS 读取文件
func readFromFS(fsys fs2.FS, path string) ([]byte, os.FileInfo, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}

	return data, info, nil
}

// readFromHTTPFS 从 http.FileSystem 读取文件
func readFromHTTPFS(fsys http.FileSystem, path string) ([]byte, os.FileInfo, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}

	return data, info, nil
}

// readFromFSOpener 使用 fs.FileOpener 接口读取文件
func readFromFSOpener(opener interface {
	Open(name string) (fs2.File, error)
}, path string) ([]byte, os.FileInfo, error) {
	file, err := opener.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}

	return data, info, nil
}

// generateETagEmbed 为 embed 文件生成 ETag
func generateETagEmbed(path string, info os.FileInfo) string {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", info.Size(), info.ModTime().UnixNano(), path)))
	return fmt.Sprintf(`"%x"`, h.Sum(nil))
}

// getCompressedData 获取压缩后的数据
func (e *StaticEngine) getCompressedData(c *gin.Context, data []byte, path string) ([]byte, string) {
	if !e.config.EnableGzip {
		return data, ""
	}

	// 小文件不压缩
	if len(data) < e.config.CompressMinSize {
		return data, ""
	}

	// 检查缓存
	if e.config.EnableCache {
		if entry, ok := e.cache.Get(path); ok && entry.Gzipped != nil {
			acceptEncoding := c.GetHeader("Accept-Encoding")
			compressed, encoding, ok := GetCompressedData(data, entry.Gzipped, nil, acceptEncoding)
			if ok {
				return compressed, encoding
			}
		}
	}

	// 实时压缩
	acceptEncoding := c.GetHeader("Accept-Encoding")
	if containsEncoding(acceptEncoding, "gzip") {
		gzData, err := GzipCompress(data, e.config.GzipLevel)
		if err == nil && len(gzData) < len(data) {
			return gzData, "gzip"
		}
	}

	return data, ""
}

// checkNotModified 检查条件请求
func (e *StaticEngine) checkNotModified(c *gin.Context, modTime time.Time, etag string) bool {
	// ETag 检查
	if e.config.UseETag {
		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			c.Status(http.StatusNotModified)
			return true
		}
	}

	// Last-Modified 检查
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

// serveError 服务错误页面
func (e *StaticEngine) serveError(c *gin.Context, status int) {
	// 尝试自定义 404
	if status == http.StatusNotFound && e.config.Custom404 != "" {
		data, err := os.ReadFile(filepath.Join(e.config.Root, e.config.Custom404))
		if err == nil {
			c.Header("Content-Type", "text/html")
			c.Data(http.StatusNotFound, "text/html", data)
			return
		}
	}

	c.Status(status)
}

// ServeHTTP 实现 http.Handler 接口
func (e *StaticEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 安全检查
	safe, cleanPath := IsPathTraversal(e.config.Root, path)
	if !safe {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// 检查隐藏文件
	if e.config.HideDotFiles && IsHiddenPath(cleanPath) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// 获取文件
	data, modTime, etag, err := e.getFile(cleanPath)
	if err != nil {
		if e.config.EnableSPA && e.config.SPAFallback {
			data, modTime, etag, err = e.getFile(e.config.IndexFile)
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}

	// 条件请求检查
	if e.checkHTTPNotModified(r, modTime, etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// 设置响应头
	mimeType := GetMimeType(cleanPath, e.config.MimeTypes)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))

	if e.config.UseETag {
		w.Header().Set("ETag", etag)
	}

	if e.config.CacheControl != "" {
		w.Header().Set("Cache-Control", e.config.CacheControl)
	}

	// Gzip 压缩
	acceptEncoding := r.Header.Get("Accept-Encoding")
	if e.config.EnableGzip && len(data) >= e.config.CompressMinSize {
		if containsEncoding(acceptEncoding, "gzip") {
			gzData, err := GzipCompress(data, e.config.GzipLevel)
			if err == nil && len(gzData) < len(data) {
				data = gzData
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Vary", "Accept-Encoding")
			}
		}
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// checkHTTPNotModified 检查 HTTP 条件请求
func (e *StaticEngine) checkHTTPNotModified(r *http.Request, modTime time.Time, etag string) bool {
	if e.config.UseETag {
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			return true
		}
	}

	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		ifMod, err := time.Parse(http.TimeFormat, ifModifiedSince)
		if err == nil && !modTime.After(ifMod) {
			return true
		}
	}

	return false
}

// Cache 返回缓存实例（用于外部访问）
func (e *StaticEngine) Cache() *Cache {
	return e.cache
}

// Config 返回配置实例（用于外部访问）
func (e *StaticEngine) Config() *Config {
	return e.config
}

// ResetCache 重置缓存
func (e *StaticEngine) ResetCache() {
	e.cache.Clear()
}

// ReloadCache 重新加载缓存
func (e *StaticEngine) ReloadCache() error {
	e.cache.Clear()
	return e.cache.preloadDirectory(e.config.Root, e.config.EnableGzip, e.config.GzipLevel)
}

// generateETag 生成 ETag（HTTP 接口专用）
func generateETagHTTP(path string, info os.FileInfo) string {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", info.Size(), info.ModTime().UnixNano(), path)))
	return fmt.Sprintf(`"%x"`, h.Sum(nil))
}

// WriteTo 将内容写入 io.Writer
func (e *StaticEngine) WriteTo(w io.Writer, path string) (int64, error) {
	data, _, _, err := e.getFile(path)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

// preloadEmbed 预加载 embed 文件到缓存
func (e *StaticEngine) preloadEmbed() {
	if e.config.EmbedFS == nil {
		return
	}

	// 递归遍历 embed.FS
	root := e.config.EmbedRoot
	if root == "" {
		root = "."
	}

	e.walkEmbed(e.config.EmbedFS, root, func(path string) {
		e.getFile(path)
	})
}

// walkEmbed 递归遍历 embed.FS
func (e *StaticEngine) walkEmbed(fs any, path string, fn func(string)) {
	type opener interface {
		Open(name string) (fs2.File, error)
	}

	f, ok := fs.(opener)
	if !ok {
		return
	}

	file, err := f.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return
	}

	if info.IsDir() {
		// 尝试使用 ReadDir
		type readDirFile interface {
			Readdir(n int) ([]fs2.DirEntry, error)
		}
		if rdf, ok := file.(readDirFile); ok {
			entries, err := rdf.Readdir(-1)
			if err != nil {
				return
			}
			for _, entry := range entries {
				subPath := path + "/" + entry.Name()
				e.walkEmbed(fs, subPath, fn)
			}
		}
	} else {
		// 移除根前缀得到相对路径
		relPath := strings.TrimPrefix(path, "./")
		relPath = strings.TrimPrefix(relPath, ".")
		relPath = strings.TrimPrefix(relPath, "/")
		fn(relPath)
	}
}
