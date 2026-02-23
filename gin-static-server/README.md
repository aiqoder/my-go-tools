# Gin Static Server

基于 Gin 框架的高性能静态文件服务器中间件，解决路由冲突问题。

## 特性

- **解决路由冲突**: 只拦截静态资源请求（.js, .css, .html 等），不阻塞 API 路由
- **高性能**: 内存缓存支持，热点文件预加载到内存
- **Gzip 压缩**: 自动 Gzip 压缩，减少网络带宽消耗
- **ETag/Last-Modified**: 支持条件请求，304 缓存优化
- **安全防护**: 目录遍历防护、点文件隐藏

## 安装

```bash
go get github.com/aiqoder/my-go-tools/gin-static-server
```

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 使用中间件方式（推荐）
    r.Use(ginstatic.StaticFileExtsMiddleware("./public"))

    // API 路由不再被阻塞
    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    })

    r.Run(":8080")
}
```

### 完整配置

```go
func main() {
    r := gin.Default()

    r.Use(ginstatic.StaticFileExtsMiddleware("./public",
        ginstatic.WithMiddlewarePrefix("/static"),        // URL 前缀
        ginstatic.WithMiddlewareCache(100*1024*1024, 1000), // 100MB, 1000 文件
        ginstatic.WithMiddlewareGzip(6),                // Gzip 压缩级别 1-9
        ginstatic.WithMiddlewareCacheControl("public, max-age=60"), // 缓存控制
        ginstatic.WithMiddlewareETag(),                  // 启用 ETag
    ))

    r.Run(":8080")
}
```

## API 文档

### 核心函数

#### `StaticFileExtsMiddleware(root string, opts ...MiddlewareOption) gin.HandlerFunc`

创建静态文件扩展名中间件。

- `root`: 静态文件根目录
- `opts`: 配置选项（可选）

#### `NewStaticFileExtsMiddlewareWithConfig(cfg *StaticExtsMiddlewareConfig) gin.HandlerFunc`

使用配置结构体创建中间件。

### 配置选项

| 函数 | 描述 | 默认值 |
|------|------|--------|
| `WithMiddlewarePrefix(prefix string)` | 设置 URL 前缀 | `""` |
| `WithMiddlewareStaticExts(exts []string)` | 自定义静态资源扩展名 | 20+ 种常见扩展名 |
| `WithMiddlewareCache(maxSize int64, maxFiles int)` | 启用内存缓存 | `true` (100MB, 500文件) |
| `DisableMiddlewareCache()` | 禁用内存缓存 | - |
| `WithMiddlewareGzip(level int)` | 启用 Gzip 压缩 | `true` (级别 1) |
| `DisableMiddlewareGzip()` | 禁用 Gzip 压缩 | - |
| `WithMiddlewareETag()` | 启用 ETag | `true` |
| `WithoutMiddlewareETag()` | 禁用 ETag | - |
| `WithMiddlewareCacheControl(control string)` | 设置缓存控制头 | `"public, max-age=60"` |
| `WithMiddlewareHideDotFiles()` | 隐藏点文件 | `true` |
| `WithMiddlewareShowDotFiles()` | 显示点文件 | - |
| `WithMiddlewareIndex()` | 启用 index.html 回退（访问 / 自动返回 index.html） | `true` |
| `DisableMiddlewareIndex()` | 禁用 index.html 回退 | - |
| `WithMiddlewareIndexFile(filename string)` | 设置默认索引文件 | `"index.html"` |
| `WithMiddlewareEmbedFS(fs any, root string)` | 使用 embed.FS | - |
| `WithMiddlewareOnRequest(fn func(string) bool)` | 请求前回调 | - |

## 使用示例

### 解决路由冲突

```go
package main

import (
    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 使用中间件 - 只拦截 .js, .css, .html 等静态资源
    r.Use(ginstatic.StaticFileExtsMiddleware("./public"))

    // API 路由正常工作，不会被拦截
    r.GET("/api/v1/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"users": []string{"user1", "user2"}})
    })

    r.GET("/api/v1/products", func(c *gin.Context) {
        c.JSON(200, gin.H{"products": []string{"prod1", "prod2"}})
    })

    r.Run(":8080")
}
```

### 使用 embed.FS

```go
package main

import (
    "embed"
    "fmt"
    "log"

    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

//go:embed dist
var assets embed.FS

func main() {
    r := gin.Default()

    r.Use(ginstatic.StaticFileExtsMiddleware("",
        ginstatic.WithMiddlewareEmbedFS(assets, "dist"),
    ))

    fmt.Println("服务地址: http://localhost:8080")
    log.Fatal(r.Run(":8080"))
}
```

### 自定义静态资源扩展名

```go
package main

import (
    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 只拦截 .js 和 .css 文件
    r.Use(ginstatic.StaticFileExtsMiddleware("./public",
        ginstatic.WithMiddlewareStaticExts([]string{".js", ".css"}),
    ))

    // .html, .json 等请求会被传递给下一个处理器
    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"data": "api response"})
    })

    r.Run(":8080")
}
```

## 性能基准

运行基准测试:

```bash
go test -bench=. -benchtime=1s ./...
```

典型结果（因硬件而异）:

```
BenchmarkGzipCompress-8           100000   ~15 µs/op
BenchmarkCacheGet-8              5000000   ~250 ns/op
BenchmarkFileServing-8            50000   ~30 µs/op
BenchmarkFileServingWithCache-8  100000   ~12 µs/op
```

## 目录结构

```
gin-static-server/
├── options.go          # 配置选项
├── handler.go         # 核心引擎
├── cache.go           # 缓存实现
├── middleware.go      # 中间件实现
├── security.go        # 安全检查
├── compress.go        # Gzip 压缩
├── middleware_test.go # 中间件测试
├── ginstatic_test.go  # 单元测试
├── examples/         # 示例程序
└── README.md         # 文档
```

## 许可证

MIT License
