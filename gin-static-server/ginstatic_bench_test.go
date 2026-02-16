package ginstatic

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// BenchmarkCacheGet 缓存获取基准测试
func BenchmarkCacheGet(b *testing.B) {
	cache := NewCache(100*1024*1024, 1000, nil)

	entry := &cacheEntry{
		Data:    make([]byte, 1024),
		Size:    1024,
		ModTime: time.Now(),
	}

	// 预填充缓存
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("dir/file%d.txt", i), entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("dir/file50.txt")
	}
}

// BenchmarkCacheSet 缓存设置基准测试
func BenchmarkCacheSet(b *testing.B) {
	cache := NewCache(100*1024*1024, 1000, nil)

	entry := &cacheEntry{
		Data: make([]byte, 1024),
		Size: 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("dir/file%d.txt", i%100), entry)
	}
}

// BenchmarkGzipCompress Gzip 压缩基准测试
func BenchmarkGzipCompress(b *testing.B) {
	data := make([]byte, 10*1024) // 10KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GzipCompress(data, gzip.BestSpeed)
	}
}

// BenchmarkGzipDecompress Gzip 解压基准测试
func BenchmarkGzipDecompress(b *testing.B) {
	data := make([]byte, 10*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	compressed, _ := GzipCompress(data, gzip.BestSpeed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GzipDecompress(compressed)
	}
}

// BenchmarkZstdCompress Zstd 压缩基准测试
func BenchmarkZstdCompress(b *testing.B) {
	data := make([]byte, 10*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ZstdCompress(data)
	}
}

// BenchmarkZstdDecompress Zstd 解压基准测试
func BenchmarkZstdDecompress(b *testing.B) {
	data := make([]byte, 10*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	compressed, _ := ZstdCompress(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ZstdDecompress(compressed)
	}
}

// BenchmarkGetMimeType MIME 类型获取基准测试
func BenchmarkGetMimeType(b *testing.B) {
	filenames := []string{
		"app.js",
		"style.css",
		"index.html",
		"image.png",
		"font.woff2",
		"unknown.xyz",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMimeType(filenames[i%len(filenames)], nil)
	}
}

// BenchmarkIsPathTraversal 路径遍历检测基准测试
func BenchmarkIsPathTraversal(b *testing.B) {
	paths := []string{
		"js/app.js",
		"css/main.css",
		"../secret.txt",
		"js/../../etc/passwd",
		"/etc/passwd",
	}

	root := "/var/www/static"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPathTraversal(root, paths[i%len(paths)])
	}
}

// BenchmarkGetCompressedData 压缩数据选择基准测试
func BenchmarkGetCompressedData(b *testing.B) {
	data := make([]byte, 10*1024)
	gzData, _ := GzipCompress(data, gzip.BestSpeed)

	encodings := []string{
		"gzip",
		"deflate",
		"gzip, deflate",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCompressedData(data, gzData, nil, encodings[i%len(encodings)])
	}
}

// BenchmarkFileServing 文件服务基准测试
func BenchmarkFileServing(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := b.TempDir()
	// 创建测试文件
	content := make([]byte, 1024)
	os.WriteFile(tmpDir+"/test.txt", content, 0644)

	_ = New(r, tmpDir, DisableCache())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test.txt", nil)
		r.ServeHTTP(w, req)
	}
}

// BenchmarkFileServingWithCache 带缓存的文件服务基准测试
func BenchmarkFileServingWithCache(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := b.TempDir()
	content := make([]byte, 1024)
	os.WriteFile(tmpDir+"/test.txt", content, 0644)

	_ = New(r, tmpDir, WithCache(100*1024*1024, 100))

	// 预热缓存
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test.txt", nil)
	r.ServeHTTP(w, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test.txt", nil)
		r.ServeHTTP(w, req)
	}
}

// BenchmarkFileServingWithGzip 带 Gzip 的文件服务基准测试
func BenchmarkFileServingWithGzip(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := b.TempDir()
	// 创建大文件以触发压缩
	content := make([]byte, 10*1024)
	os.WriteFile(tmpDir+"/large.txt", content, 0644)

	_ = New(r, tmpDir, WithGzip(gzip.BestSpeed))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/large.txt", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentCacheGet 并发缓存获取基准测试
func BenchmarkConcurrentCacheGet(b *testing.B) {
	cache := NewCache(100*1024*1024, 1000, nil)

	entry := &cacheEntry{
		Data:    make([]byte, 1024),
		Size:    1024,
		ModTime: time.Now(),
	}

	// 预填充缓存
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("dir/file%d.txt", i), entry)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("dir/file50.txt")
		}
	})
}
