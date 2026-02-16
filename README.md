# My Go Tools

Go 工具库集合，提供常用的高性能工具和组件。

## 模块列表

### gin-static-server

基于 Gin 框架的高性能静态文件服务器，支持服务 Vue、React 等前端打包后的静态资源。

**功能特性：**

- 内存缓存（LRO 淘汰策略）
- Gzip/Zstd 压缩
- ETag/Last-Modified 条件请求
- SPA 路由回退支持
- 目录遍历防护
- 链式 API 设计

**快速开始：**

```go
import "github.com/aiqoder/my-go-tools/gin-static-server"

r := gin.Default()

ginstatic.New(r, "./public",
    ginstatic.WithSPA("index.html"),
    ginstatic.WithGzip(6),
)
```

**文档：** [gin-static-server/README.md](gin-static-server/README.md)

---

## 开发指南

本项目使用 Go Workspace 管理多个模块。

### 运行测试

```bash
# 运行所有模块测试
go test ./...

# 运行特定模块测试
cd gin-static-server
go test -v ./...
```

### 添加新模块

1. 在根目录创建子目录，如 `modules/your-module`
2. 在子目录中创建 `go.mod` 文件
3. 更新 `go.work` 文件，添加新模块路径
4. 在 `go.work` 中添加 `use` 路径

## 许可证

MIT License
