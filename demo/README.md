# Gin Static Server Demo

使用 gin-static-server 托管 Vue 前端项目的完整示例。

## 项目结构

```
demo/
├── frontend/           # Vue 前端项目
│   ├── src/          # 源代码
│   ├── dist/        # 构建产物（静态文件）
│   └── ...
├── server/           # Go 后端服务
│   ├── main.go      # 主程序
│   ├── server       # 编译后的可执行文件
│   └── go.mod
├── start.sh         # 启动脚本
├── build.sh         # 构建脚本
└── README.md        # 本文件
```

## 快速开始

### 方式一：直接运行（已构建）

```bash
cd demo
./start.sh
```

服务将在 http://localhost:8080 启动。

### 方式二：重新构建

```bash
cd demo
./build.sh
./start.sh
```

### 方式三：分别启动前后端

#### 1. 启动后端

```bash
cd demo/server
./server
```

#### 2. 开发模式（前端热重载）

```bash
cd demo/frontend
npm run dev
```

前端开发服务器会启动在 http://localhost:5173，后端服务在 http://localhost:8080。

## 功能特性

- ✅ SPA 路由支持（Vue Router）
- ✅ Gzip 压缩
- ✅ 内存缓存
- ✅ ETag/Last-Modified
- ✅ 目录遍历防护
- ✅ 点文件隐藏

## API 使用示例

gin-static-server 配置了以下特性：

```go
ginstatic.New(r, "../frontend/dist",
    ginstatic.WithPrefix(""),                    // 根路径访问
    ginstatic.WithSPA("index.html"),             // SPA 回退
    ginstatic.WithGzip(6),                      // Gzip 级别 6
    ginstatic.WithCache(50*1024*1024, 100),     // 50MB 缓存
    ginstatic.WithCacheControl("public, max-age=31536000"),
    ginstatic.WithETag(),
)
```

## 更多信息

- [gin-static-server 文档](../../gin-static-server/README.md)
- [Vue 文档](https://vuejs.org)
- [Vite 文档](https://vitejs.dev)
