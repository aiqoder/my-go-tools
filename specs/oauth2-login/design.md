# OAuth2 登录模块设计方案

## 1. 架构概述

本模块采用分层架构设计，分为以下层次：

- **配置层**: OAuth2 配置定义和加载
- **类型层**: 数据类型定义
- **服务层**: OAuth2 核心业务逻辑
- **处理器层**: HTTP 请求处理
- **路由层**: API 路由注册

## 2. 目录结构

```
oauth2/
├── config/
│   └── config.go          # OAuth2 配置结构体
├── types/
│   └── oauth2.go          # 数据类型定义
├── services/
│   └── oauth2.go          # 服务层实现
├── handlers/
│   └── oauth2.go          # 处理器层实现
└── router.go              # 路由配置
```

## 3. 核心设计

### 3.1 配置设计

**OAuth2Config** 定义：

```go
type OAuth2Config struct {
    Server       string // OAuth2 服务器地址
    ClientID     string // OAuth2 客户端 ID
    ClientSecret string // OAuth2 客户端密钥
    RedirectURI  string // OAuth2 重定向 URI
}
```

配置加载使用环境变量，遵循 12-Factor App 原则。

### 3.2 服务层设计

**OAuth2Service** 结构：

```go
type OAuth2Service struct {
    oauth2Server string
    clientID     string
    clientSecret string
    redirectURI  string
}
```

核心方法：

| 方法 | 功能 | 技术方案 |
|------|------|----------|
| `ExchangeCodeForToken` | 授权码换令牌 | POST 表单请求到 `/oauth2/token` |
| `GetUserInfo` | 获取用户信息 | GET 请求 + Bearer Token |
| `RefreshToken` | 刷新令牌 | POST 表单请求到 `/oauth2/token` |
| `GetConfig` | 获取公开配置 | 返回不含密钥的配置 |

### 3.3 处理器层设计

**OAuth2Handler** 结构：

```go
type OAuth2Handler struct {
    oauth2Service *OAuth2Service
}
```

路由处理函数：

| 函数 | 路由 | 功能 |
|------|------|------|
| `GetConfig` | GET /api/oauth2/config | 获取 OAuth2 配置 |
| `Callback` | POST /api/oauth2/callback | 处理授权码回调 |
| `GetUserInfo` | GET /api/oauth2/userinfo | 获取用户信息 |
| `RefreshToken` | POST /api/oauth2/refresh | 刷新令牌 |

### 3.4 错误处理设计

统一错误响应格式：

```go
type OAuth2Error struct {
    Error             string `json:"error"`
    ErrorDescription  string `json:"error_description,omitempty"`
}
```

HTTP 状态码使用：

- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 令牌无效或过期
- `500 Internal Server Error`: 服务器内部错误

## 4. 数据流设计

### 4.1 授权码换令牌流程

```
前端                    后端                      OAuth2服务器
  |                       |                         |
  |---POST /callback----->|                         |
  |   {code: "xxx"}       |                         |
  |                       |---POST /oauth2/token--->|
  |                       |   grant_type=authorization_code
  |                       |   code=xxx             |
  |                       |   client_id=xxx         |
  |                       |   client_secret=xxx     |
  |                       |   redirect_uri=xxx     |
  |                       |<--token response-------|
  |<--{token}-------------|                         |
```

### 4.2 用户信息获取流程

```
前端                    后端                      OAuth2服务器
  |                       |                         |
  |---GET /userinfo------>|                         |
  |   Authorization:     |                         |
  |   Bearer {token}     |                         |
  |                       |---GET /oauth2/userinfo->|
  |                       |   Authorization: Bearer |
  |                       |<--user info------------|
  |<--{user info}--------|                         |
```

## 5. API 响应格式

### 5.1 令牌响应

```json
{
    "access_token": "xxx",
    "token_type": "Bearer",
    "expires_in": 7200,
    "refresh_token": "xxx",
    "refresh_expires_in": 604800,
    "scope": "read"
}
```

### 5.2 用户信息响应

```json
{
    "sub": "user123",
    "username": "john",
    "status": 1,
    "client_id": "app123",
    "expireAt": "2026-02-17T00:00:00Z",
    "isExpired": false
}
```

### 5.3 错误响应

```json
{
    "error": "invalid_request",
    "error_description": "缺少授权码: code is required"
}
```

## 6. 依赖设计

本模块依赖以下标准库：

- `net/http`: HTTP 客户端和服务器
- `net/url`: URL 解析和编码
- `encoding/json`: JSON 编解码
- `fmt`: 格式化输出
- `errors`: 错误处理

无外部第三方依赖，遵循最小依赖原则。

## 7. 扩展性设计

### 7.1 接口抽象

服务层可通过接口抽象，便于单元测试：

```go
type OAuth2ServiceInterface interface {
    ExchangeCodeForToken(code string) (*TokenResponse, error)
    GetUserInfo(accessToken string) (*UserInfo, error)
    RefreshToken(refreshToken string) (*TokenResponse, error)
    GetConfig() *OAuth2Config
}
```

### 7.2 中间件支持

可扩展实现认证中间件：

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证令牌逻辑
    }
}
```

## 8. 安全考虑

- client_secret 仅存在于服务端内存，不持久化存储
- 令牌交换使用 POST 方法防止敏感信息暴露在 URL
- 使用 Form 编码而非 JSON 编码（OAuth2 RFC 推荐）
- 错误信息不泄露内部实现细节
