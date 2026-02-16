# UF 用户反馈服务 - 设计文档

## 1. 架构设计

### 1.1 整体架构

```
github.com/aiqoder/go-tools/uf/
├── client.go          # 客户端核心实现
├── config.go         # 配置管理
├── types.go          # 类型定义
├── errors.go         # 错误定义
├── examples_test.go  # 使用示例
├── go.mod            # 模块定义
└── README.md         # 包文档
```

### 1.2 设计模式

采用**工厂模式 + 选项模式**：
- `NewClient()` 创建客户端实例
- 通过 `ClientOption` 选项函数自定义配置

### 1.3 模块划分

| 模块 | 职责 |
|------|------|
| `config` | 配置管理（BaseURL、超时设置等） |
| `client` | HTTP 客户端封装、公共请求逻辑 |
| `types` | 请求/响应类型定义 |
| `errors` | 统一错误处理 |

## 2. 数据结构设计

### 2.1 配置结构

```go
// Config 客户端配置
type Config struct {
    BaseURL    string        // API 基础地址（固定为 https://uf.yigechengzi.com/）
    Timeout    time.Duration // 请求超时时间
    HTTPClient *http.Client  // 自定义 HTTP 客户端
}
```

### 2.2 客户端结构

```go
// Client UF API 客户端
//
// 提供对用户反馈服务的统一访问入口
type Client struct {
    baseURL    string
    httpClient *http.Client
    // 可扩展：添加不同 API 模块
    // activation *ActivationAPI
    // feedback  *FeedbackAPI
}
```

### 2.3 响应结构

参考 `ACTIVATION_CHECK_API.md` 的响应格式：

```go
// Response 通用响应结构
type Response struct {
    OK    bool   `json:"ok"`
    Error string `json:"error,omitempty"`
}
```

## 3. 接口设计

### 3.1 客户端创建

```go
// NewClient 创建 UF 服务客户端
//
// 默认 BaseURL 为 https://uf.yigechengzi.com/
// 可通过选项自定义配置
func NewClient(opts ...ClientOption) *Client

// ClientOption 客户端配置选项
type ClientOption func(*Client)
```

### 3.2 公共方法

```go
// GetBaseURL 返回客户端配置的 BaseURL
func (c *Client) GetBaseURL() string

// SetHTTPClient 设置自定义 HTTP 客户端
func (c *Client) SetHTTPClient(client *http.Client)

// doRequest 发起 HTTP 请求（内部方法）
func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error)
```

## 4. 扩展性设计

### 4.1 模块化 API 注册

为支持将来对接该域名下的多个 API，采用模块化设计：

```go
// APIModule API 模块接口
//
// 不同的 API 功能实现此接口
type APIModule interface {
    // SetClient 设置关联的客户端
    SetClient(c *Client)
    // Path 返回 API 路径前缀
    Path() string
}
```

### 4.2 预留扩展点

```go
// Client 扩展字段（注释说明扩展方式）
type Client struct {
    // ... 现有字段
    
    // 以下为预留扩展字段
    // activation *ActivationAPI // 激活检测模块
    // feedback  *FeedbackAPI    // 反馈服务模块
}
```

## 5. 错误处理设计

### 5.1 错误类型

```go
// Error 错误结构
type Error struct {
    Code    string // 错误码
    Message string // 错误信息
    Err     error  // 原始错误
}

// Error 实现 error 接口
func (e *Error) Error() string

// Unwrap 返回原始错误
func (e *Error) Unwrap() error
```

### 5.2 错误处理策略

- 网络错误：返回包含原始错误的包装错误
- API 错误：解析响应中的 error 字段
- 超时错误：明确区分超时类型

## 6. 代码风格规范

### 6.1 注释要求

- 包注释：说明包用途
- 公共函数注释：遵循 go doc 标准，包含参数、返回值说明
- 示例代码：包含实际可运行的示例

### 6.2 参考风格

参考 `ACTIVATION_CHECK_API.md` 和 `oauth2` 包：
- 使用中文注释
- 清晰的错误信息
- 完善的 go doc

## 7. 技术决策

### 7.1 技术选型

| 决策项 | 选择 | 理由 |
|--------|------|------|
| HTTP 库 | 标准库 net/http | 无外部依赖，保持轻量 |
| JSON 解析 | 标准库 encoding/json | 无外部依赖 |
| 超时控制 | 通过 http.Client.Timeout | 简单可靠 |

### 7.2 配置策略

- BaseURL 固定写入代码（`https://uf.yigechengzi.com/`）
- 超时时间可配置（默认 30 秒）
- 支持自定义 HTTP 客户端

## 8. 测试策略

### 8.1 单元测试
- 客户端创建测试
- 配置选项测试
- 错误处理测试

### 8.2 示例测试
- 使用 `examples_test.go` 展示典型用法

## 9. 实现计划

### Phase 1: 基础框架
1. 创建 `go.mod`
2. 实现配置结构
3. 实现客户端创建

### Phase 2: 核心功能
1. 实现 HTTP 请求封装
2. 实现错误处理
3. 添加单元测试

### Phase 3: 完善
1. 添加示例代码
2. 添加 README 文档
3. 代码格式化检查
