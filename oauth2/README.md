# OAuth2 登录模块

基于 Gin 框架的 OAuth2 授权码模式登录后端实现。

## 功能特性

- 授权码模式（Authorization Code Flow）完整支持
- 授权码换取访问令牌和刷新令牌
- 获取用户信息
- 令牌自动刷新
- 无 JWT，完全依赖外部 OAuth2 服务器
- 完整的单元测试和基准测试

## 安装

```bash
go get github.com/aiqoder/my-go-tools/oauth2
```

## 快速开始

### 1. 创建配置

```go
import "github.com/aiqoder/my-go-tools/oauth2"

cfg := &oauth2.Config{
    Server:       "http://localhost:8080",      // OAuth2 服务器地址（默认 https://uf.yigechengzi.com/）
    ClientID:     "your-client-id",             // 客户端 ID
    ClientSecret: "your-client-secret",         // 客户端密钥
    RedirectURI:  "http://localhost:3000/callback", // 回调地址，通常是前端页面，用于引导用户跳转到系统对应页面，需在 OAuth2 服务提供商处配置
}
// 注意：OAuth2 授权页面需要自行在前端实现，跳转到授权 URL
```

### 2. 创建服务和处理器

```go
// 创建 OAuth2 服务
svc := oauth2.NewOAuth2Service(cfg)

// 创建 HTTP 处理器
handler := oauth2.NewOAuth2Handler(svc)
```

### 3. 注册路由

```go
import "github.com/gin-gonic/gin"

r := gin.Default()

// 注册 OAuth2 路由（/api 前缀）
oauth2.SetupRouter(r, handler)
```

## 环境变量配置

| 环境变量 | 说明 | 默认值 |
|---------|------|-------|
| `OAUTH2_SERVER` | OAuth2 服务器地址 | `http://localhost:8080` |
| `OAUTH2_CLIENT_ID` | 客户端 ID | - |
| `OAUTH2_CLIENT_SECRET` | 客户端密钥 | - |
| `OAUTH2_REDIRECT_URI` | 回调地址 | `http://localhost:3000/callback` |

## API 参考

### 路由列表

| 路由 | 方法 | 功能 |
|------|------|------|
| `/api/oauth2/config` | GET | 获取 OAuth2 配置 |
| `/api/oauth2/authorize` | GET | 构建授权 URL |
| `/api/oauth2/callback` | POST | 处理授权码回调 |
| `/api/oauth2/userinfo` | GET | 获取用户信息 |
| `/api/oauth2/refresh` | POST | 刷新令牌 |

### 获取配置

```bash
GET /api/oauth2/config
```

响应示例：

```json
{
    "oauth_server": "http://localhost:8080",
    "client_id": "your-client-id",
    "redirect_uri": "http://localhost:3000/callback"
}
```

### 授权码回调

```bash
POST /api/oauth2/callback
Content-Type: application/json

{
    "code": "authorization-code"
}
```

响应示例：

```json
{
    "access_token": "xxx",
    "token_type": "Bearer",
    "expires_in": 3600,
    "refresh_token": "xxx",
    "refresh_expires_in": 86400,
    "scope": "read"
}
```

### 获取用户信息

```bash
GET /api/oauth2/userinfo
Authorization: Bearer {access_token}
```

响应示例：

```json
{
    "sub": "user123",
    "username": "john",
    "status": 1,
    "client_id": "app123",
    "expireAt": "2026-12-31T23:59:59Z",
    "isExpired": false,
    "machineCode": "YKEY123456789"
}
```

### 刷新令牌

```bash
POST /api/oauth2/refresh
Content-Type: application/json

{
    "refresh_token": "xxx"
}
```

## 使用示例

### 完整示例

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/aiqoder/my-go-tools/oauth2"
)

func main() {
    // 创建配置
    // 注意：Server 默认为 https://uf.yigechengzi.com/，如需自定义请设置
    // 注意：RedirectURI 需要自行在 OAuth2 服务提供商处配置回调地址
    // 注意：OAuth2 授权页面需要自行在前端实现，跳转到授权 URL
    cfg := &oauth2.Config{
        Server:       "http://localhost:8080",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURI:  "http://localhost:3000/callback",
    }

    // 创建服务和处理器
    svc := oauth2.NewOAuth2Service(cfg)
    handler := oauth2.NewOAuth2Handler(svc)

    // 创建 Gin 引擎
    r := gin.Default()

    // 注册路由
    oauth2.SetupRouter(r, handler)

    // 启动服务器
    log.Fatal(r.Run(":8080"))
}
```

### 使用自定义选项

```go
import "time"

// 自定义 HTTP 客户端
client := &http.Client{
    Timeout: 60 * time.Second,
}

svc := oauth2.NewOAuth2Service(cfg, oauth2.WithHTTPClient(client))
```

## 前端集成

### 1. 重定向到授权页面

```javascript
// 生成 state
const state = generateRandomState();

// 保存 state 到 sessionStorage
sessionStorage.setItem('oauth2_state', state);

// 跳转到授权页面
const config = await fetch('/api/oauth2/config').then(r => r.json());
const authorizeUrl = `${config.oauth_server}/oauth2/authorize?` +
    `client_id=${config.client_id}&` +
    `redirect_uri=${encodeURIComponent(config.redirect_uri)}&` +
    `response_type=code&` +
    `state=${state}&` +
    `scope=read`;

window.location.href = authorizeUrl;
```

### 2. 处理回调

```javascript
// 从 URL 获取授权码
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');
const state = urlParams.get('state');

// 验证 state
const savedState = sessionStorage.getItem('oauth2_state');
if (state !== savedState) {
    throw new Error('State 不匹配');
}

// 发送给后端换取令牌
const tokenResponse = await fetch('/api/oauth2/callback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code })
}).then(r => r.json());

// 保存令牌
localStorage.setItem('access_token', tokenResponse.access_token);
localStorage.setItem('refresh_token', tokenResponse.refresh_token);
```

### 3. 获取用户信息

```javascript
const accessToken = localStorage.getItem('access_token');

const userInfo = await fetch('/api/oauth2/userinfo', {
    headers: { 'Authorization': `Bearer ${accessToken}` }
}).then(r => r.json());

console.log('用户:', userInfo);
```

### 4. 刷新令牌

```javascript
const refreshToken = localStorage.getItem('refresh_token');

const tokenResponse = await fetch('/api/oauth2/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken })
}).then(r => r.json());

// 更新保存的令牌
localStorage.setItem('access_token', tokenResponse.access_token);
localStorage.setItem('refresh_token', tokenResponse.refresh_token);
```

## 测试

```bash
# 运行单元测试
go test -v ./...

# 运行基准测试
go test -bench=. ./...

# 查看测试覆盖率
go test -cover ./...
```

## 性能基准

参考 `oauth2_bench_test.go` 中的基准测试结果。

## 目录结构

```
oauth2/
├── types.go           # 数据类型定义
├── service.go         # 服务层实现
├── handler.go         # HTTP 处理器
├── router.go          # 路由配置
├── oauth2_test.go     # 单元测试
├── oauth2_bench_test.go # 基准测试
├── examples_test.go  # 使用示例
└── README.md          # 本文档
```

## 安全说明

- `client_secret` 仅存在于服务端内存，不通过网络传输
- 令牌交换使用 POST 方法，防止敏感信息暴露在 URL
- 建议生产环境使用 HTTPS 传输

## 许可证

MIT License
