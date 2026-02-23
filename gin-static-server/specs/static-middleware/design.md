# 静态文件中间件设计规格 (Design.md)

## 1. 架构设计

### 1.1 设计目标

将静态文件服务从路由处理方式改为 Gin 中间件方式，解决路由冲突问题。设计原则：

1. **最小干预**: 只拦截以静态资源扩展名结尾的请求
2. **复用性**: 尽可能复用现有 `StaticEngine` 的核心逻辑
3. **灵活性**: 支持配置化的扩展名匹配规则
4. **一致性**: 与现有 API 风格保持一致

### 1.2 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Gin Engine                            │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│  │   API 路由   │    │   其他路由   │    │ StaticFile  │      │
│  │  /api/v1/*  │    │   /docs/*   │    │  Middleware │      │
│  └─────────────┘    └─────────────┘    └─────────────┘      │
│                                                 │             │
│                                    ┌────────────┴────────┐   │
│                                    │  请求拦截逻辑         │   │
│                                    │  1. 检查扩展名       │   │
│                                    │  2. 查找物理文件     │   │
│                                    │  3. 返回文件/Next() │   │
│                                    └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 模块设计

新增 `middleware.go` 文件，包含以下核心组件：

```go
// StaticFileMiddleware 静态文件中间件
type StaticFileMiddleware struct {
    config *MiddlewareConfig
    engine *StaticEngine  // 复用 StaticEngine
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
    Root            string        // 静态文件根目录
    Prefix          string        // URL 路径前缀
    EmbedFS         any           // embed.FS
    EmbedRoot       string        // embed.FS 根目录
    StaticExts      []string      // 需要拦截的静态资源扩展名
    EnableCache     bool
    EnableGzip      bool
    UseETag         bool
    // ... 其他配置
}
```

## 2. 核心逻辑设计

### 2.1 中间件处理流程

```
请求进入
    │
    ▼
┌─────────────────┐
│ 检查路径扩展名   │── 否 ──▶ c.Next() ──▶ 继续传递
└────────┬────────┘
         │ 是
         ▼
┌─────────────────┐
│ 移除前缀路径    │  例如: /static/app.js → /app.js
└────────┬────────┘
         ▼
┌─────────────────┐
│ 安全检查        │── 失败 ──▶ 403 Forbidden
│ (目录遍历)      │
└────────┬────────┘
         │ 通过
         ▼
┌─────────────────┐
│ 查找物理文件    │── 不存在 ──▶ c.Next() ──▶ 继续传递
└────────┬────────┘
         │ 存在
         ▼
┌─────────────────┐
│ 返回静态文件    │  设置响应头、压缩、缓存
└─────────────────┘
    │
    ▼
 请求结束
```

### 2.2 扩展名匹配策略

默认拦截的静态资源扩展名：

```go
var defaultStaticExts = []string{
    ".html", ".htm",    // HTML 文件
    ".js", ".mjs",      // JavaScript 文件
    ".css", ".scss",    // 样式文件
    ".json",            // JSON 文件
    ".map",             // Source Map
    ".woff", ".woff2",  // 字体
    ".ttf", ".eot", ".otf",
    ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp", // 图片
    ".wasm",            // WebAssembly
}
```

### 2.3 路径处理逻辑

1. **前缀移除**: 当配置 `Prefix: /static` 时
   - 请求 `/static/js/app.js` → 映射到 `./public/js/app.js`

2. **安全检查**: 复用 `IsPathTraversal` 函数，防止 `../../etc/passwd` 攻击

3. **文件不存在处理**: 调用 `c.Next()` 让请求继续传递，不阻断正常 API 路由

## 3. API 设计

### 3.1 工厂函数

```go
// NewStaticFileMiddleware 创建静态文件中间件
// root: 静态文件根目录
// opts: 配置选项
func NewStaticFileMiddleware(root string, opts ...MiddlewareOption) gin.HandlerFunc

// NewStaticFileMiddlewareWithConfig 使用配置创建中间件
func NewStaticFileMiddlewareWithConfig(cfg *MiddlewareConfig) gin.HandlerFunc
```

### 3.2 配置选项

```go
// WithStaticExts 自定义静态资源扩展名
func WithStaticExts(exts []string) MiddlewareOption

// WithPrefix 设置 URL 前缀
func WithPrefix(prefix string) MiddlewareOption

// WithEmbedFS 使用 embed.FS
func WithEmbedFS(fs any, root string) MiddlewareOption

// WithCache 启用缓存
func WithCache(maxSize int64, maxFiles int) MiddlewareOption

// WithGzip 启用 Gzip
func WithGzip(level int) MiddlewareOption
```

### 3.3 使用示例

```go
r := gin.Default()

// 注册中间件（解决路由冲突）
r.Use(ginstatic.StaticFileMiddleware("./public"))

// 定义 API 路由（不再被静态文件拦截）
r.GET("/api/v1/users", handler.GetUsers)

// 其他路由正常
r.GET("/health", handler.Health)
```

## 4. 复用设计

为减少代码重复，中间件内部创建 `StaticEngine` 实例，复用其核心逻辑：

1. **文件读取**: 复用 `getFile()` 方法
2. **压缩处理**: 复用 `getCompressedData()` 方法
3. **ETag/缓存**: 复用 `checkNotModified()` 方法
4. **MIME 类型**: 复用 `GetMimeType()` 函数

具体实现方式：

```go
func (m *StaticFileMiddleware) serveFile(c *gin.Context) {
    // 提取路径，调用 StaticEngine 的方法处理
    // 由于 StaticEngine 是为路由设计的，需要适配中间件场景
    // 方案：内部使用简化版的文件处理逻辑
}
```

## 5. 安全设计

### 5.1 目录遍历防护

复用现有 `IsPathTraversal()` 函数：

```go
safe, cleanPath := IsPathTraversal(root, requestPath)
if !safe {
    c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
    c.Abort()
    return
}
```

### 5.2 隐藏文件

可选配置 `HideDotFiles: true`，拒绝访问 `.` 开头的文件。

## 6. 文件结构

```
gin-static-server/
├── handler.go        # StaticEngine 实现（现有）
├── options.go        # 配置选项（现有）
├── middleware.go    # 新增：中间件实现
├── middleware_test.go # 新增：中间件测试
└── specs/
    └── static-middleware/
        ├── requirements.md
        ├── design.md
        └── tasks.md
```

## 7. 兼容性考虑

1. **向后兼容**: 保留现有 `StaticEngine` 的所有功能
2. **渐进式迁移**: 用户可选择使用中间件或继续使用路由方式
3. **文档更新**: 更新 README.md 说明两种使用方式
