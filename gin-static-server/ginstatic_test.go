package ginstatic

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 创建缓存
	cache := NewCache(1024*1024, 10, nil)

	// 测试 Set 和 Get
	entry := &cacheEntry{
		Data:    []byte("Hello, World!"),
		ModTime: time.Now(),
		Size:    13,
		ETag:    "test-etag",
		Path:    testFile,
	}

	cache.Set("test.txt", entry)

	// 测试 Get
	got, ok := cache.Get("test.txt")
	if !ok {
		t.Fatal("failed to get cached entry")
	}
	if string(got.Data) != "Hello, World!" {
		t.Errorf("expected data 'Hello, World!', got '%s'", string(got.Data))
	}

	// 测试 Size 和 FileCount
	if cache.Size() == 0 {
		t.Error("cache size should be greater than 0")
	}
	if cache.FileCount() != 1 {
		t.Errorf("expected file count 1, got %d", cache.FileCount())
	}

	// 测试 Delete
	cache.Delete("test.txt")
	if cache.FileCount() != 0 {
		t.Errorf("expected file count 0, got %d", cache.FileCount())
	}

	// 测试 Clear
	cache.Set("test1.txt", entry)
	cache.Set("test2.txt", entry)
	cache.Clear()
	if cache.FileCount() != 0 {
		t.Errorf("expected file count 0 after clear, got %d", cache.FileCount())
	}
}

func TestCacheEviction(t *testing.T) {
	// 创建缓存，限制为只能存 2 个文件
	cache := NewCache(1024*1024, 2, nil)

	// 添加 3 个文件，触发淘汰
	entry := &cacheEntry{
		Data: []byte("test"),
		Size: 4,
	}

	cache.Set("file1.txt", entry)
	cache.Set("file2.txt", entry)
	cache.Set("file3.txt", entry)

	// 第三个文件应该被添加，前两个中的一个应该被淘汰
	if cache.FileCount() > 2 {
		t.Errorf("expected file count <= 2, got %d", cache.FileCount())
	}
}

func TestCacheLoadFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Hello, World!"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cache := NewCache(1024*1024, 10, nil)

	// 测试 LoadFile
	entry, err := cache.LoadFile(tmpDir, "test.txt", true, gzip.BestSpeed)
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}

	if entry == nil {
		t.Fatal("expected entry, got nil")
	}

	if string(entry.Data) != "Hello, World!" {
		t.Errorf("expected data 'Hello, World!', got '%s'", string(entry.Data))
	}

	// 测试缓存
	entry2, err := cache.LoadFile(tmpDir, "test.txt", true, gzip.BestSpeed)
	if err != nil {
		t.Fatalf("failed to load file from cache: %v", err)
	}

	// 应该是同一个指针
	if entry != entry2 {
		t.Error("expected same entry from cache")
	}
}

func TestGzipCompress(t *testing.T) {
	data := []byte("Hello, World! This is a test message for gzip compression.")

	// 测试压缩
	compressed, err := GzipCompress(data, gzip.BestSpeed)
	if err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	// 压缩后的数据应该比原始数据小
	if len(compressed) >= len(data) {
		t.Errorf("compressed data should be smaller than original, got %d >= %d", len(compressed), len(data))
	}

	// 测试解压
	decompressed, err := GzipDecompress(compressed)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != string(data) {
		t.Errorf("decompressed data mismatch")
	}
}

func TestZstdCompress(t *testing.T) {
	data := []byte("Hello, World! This is a test message for zstd compression.")

	// 测试压缩
	compressed, err := ZstdCompress(data)
	if err != nil {
		t.Fatalf("failed to compress: %v", err)
	}

	// 压缩后的数据应该比原始数据小
	if len(compressed) >= len(data) {
		t.Errorf("compressed data should be smaller than original, got %d >= %d", len(compressed), len(data))
	}

	// 测试解压
	decompressed, err := ZstdDecompress(compressed)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decompressed) != string(data) {
		t.Errorf("decompressed data mismatch")
	}
}

func TestGetCompressedData(t *testing.T) {
	data := []byte("Hello, World!")
	gzData := []byte{0x1f, 0x8b} // 简化的 gzip 数据

	tests := []struct {
		name         string
		acceptEnc    string
		wantEncoding string
		wantUsed     bool
	}{
		{"gzip", "gzip", "gzip", true},
		{"gzip with q", "gzip;q=0.5", "gzip", true},
		{"no gzip", "deflate", "", false},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, encoding, used := GetCompressedData(data, gzData, nil, tt.acceptEnc)
			if encoding != tt.wantEncoding {
				t.Errorf("expected encoding '%s', got '%s'", tt.wantEncoding, encoding)
			}
			if used != tt.wantUsed {
				t.Errorf("expected used %v, got %v", tt.wantUsed, used)
			}
			_ = result
		})
	}
}

