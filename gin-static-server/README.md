# Gin Static Server

基于 Gin 框架的高性能静态文件服务器库，可用于服务 Vue、React 等前端打包后的静态资源。

## 特性

- **高性能**: 内存缓存支持，热点文件预加载到内存
- **Gzip 压缩**: 自动 Gzip 压缩，减少网络带宽消耗
- **ETag/Last-Modified**: 支持条件请求，304 缓存优化
- **SPA 支持**: 完善的 SPA（单页应用）路由回退支持
- **安全防护**: 目录遍历防护、点文件隐藏
- **易于使用**: 链式 API 设计，单行代码集成

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
    
    // 简单使用 - 服务 ./public 目录
    ginstatic.New(r, "./public")
    
    r.Run(":8080")
}
```

### 完整配置

```go
func main() {
    r := gin.Default()
    
    ginstatic.New(r, "./public").
        WithPrefix("/static").                  // URL 前缀
        WithCache(100*1024*1024, 1000).         // 100MB, 1000 文件
        WithGzip(6).                            // Gzip 压缩级别 1-9
        WithSPA("index.html").                  // SPA 回退
        WithCacheControl("public, max-age=31536000").  // 缓存控制
        WithETag().                             // 启用 ETag
    
    r.Run(":8080")
}
```

## API 文档

### 核心函数

#### `New(router *gin.Engine, root string, opts ...Option) *StaticEngine`

创建静态文件服务器。

- `router`: Gin 引擎实例
- `root`: 静态文件根目录
- `opts`: 配置选项（可选）

#### `NewWithConfig(router *gin.Engine, cfg *Config) *StaticEngine`

使用配置结构体创建静态文件服务器。

### 配置选项

| 函数 | 描述 | 默认值 |
|------|------|--------|
| `WithPrefix(prefix string)` | 设置 URL 前缀 | `""` |
| `WithCache(maxSize int64, maxFiles int)` | 启用内存缓存 | `true` (100MB, 500文件) |
| `DisableCache()` | 禁用内存缓存 | - |
| `WithGzip(level int)` | 启用 Gzip 压缩 | `true` (级别 1) |
| `DisableGzip()` | 禁用 Gzip 压缩 | - |
| `WithSPA(indexFile string)` | 启用 SPA 回退 | `false` |
| `WithCacheControl(header string)` | 设置缓存控制头 | `"public, max-age=31536000"` |
| `WithETag()` | 启用 ETag | `true` |
| `HideDotFiles()` | 隐藏点文件 | `true` |
| `ShowDotFiles()` | 显示点文件 | - |
| `WithPreloadOnStart()` | 启动时预加载 | `false` |
| `WithMimeTypes(types map[string]string)` | 自定义 MIME 类型 | - |

### 缓存操作

```go
engine := ginstatic.New(r, "./public")

// 重置缓存
engine.ResetCache()

// 重新加载缓存
engine.ReloadCache()

// 获取缓存实例
cache := engine.Cache()
```

## 示例项目

### 服务 Vue/React 构建产物

```go
package main

import (
    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 服务前端构建产物
    ginstatic.New(r, "./dist",
        ginstatic.WithSPA("index.html"),           // Vue/React 路由支持
        ginstatic.WithGzip(6),                     // Gzip 压缩
        ginstatic.WithCache(50*1024*1024, 200),   // 50MB 缓存
    )
    
    // API 路由
    api := r.Group("/api")
    api.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    })
    
    r.Run(":8080")
}
```

### 服务静态资源目录

```go
package main

import (
    "github.com/aiqoder/my-go-tools/gin-static-server"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 服务静态文件
    ginstatic.New(r, "./static",
        ginstatic.WithPrefix("/static"),
        ginstatic.WithCacheControl("public, max-age=86400"),
    )
    
    r.Run(":8080")
}
```

### 使用 Embed 嵌入静态文件

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

    // 使用 embed.FS 服务静态文件
    ginstatic.NewEmbed(r, assets,
        ginstatic.WithPrefix(""),
        ginstatic.WithEmbedRoot("dist"),  // 指定嵌入的子目录
        ginstatic.WithSPA("index.html"),
        ginstatic.WithGzip(6),
    )

    fmt.Println("服务地址: http://localhost:8080")
    log.Fatal(r.Run(":8080"))
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
├── engine.go           # 核心引擎（handler.go）
├── cache.go            # 缓存实现
├── middleware.go       # 中间件
├── security.go         # 安全检查
├── compress.go         # Gzip 压缩
├── ginstatic_test.go   # 单元测试
├── examples_test.go    # 示例测试
├── ginstatic_bench_test.go  # 基准测试
└── README.md          # 文档
```

## 许可证

MIT License
