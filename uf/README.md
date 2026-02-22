# UF 用户反馈服务客户端

[![Go Version](https://img.shields.io/github/go-mod/go-version/aiqoder/my-go-tools?label=go)](https://github.com/aiqoder/my-go-tools)
[![Module](https://img.shields.io/badge/module-github.com/aiqoder/my-go-tools/uf-blue)](https://pkg.go.dev/github.com/aiqoder/my-go-tools/uf)

Go 语言客户端，用于对接用户反馈服务 API (`https://uf.yigechengzi.com/`)。

## 特性

- 简洁易用的 API 设计
- 支持自定义 HTTP 客户端配置
- 统一的错误处理机制
- 完整的单元测试和示例代码

## 安装

```bash
go get github.com/aiqoder/my-go-tools/uf
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/aiqoder/my-go-tools/uf"
)

func main() {
    client := uf.NewClient()

    // 记录软件活跃度
    resp, err := client.RecordActivity(1)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    fmt.Printf("活跃度记录: %v\n", resp.IsOK())

    // 检查软件激活状态
    resp2, err := client.CheckActivation(1, "ABC-123-XYZ")
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }
    fmt.Printf("已激活: %v\n", resp2.Activated)
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

### NewClient

```go
func NewClient(opts ...ClientOption) *Client
```

创建 UF 服务客户端。默认 BaseURL 为 `https://uf.yigechengzi.com/`，默认超时时间为 30 秒。

### RecordActivity

```go
func (c *Client) RecordActivity(softwareId uint) (*ActivityResponse, error)
```

记录软件活跃度。

参数：
- `softwareId` - 软件 ID

返回：
- `*ActivityResponse` - 活跃度记录响应
- `error` - 错误信息

### CheckActivation

```go
func (c *Client) CheckActivation(softwareId uint, machineCode string) (*ActivationCheckResponse, error)
```

检查软件激活状态。

参数：
- `softwareId` - 软件 ID
- `machineCode` - 机器码

返回：
- `*ActivationCheckResponse` - 激活检查响应
- `error` - 错误信息

### 响应类型

#### ActivityResponse

```go
type ActivityResponse struct {
    OK    bool   `json:"ok"`
    ID    uint   `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}
```

活跃度记录响应。

#### ActivationCheckResponse

```go
type ActivationCheckResponse struct {
    OK        bool   `json:"ok"`
    Activated bool   `json:"activated"`
    ExpireAt string `json:"expireAt,omitempty"`
    Error     string `json:"error,omitempty"`
}
```

激活检查响应。

### 错误处理

包提供了统一的错误类型 `Error`，包含错误码和错误信息：

```go
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

## 示例代码

更多示例请参考 `examples_test.go`。

## 许可证

MIT License
