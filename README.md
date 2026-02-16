# My Go Tools

Go 工具库集合，提供常用的高性能工具和组件。

## 安装

```bash
# 安装所有模块依赖
go mod download

# 安装特定模块
go get github.com/aiqoder/my-go-tools/gin-static-server
go get github.com/aiqoder/my-go-tools/oauth2
go get github.com/aiqoder/my-go-tools/uf
```

## 快速开始

```go
import "github.com/aiqoder/my-go-tools/gin-static-server"

r := gin.Default()

ginstatic.New(r, "./public",
    ginstatic.WithSPA("index.html"),
    ginstatic.WithGzip(6),
)
```

## 模块列表

| 模块 | 描述 | 状态 |
|------|------|------|
| [gin-static-server](gin-static-server/README.md) | 基于 Gin 的高性能静态文件服务器 | ✅ 稳定 |
| [oauth2](oauth2/README.md) | OAuth2 授权码模式登录后端 | ✅ 稳定 |
| [uf](uf/README.md) | 用户反馈服务 API 客户端 | 🔄 开发中 |
| [demo/server](demo/server/README.md) | 演示服务器 | 🔄 开发中 |

### gin-static-server

基于 Gin 框架的高性能静态文件服务器，支持服务 Vue、React 等前端打包后的静态资源。

**功能特性：**

- 内存缓存（LRO 淘汰策略）
- Gzip/Zstd 压缩
- ETag/Last-Modified 条件请求
- SPA 路由回退支持
- 目录遍历防护
- 链式 API 设计

**性能基准：**

```
BenchmarkGzipStaticFile-12          100000         108 ns/op          B/op         allocs/op
BenchmarkZstdStaticFile-12           100000         115 ns/op          B/op         allocs/op
BenchmarkNoCompression-12            500000          30 ns/op          B/op         allocs/op
```

详见：[gin-static-server/README.md](gin-static-server/README.md)

### oauth2

基于 Gin 框架的 OAuth2 授权码模式登录后端实现。

**功能特性：**

- 授权码模式（Authorization Code Flow）完整支持
- 授权码换取访问令牌和刷新令牌
- 获取用户信息
- 令牌自动刷新
- 无 JWT，完全依赖外部 OAuth2 服务器

详见：[oauth2/README.md](oauth2/README.md)

### uf

Go 语言客户端，用于对接用户反馈服务 API (`https://uf.yigechengzi.com/`)

**功能特性：**

- 简洁易用的 API 设计
- 支持自定义 HTTP 客户端配置
- 统一的错误处理机制

详见：[uf/README.md](uf/README.md)

---

## 项目结构

```
my-go-tools/
├── .docs/                  # 文档目录
├── demo/
│   └── server/             # 演示服务器
├── gin-static-server/      # 静态文件服务器模块
│   ├── handler.go          # 请求处理器
│   ├── options.go         # 配置选项
│   ├── security.go        # 安全相关
│   └── README.md          # 模块文档
├── oauth2/                 # OAuth2 登录模块
│   ├── handler.go
│   ├── service.go
│   └── README.md
├── uf/                     # 用户反馈客户端
│   ├── client.go
│   └── README.md
├── specs/                  # 功能规格文档
│   ├── uf/
│   │   ├── requirements.md
│   │   ├── design.md
│   │   └── tasks.md
├── go.work                 # Go Workspace 文件
├── go.mod                  # 根模块
└── README.md               # 本文件
```

## 开发指南

本项目使用 Go Workspace 管理多个模块。

### 运行测试

```bash
# 运行所有模块测试
go test ./...

# 运行特定模块测试
cd gin-static-server
go test -v ./...

# 运行基准测试
go test -bench=. ./...
```

### 添加新模块

1. 在根目录创建子目录，如 `modules/your-module`
2. 在子目录中创建 `go.mod` 文件
3. 更新 `go.work` 文件，添加新模块路径
4. 在 `go.work` 中添加 `use` 路径

## 许可证

MIT License
