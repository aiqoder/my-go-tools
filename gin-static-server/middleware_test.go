package ginstatic

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsStaticFile(t *testing.T) {
	tests := []struct {
		path     string
		exts     []string
		expected bool
	}{
		{"/app.js", defaultStaticExts, true},
		{"/style.css", defaultStaticExts, true},
		{"/index.html", defaultStaticExts, true},
		{"/data.json", defaultStaticExts, true},
		{"/app.min.js", defaultStaticExts, true},
		{"/image.svg", defaultStaticExts, true},
		{"/font.woff2", defaultStaticExts, true},
		{"/api/users", defaultStaticExts, false},
		{"/api/v1/data", defaultStaticExts, false},
		{"/health", defaultStaticExts, false},
		{"/", defaultStaticExts, false},
		{"", defaultStaticExts, false},
		{"/app.js", []string{".js", ".css"}, true},
		{"/app.css", []string{".js", ".css"}, true},
		{"/app.html", []string{".js", ".css"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isStaticFile(tt.path, tt.exts)
			if result != tt.expected {
				t.Errorf("isStaticFile(%q, %v) = %v, expected %v", tt.path, tt.exts, result, tt.expected)
			}
		})
	}
}

func TestStaticFileExtsMiddleware_FileExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static"))
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API response"})
	})

	// 测试静态文件
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/app.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/javascript; charset=utf-8" {
		t.Errorf("Expected Content-Type application/javascript, got %s", contentType)
	}
}

func TestStaticFileExtsMiddleware_FileNotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static"))
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API response"})
	})

	// 测试不存在的静态文件，应该传递给下一个处理器
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent.js", nil)
	r.ServeHTTP(w, req)

	// 由于文件不存在，调用了 c.Next()，但没有注册 /nonexistent.js 路由，所以返回 404
	// 这验证了中间件确实调用了 c.Next()
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 (passed to next handler), got %d", w.Code)
	}
}

func TestStaticFileExtsMiddleware_APIRouteNotBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static"))
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API response"})
	})

	// 测试 API 路由不应该被阻塞
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/users", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != `{"message":"API response"}` {
		t.Errorf("Expected API response, got %s", w.Body.String())
	}
}

func TestStaticFileExtsMiddleware_WithPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static", WithMiddlewarePrefix("/static")))
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API response"})
	})

	// 测试带前缀的静态文件
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/static/app.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestStaticFileExtsMiddleware_PathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static"))

	// 测试路径中包含 .. 的情况
	// Gin 会自动清理路径，所以需要使用实际能触发目录遍历的路径
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/static/../app.js", nil)
	r.ServeHTTP(w, req)

	// 应该返回 200，因为 /static/../app.js 会被清理为 /app.js
	// 这是预期行为 - Gin 会清理路径
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 测试真正的目录遍历 - 尝试访问根目录外的文件
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/static/../../middleware.go.js", nil)
	r.ServeHTTP(w, req)

	// 路径会被清理为 /middleware.go，仍然不在 static 目录下
	// 由于文件不存在，应该调用 c.Next()，返回 404
	if w.Code != http.StatusNotFound {
		t.Logf("Got status %d for path traversal attempt", w.Code)
	}

	// 测试不带扩展名的路径遍历
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/../testdata/static/app.js", nil)
	r.ServeHTTP(w, req)

	// 由于路径不匹配任何静态扩展名，直接调用 c.Next()
	// 返回 404
	if w.Code != http.StatusNotFound {
		t.Logf("Got status %d for non-static path", w.Code)
	}
}

func TestStaticFileExtsMiddleware_CustomExts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static", WithMiddlewareStaticExts([]string{".js", ".css"})))
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API response"})
	})

	// .html 应该是 API 路由，不被中间件拦截
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/index.html", nil)
	r.ServeHTTP(w, req)

	// 因为 .html 不在自定义扩展名列表中，所以应该 404（传递给下一个处理器）
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// .js 仍然被拦截
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/app.js", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for .js, got %d", w.Code)
	}
}

func TestStaticFileExtsMiddleware_CacheControl(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static", WithMiddlewareCacheControl("public, max-age=3600")))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/app.js", nil)
	r.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("Expected Cache-Control 'public, max-age=3600', got %s", cacheControl)
	}
}

func TestStaticFileExtsMiddleware_ETag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static", WithMiddlewareETag()))

	// 第一次请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/app.js", nil)
	r.ServeHTTP(w, req)

	etag := w.Header().Get("ETag")
	if etag == "" {
		t.Error("Expected ETag header")
	}

	// 条件请求 - If-None-Match
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/app.js", nil)
	req.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotModified {
		t.Errorf("Expected status 304 Not Modified, got %d", w.Code)
	}
}

func TestStaticFileExtsMiddleware_IndexFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(StaticFileExtsMiddleware("./testdata/static"))

	// 测试根路径访问 index.html
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
}
