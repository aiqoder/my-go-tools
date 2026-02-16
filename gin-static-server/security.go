package ginstatic

import (
	"path/filepath"
	"strings"
)

// IsPathTraversal 检查路径是否存在目录遍历风险
// root: 静态文件根目录
// relPath: 请求的相对路径
// 返回: 是否安全，以及清理后的路径
func IsPathTraversal(root, relPath string) (bool, string) {
	// 清理路径
	cleanPath := filepath.Clean("/" + relPath)

	// 再次检查是否包含 ..
	if strings.Contains(cleanPath, "..") {
		return false, ""
	}

	// 计算绝对路径
	absPath := filepath.Join(root, cleanPath)

	// 检查绝对路径是否在根目录内
	rootAbs := filepath.Clean(root)
	absAbs := filepath.Clean(absPath)

	// 使用路径前缀检查
	if !strings.HasPrefix(absAbs, rootAbs) {
		return false, ""
	}

	// 再次检查清理后的路径
	cleanAbsPath := filepath.Clean(absPath)
	if !strings.HasPrefix(cleanAbsPath, rootAbs) {
		return false, ""
	}

	return true, cleanPath
}

// IsHiddenFile 检查是否为隐藏文件（以 . 开头）
func IsHiddenFile(path string) bool {
	// 检查路径的最后一个元素
	base := filepath.Base(path)
	return len(base) > 0 && base[0] == '.'
}

// IsHiddenPath 检查路径中是否包含隐藏目录或文件
func IsHiddenPath(path string) bool {
	parts := filepath.SplitList(path)
	for _, part := range parts {
		if len(part) > 0 && part[0] == '.' && part != "." && part != ".." {
			return true
		}
	}
	return false
}

// IsValidFileName 检查文件名是否有效
func IsValidFileName(name string) bool {
	if name == "" {
		return false
	}

	// 检查无效字符
	invalidChars := []string{"/", "\\", "\x00", "\n", "\r"}
	for _, c := range invalidChars {
		if strings.Contains(name, c) {
			return false
		}
	}

	// 检查保留名称（Windows）
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4",
		"COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4",
		"LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	for _, r := range reserved {
		if upperName == r || strings.HasPrefix(upperName, r+".") {
			return false
		}
	}

	return true
}

// SanitizePath 清理并规范化路径
func SanitizePath(path string) string {
	// 移除多余的斜杠
	path = filepath.ToSlash(path)
	path = strings.ReplaceAll(path, "//", "/")

	// 清理路径
	path = filepath.Clean(path)

	// 移除前导斜杠（相对于根目录）
	path = strings.TrimPrefix(path, "/")

	return path
}

// ShouldServeFile 检查文件是否应该被服务
// hideDotFiles: 是否隐藏点文件
// custom404: 自定义 404 页面路径
func ShouldServeFile(path string, hideDotFiles bool, custom404 string) (bool, string) {
	// 检查隐藏文件
	if hideDotFiles && IsHiddenPath(path) {
		return false, custom404
	}

	// 检查无效文件名
	if !IsValidFileName(filepath.Base(path)) {
		return false, custom404
	}

	return true, ""
}

// ContentTypeFromPath 从路径推断内容类型
func ContentTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".webp":
		return "image/webp"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".otf":
		return "font/otf"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}
