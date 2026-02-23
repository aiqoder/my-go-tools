# 静态文件中间件任务清单 (Tasks.md)

## 任务列表

### Task 1: 定义 MiddlewareConfig 配置结构

**验收标准**:
- [x] 定义 `MiddlewareConfig` 结构体，包含 Root、Prefix、EmbedFS、StaticExts 等字段
- [x] 定义 `MiddlewareOption` 函数类型
- [x] 实现默认配置函数 `defaultMiddlewareConfig()`

**实现位置**: `middleware.go`

---

### Task 2: 实现静态资源扩展名检测逻辑

**验收标准**:
- [x] 定义默认静态资源扩展名列表 `defaultStaticExts`
- [x] 实现扩展名匹配函数 `isStaticFile(path string, exts []string) bool`
- [x] 支持大小写不敏感匹配

**实现位置**: `middleware.go`

---

### Task 3: 实现 Middleware 中间件结构体

**验收标准**:
- [x] 定义 `StaticFileMiddleware` 结构体
- [x] 实现 `ServeHTTP` 方法（gin.HandlerFunc 接口）

**实现位置**: `middleware.go`

---

### Task 4: 实现中间件核心处理逻辑

**验收标准**:
- [x] 实现路径提取（移除 URL 前缀）
- [x] 实现安全检查（目录遍历防护）
- [x] 实现文件查找（支持 OS 文件系统和 embed.FS）
- [x] 实现文件不存在时的 `c.Next()` 传递

**实现位置**: `middleware.go`

---

### Task 5: 实现文件响应处理

**验收标准**:
- [x] 设置正确的 Content-Type（MIME 类型）
- [x] 设置 Last-Modified 响应头
- [x] 支持 ETag（可选）
- [x] 支持 Gzip 压缩（可选）
- [x] 支持条件请求（If-None-Match、If-Modified-Since）

**实现位置**: `middleware.go`

---

### Task 6: 实现工厂函数和配置选项

**验收标准**:
- [x] 实现 `NewStaticFileMiddleware(root string, opts ...MiddlewareOption) gin.HandlerFunc`
- [x] 实现 `NewStaticFileMiddlewareWithConfig(cfg *MiddlewareConfig) gin.HandlerFunc`
- [x] 实现常用配置选项函数（WithPrefix、WithStaticExts、WithEmbedFS 等）

**实现位置**: `middleware.go`

---

### Task 7: 编写单元测试

**验收标准**:
- [x] 测试扩展名匹配逻辑
- [x] 测试中间件基本功能（文件存在返回 200）
- [x] 测试中间件放行功能（文件不存在调用 Next）
- [x] 测试目录遍历防护
- [x] 测试条件请求
- [x] 测试 Gzip 压缩

**实现位置**: `middleware_test.go`

---

### Task 8: 更新 README.md 文档

**验收标准**:
- [ ] 添加中间件使用方式的说明
- [ ] 提供使用示例代码
- [ ] 说明与路由方式的区别

**实现位置**: `README.md`

---

## 任务依赖关系

```
Task 1 ──┬── Task 2 ──┬── Task 3 ──┬── Task 4 ──┬── Task 5 ──┬── Task 6 ──► Task 7
         │            │            │            │            │
         │            │            │            │            │
         └────────────┴────────────┴────────────┴────────────┘
                                                      │
                                                      ▼
                                                   Task 8
```

## 实施顺序

1. 首先完成 Task 1-2（配置和基础函数）
2. 然后完成 Task 3-6（核心实现）
3. 最后完成 Task 7-8（测试和文档）
