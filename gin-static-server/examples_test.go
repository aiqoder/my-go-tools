package ginstatic

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// Example1 基础用法示例
func Example1() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 简单使用
	_ = New(r, "./public")

	// 启动服务器
	// r.Run(":8080")
}

// Example2 完整配置示例
func Example2() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 完整配置
	_ = New(r, "./public",
		WithPrefix("/static"),
		WithCache(100*1024*1024, 1000), // 100MB, 1000 文件
		WithGzip(6),
		WithSPA("index.html"),
		WithCacheControl("public, max-age=31536000"),
		WithETag(),
	)
}

// Example3 禁用缓存和 Gzip
func Example3() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	_ = New(r, "./public",
		DisableCache(),
		DisableGzip(),
	)
}

// Example4 SPA 应用
func Example4() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// SPA 模式 - 所有路由都返回 index.html
	_ = New(r, "./dist",
		WithSPA("index.html"),
		WithPrefix(""),
	)
}

// Example5 自定义 MIME 类型
func Example5() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	_ = New(r, "./public",
		WithMimeTypes(map[string]string{
			"xyz": "application/x-custom",
			"abc": "text/plain",
		}),
	)
}

// TestNewStaticEngine 测试静态引擎创建
func TestNewStaticEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	engine := New(r, "./testdata")

	if engine == nil {
		t.Fatal("expected engine, got nil")
	}

	if engine.config == nil {
		t.Error("expected config, got nil")
	}

	if engine.cache == nil {
		t.Error("expected cache, got nil")
	}
}

// TestNewWithConfig 测试带配置的创建
func TestNewWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &Config{
		Root:       "./testdata",
		Prefix:     "/static",
		EnableGzip: true,
		EnableCache: true,
	}

	engine := NewWithConfig(r, cfg)

	if engine == nil {
		t.Fatal("expected engine, got nil")
	}

	if engine.config.Prefix != "/static" {
		t.Errorf("expected prefix '/static', got '%s'", engine.config.Prefix)
	}
}

// TestStaticEngineServe 测试文件服务
func TestStaticEngineServe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 创建测试目录和文件
	tmpDir := t.TempDir()
	testContent := []byte("Hello, World!")
	err := os.WriteFile(tmpDir+"/test.txt", testContent, 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_ = New(r, tmpDir)

	// 创建测试请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello, World!" {
		t.Errorf("expected body 'Hello, World!', got '%s'", w.Body.String())
	}
}

// TestStaticEngine404 测试 404 处理
func TestStaticEngine404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	_ = New(r, tmpDir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestStaticEngineSPA 测试 SPA 回退
func TestStaticEngineSPA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	// 创建 index.html
	err := os.WriteFile(tmpDir+"/index.html", []byte("<html>SPA</html>"), 0644)
	if err != nil {
		t.Fatalf("failed to create index.html: %v", err)
	}

	_ = New(r, tmpDir, WithSPA("index.html"))

	// 请求不存在的路径
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/any/path", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "<html>SPA</html>" {
		t.Errorf("expected body '<html>SPA</html>', got '%s'", w.Body.String())
	}
}

// TestStaticEngineGzip 测试 Gzip 压缩
func TestStaticEngineGzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	// 创建大文件以触发压缩
	content := make([]byte, 2048)
	for i := range content {
		content[i] = 'x'
	}
	err := os.WriteFile(tmpDir+"/large.txt", content, 0644)
	if err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	_ = New(r, tmpDir, WithGzip(6))

	// 请求带 Accept-Encoding: gzip
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/large.txt", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	encoding := w.Header().Get("Content-Encoding")
	if encoding != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got '%s'", encoding)
	}
}

// TestStaticEngineETag 测试 ETag
func TestStaticEngineETag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	err := os.WriteFile(tmpDir+"/test.txt", []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_ = New(r, tmpDir, WithETag())

	// 第一次请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test.txt", nil)
	r.ServeHTTP(w, req)

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("expected ETag header")
	}

	// 使用 ETag 再次请求
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test.txt", nil)
	req2.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Errorf("expected status 304, got %d", w2.Code)
	}
}

// TestStaticEnginePrefix 测试 URL 前缀
func TestStaticEnginePrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	err := os.WriteFile(tmpDir+"/test.txt", []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_ = New(r, tmpDir, WithPrefix("/static"))

	// 请求带前缀
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/static/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestStaticEngineHideDotFiles 测试隐藏点文件
func TestStaticEngineHideDotFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	// 创建隐藏文件
	err := os.WriteFile(tmpDir+"/.hidden", []byte("secret"), 0644)
	if err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}

	_ = New(r, tmpDir, HideDotFiles())

	// 请求隐藏文件
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/.hidden", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// TestStaticEnginePathTraversal 测试路径遍历防护
func TestStaticEnginePathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	_ = New(r, tmpDir)

	// 尝试路径遍历
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/../../etc/passwd", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// TestCacheOperations 测试缓存操作
func TestCacheOperations(t *testing.T) {
	cache := NewCache(1024, 2, nil)

	entry := &cacheEntry{
		Data: []byte("test"),
		Size: 4,
	}

	// 添加文件
	cache.Set("file1.txt", entry)
	if cache.FileCount() != 1 {
		t.Errorf("expected 1 file, got %d", cache.FileCount())
	}

	// 获取文件
	e, ok := cache.Get("file1.txt")
	if !ok {
		t.Error("expected to get file from cache")
	}
	if string(e.Data) != "test" {
		t.Errorf("expected 'test', got '%s'", string(e.Data))
	}

	// 超过限制，触发淘汰
	cache.Set("file2.txt", entry)
	cache.Set("file3.txt", entry)

	// 缓存大小应该保持
	if cache.FileCount() > 2 {
		t.Errorf("expected <= 2 files, got %d", cache.FileCount())
	}
}