func TestContainsEncoding(t *testing.T) {
	tests := []struct {
		accept string
		enc    string
		want   bool
	}{
		{"gzip", "gzip", true},
		{"gzip, deflate", "gzip", true},
		{"deflate, gzip", "gzip", true},
		{"gzip;q=0.5", "gzip", true},
		{"gzip;q=0", "gzip", false},
		{"deflate", "gzip", false},
		{"", "gzip", false},
	}

	for _, tt := range tests {
		t.Run(tt.accept+"_"+tt.enc, func(t *testing.T) {
			got := containsEncoding(tt.accept, tt.enc)
			if got != tt.want {
				t.Errorf("containsEncoding(%q, %q) = %v, want %v", tt.accept, tt.enc, got, tt.want)
			}
		})
	}
}

func TestIsPathTraversal(t *testing.T) {
	root := "/var/www/static"

	tests := []struct {
		name    string
		path    string
		wantOk  bool
		wantStr string
	}{
		{"normal", "js/app.js", true, "/js/app.js"},
		{"normal subdir", "css/main.css", true, "/css/main.css"},
		{"traversal", "../secret.txt", false, ""},
		{"double traversal", "../../etc/passwd", false, ""},
		{"traversal with path", "js/../../secret.txt", false, ""},
		{"absolute", "/etc/passwd", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, got := IsPathTraversal(root, tt.path)
			if ok != tt.wantOk {
				t.Errorf("IsPathTraversal(%q) ok = %v, want %v", tt.path, ok, tt.wantOk)
			}
			if got != tt.wantStr {
				t.Errorf("IsPathTraversal(%q) = %q, want %q", tt.path, got, tt.wantStr)
			}
		})
	}
}

func TestIsHiddenFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".gitignore", true},
		{".env", true},
		{"config.json", false},
		{"dir/.hidden", true},
		{"dir/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsHiddenFile(tt.path)
			if got != tt.want {
				t.Errorf("IsHiddenFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsHiddenPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".gitignore", true},
		{".env", true},
		{"css/style.css", false},
		{"dir/.hidden/file.txt", true},
		{"dir/file.txt", false},
		{"./file.txt", true},
		{"..hidden/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsHiddenPath(tt.path)
			if got != tt.want {
				t.Errorf("IsHiddenPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsValidFileName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"file.txt", true},
		{"file.txt.bak", true},
		{"", false},
		{"nul", false},
		{"CON", false},
		{"LPT1", false},
		{"file/name.txt", false},
		{"file\\name.txt", false},
		{"file\x00name.txt", false},
		{"normal-file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidFileName(tt.name)
			if got != tt.want {
				t.Errorf("IsValidFileName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"js/app.js", "js" + string(filepath.Separator) + "app.js"},
		{"/js/app.js", "js" + string(filepath.Separator) + "app.js"},
		{"//js//app.js", "js" + string(filepath.Separator) + "app.js"},
		{"js/../css/app.css", "css" + string(filepath.Separator) + "app.css"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizePath(tt.input)
			// 跨平台处理
			want := filepath.Clean(tt.want)
			if got != want {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, got, want)
			}
		})
	}
}

func TestShouldServeFile(t *testing.T) {
	tests := []struct {
		path         string
		hideDotFiles bool
		wantOk       bool
	}{
		{"file.txt", true, true},
		{".hidden", true, false},
		{"file.txt", false, true},
		{".hidden", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			ok, _ := ShouldServeFile(tt.path, tt.hideDotFiles, "")
			if ok != tt.wantOk {
				t.Errorf("ShouldServeFile(%q, %v) = %v, want %v", tt.path, tt.hideDotFiles, ok, tt.wantOk)
			}
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"app.js", "application/javascript; charset=utf-8"},
		{"style.css", "text/css; charset=utf-8"},
		{"index.html", "text/html; charset=utf-8"},
		{"image.png", "image/png"},
		{"image.svg", "image/svg+xml"},
		{"font.woff2", "font/woff2"},
		{"unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetMimeType(tt.filename, nil)
			if got != tt.want {
				t.Errorf("GetMimeType(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestGetMimeTypeCustom(t *testing.T) {
	customTypes := map[string]string{
		"xyz": "application/x-custom",
	}

	got := GetMimeType("file.xyz", customTypes)
	if got != "application/x-custom" {
		t.Errorf("GetMimeType with custom types = %q, want %q", got, "application/x-custom")
	}
}

func TestGetGzipLevel(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{1, 1},
		{5, 5},
		{9, 9},
		{0, gzip.BestSpeed},
		{-1, gzip.BestSpeed},
		{10, gzip.BestSpeed},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.input)), func(t *testing.T) {
			got := GetGzipLevel(tt.input)
			if got != tt.want {
				t.Errorf("GetGzipLevel(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
