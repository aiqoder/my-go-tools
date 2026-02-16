# Gin 静态文件服务器设计文档

## 1. 架构概述

### 1.1 模块设计

```
┌─────────────────────────────────────────────────────────┐
│                    gin-static-server                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────┐    ┌──────────────┐    ┌───────────┐  │
│  │  Option API │───▶│ StaticEngine │───▶│  Handlers │  │
│  │  (配置层)    │    │   (核心引擎)   │    │  (处理器)  │  │
│  └─────────────┘    └──────────────┘    └───────────┘  │
│         │                  │                   │        │
│         ▼                  ▼                   ▼        │
│  ┌─────────────────────────────────────────────────────┐│
│  │              Middleware (中间件层)                    ││
│  │  • 缓存控制    • Gzip 压缩    • 安全检查             ││
│  └─────────────────────────────────────────────────────┘│
│                          │                               │
│                          ▼                               │
│  ┌─────────────────────────────────────────────────────┐│
│  │         Storage Layer (存储层)                        ││
│  │  • 文件系统    • 内存缓存     • Byte Pool            ││
│  └─────────────────────────────────────────────────────┘│
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 1.2 核心组件

| 组件 | 职责 |
|------|------|
| `Option` | 配置选项模式，支持链式调用 |
| `StaticEngine` | 静态文件服务引擎，管理缓存和配置 |
| `FileHandler` | 文件请求处理，包含安全检查、缓存、压缩 |
| `Cache` | 内存缓存层，使用 sync.Map 存储 |
| `Middleware` | Gin 中间件，处理请求前后逻辑 |

## 2. 数据结构设计

### 2.1 配置结构

```go
type Config struct {
    // 基础配置
    Root       string   // 静态文件根目录
    Prefix     string   // URL 前缀路径
    
    // 缓存配置
    EnableCache    bool      // 是否启用内存缓存
    MaxCacheSize   int64     // 最大缓存大小（字节）
    MaxCacheFiles  int       // 最大缓存文件数
    
    // 压缩配置
    EnableGzip      bool   // 是否启用 Gzip 压缩
    GzipLevel       int    // Gzip 压缩级别 (1-9)
    CompressMinSize int    // 最小压缩大小
    
    // SPA 支持
    EnableSPA        bool   // 是否启用 SPA 回退
    IndexFile        string // index.html 路径
    
    // 安全配置
    HideDotFiles bool // 是否隐藏点文件
    
    // 缓存控制
    CacheControl string // 缓存控制头
    UseETag      bool   // 是否使用 ETag
}
```

### 2.2 缓存结构

```go
type cacheEntry struct {
    data       []byte  // 文件内容
    gzipped    []byte // Gzip 压缩后的内容
    modTime    time.Time
    size       int64
    etag       string
    lastAccess int64  // 最后访问时间（用于 LRU 淘汰）
}
```

## 3. API 设计

### 3.1 入口函数

```go
// 默认配置创建静态服务器
func New(router *gin.Engine, root string) *StaticEngine

// 自定义配置创建静态服务器
func NewWithConfig(router *gin.Engine, cfg Config) *StaticEngine
```

### 3.2 链式配置 API

```go
// 启用内存缓存
func (s *StaticEngine) WithCache(maxSize int64, maxFiles int) *StaticEngine

// 启用 Gzip 压缩
func (s *StaticEngine) WithGzip(level int) *StaticEngine

// 设置 URL 前缀
func (s *StaticEngine) WithPrefix(prefix string) *StaticEngine

// 启用 SPA 回退
func (s *StaticEngine) WithSPA(indexFile string) *StaticEngine

// 设置缓存控制
func (s *StaticEngine) WithCacheControl(header string) *StaticEngine

// 使用 ETag
func (s *StaticEngine) WithETag() *StaticEngine

// 隐藏点文件
func (s *StaticEngine) HideDotFiles() *StaticEngine
```

## 4. 核心流程设计

### 4.1 文件请求处理流程

```
请求进入
    │
    ▼
┌─────────────────┐
│ 安全检查        │ ──▶ 目录遍历检测 ──▶ 拒绝访问
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 点文件检查      │ ──▶ 是否隐藏 ──▶ 拒绝访问
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 缓存查找        │ ──▶ 命中 ──▶ 返回缓存内容
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 读取文件        │ ──▶ 加载到缓存（如启用）
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Gzip 处理       │ ──▶ 检查 Accept-Encoding
│                 │ ──▶ 返回压缩/非压缩内容
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 缓存头设置      │ ──▶ ETag / Last-Modified
│                 │ ──▶ Cache-Control
└─────────────────┘
    │
    ▼
响应返回
```

### 4.2 条件请求处理 (304 Not Modified)

```
接收请求
    │
    ▼
┌─────────────────┐
│ 检查 ETag       │ ──▶ If-None-Match 匹配 ──▶ 返回 304
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 检查时间        │ ──▶ If-Modified-Since 匹配 ──▶ 返回 304
└─────────────────┘
    │
    ▼
返回完整内容 (200)
```

## 5. 技术决策

### 5.1 缓存策略

- **存储**: 使用 `sync.Map` 存储缓存项，天然支持并发安全
- **淘汰**: 简单 LRU 实现，当缓存大小或文件数超限时，清理最旧的条目
- **同步**: 缓存加载使用 `sync.Once` 防止并发重复加载

### 5.2 Gzip 压缩

- **时机**: 服务启动时预压缩（可选）或首次访问时压缩后缓存
- **级别**: 默认使用 `gzip.BestSpeed` (1)，可配置
- **协商**: 根据请求的 `Accept-Encoding` 头决定是否返回压缩内容

### 5.3 安全防护

- **目录遍历**: 使用 `filepath.Clean` 清理路径，确保解析后仍在根目录内
- **点文件**: 默认隐藏以 `.` 开头的文件，可配置关闭

### 5.4 Byte Pool

- 使用 `sync.Pool` 复用 Gzip 压缩缓冲区，减少内存分配

## 6. 使用示例

### 6.1 基础用法

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/aiqoder/my-go-tools/gin-static-server"
)

func main() {
    r := gin.Default()
    
    // 简单使用
    gss := ginstatic.New(r, "./public")
    
    r.Run(":8080")
}
```

### 6.2 完整配置

```go
func main() {
    r := gin.Default()
    
    ginstatic.New(r, "./public").
        WithPrefix("/static").
        WithCache(100*1024*1024, 1000).  // 100MB, 1000 文件
        WithGzip(gzip.BestSpeed).
        WithSPA("index.html").
        WithCacheControl("public, max-age=31536000").
        WithETag()
    
    r.Run(":8080")
}
```

## 7. 性能优化措施

| 优化项 | 实现方式 |
|--------|----------|
| 零拷贝 | 使用 `io.Copy` 和 `io.Buffer` |
| 内存缓存 | 热点文件预加载到内存 |
| Gzip 缓存 | 压缩内容缓存，避免重复压缩 |
| Byte Pool | 压缩缓冲区复用 |
| 条件请求 | 支持 304 减少数据传输 |
| 预压缩 | 启动时预压缩静态资源 |

## 8. 目录结构

```
gin-static-server/
├── options.go          # 配置选项
├── engine.go           # 核心引擎
├── handler.go          # 请求处理器
├── cache.go            # 缓存实现
├── middleware.go       # 中间件
├── security.go         # 安全检查
├── compress.go         # Gzip 压缩
├── examples_test.go    # 示例测试
└── README.md           # 文档
```
