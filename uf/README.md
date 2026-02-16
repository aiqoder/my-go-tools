# UF 用户反馈服务客户端

[![Go Version](https://img.shields.io/github/go-mod/go-version/aiqoder/go-tools?label=go)](https://github.com/aiqoder/go-tools)
[![Module](https://img.shields.io/badge/module-github.com/aiqoder/go-tools/uf-blue)](https://pkg.go.dev/github.com/aiqoder/go-tools/uf)

Go 语言客户端，用于对接用户反馈服务 API (`https://uf.yigechengzi.com/`)。

## 特性

- 简洁易用的 API 设计
- 支持自定义 HTTP 客户端配置
- 统一的错误处理机制
- 模块化架构，便于扩展
- 完整的单元测试和示例代码

## 安装

```bash
go get github.com/aiqoder/go-tools/uf
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/aiqoder/go-tools/uf"
)

func main() {
    // 创建默认配置的客户端
    client := uf.NewClient()

    // 获取 BaseURL
    fmt.Printf("API BaseURL: %s\n", client.GetBaseURL())
    // Output: API BaseURL: https://uf.yigechengzi.com/
}
```

## 配置选项

`NewClient` 函数支持以下选项：

```go
// 自定义超时时间（默认 30 秒）
client := uf.NewClient(
    uf.WithTimeout(60 * time.Second),
)

// 自定义 HTTP 客户端
client := uf.NewClient(
    uf.WithHTTPClient(&http.Client{
        Timeout: 60 * time.Second,
    }),
)

// 自定义 BaseURL
client := uf.NewClient(
    uf.WithBaseURL("https://custom.example.com/"),
)
```

## API 参考

### 客户端创建

#### NewClient

```go
func NewClient(opts ...ClientOption) *Client
```

创建 UF 服务客户端。默认 BaseURL 为 `https://uf.yigechengzi.com/`，默认超时时间为 30 秒。

### 客户端方法

#### Get

```go
func (c *Client) Get(path string) ([]byte, error)
```

发起 GET 请求。

#### Post

```go
func (c *Client) Post(path string, data interface{}) ([]byte, error)
```

发起 POST 请求（JSON 格式）。

#### PostForm

```go
func (c *Client) PostForm(path string, data url.Values) ([]byte, error)
```

发起表单 POST 请求。

#### Put

```go
func (c *Client) Put(path string, data interface{}) ([]byte, error)
```

发起 PUT 请求（JSON 格式）。

#### Delete

```go
func (c *Client) Delete(path string) ([]byte, error)
```

发起 DELETE 请求。

#### GetWithContext

```go
func (c *Client) GetWithContext(ctx context.Context, path string) ([]byte, error)
```

发起带上下文的 GET 请求。

#### PostWithContext

```go
func (c *Client) PostWithContext(ctx context.Context, path string, data interface{}) ([]byte, error)
```

发起带上下文的 POST 请求。

#### DoJSONRequest

```go
func (c *Client) DoJSONRequest(method, path string, reqBody, respBody interface{}) error
```

发起通用 JSON 请求，自动处理请求/响应序列化。

#### GetBaseURL

```go
func (c *Client) GetBaseURL() string
```

返回客户端配置的 BaseURL。

#### SetHTTPClient

```go
func (c *Client) SetHTTPClient(client *http.Client)
```

设置自定义 HTTP 客户端。

### 响应类型

#### Response

```go
type Response struct {
    OK    bool   `json:"ok"`
    Error string `json:"error,omitempty"`
}
```

通用响应结构。

### 错误处理

包提供了统一的错误类型 `Error`，包含错误码和错误信息：

```go
// 检查错误类型
if err != nil {
    var ufErr *uf.Error
    if errors.As(err, &ufErr) {
        fmt.Printf("错误码: %s, 错误信息: %s\n", ufErr.Code, ufErr.Message)
    }
}
```

错误码常量：

- `ErrCodeRequestFailed` - 请求失败
- `ErrCodeInvalidResponse` - 响应解析失败
- `ErrCodeTimeout` - 请求超时
- `ErrCodeNetworkError` - 网络错误
- `ErrCodeServerError` - 服务器错误
- `ErrCodeInvalidParams` - 参数错误

## 扩展性设计

该包采用模块化设计，支持扩展对接 `https://uf.yigechengzi.com/` 域名下的多个 API。

### 添加新 API 模块

```go
// 1. 在 client.go 中添加模块字段
type Client struct {
    baseURL    string
    httpClient *http.Client
    // 添加新模块
    activation *ActivationAPI
}

// 2. 创建 API 模块
type ActivationAPI struct {
    client *Client
}

func (a *ActivationAPI) SetClient(c *Client) {
    a.client = c
}

func (a *ActivationAPI) Path() string {
    return "/api/activation"
}
```

## 示例代码

更多示例请参考 `examples_test.go`。

## 许可证

MIT License
