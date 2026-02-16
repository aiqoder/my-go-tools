# OAuth2 登录流程开发文档

## 一、项目概述

本项目采用 **OAuth2 授权码模式**（Authorization Code Flow）进行用户认证，是一个 **Vue3 + Gin** 的全栈应用。项目**不使用JWT**，完全依赖外部OAuth2服务器颁发的访问令牌进行身份验证。

---

## 二、OAuth2 配置

### 2.1 配置定义

配置文件位于 `internal/config/config.go`：

```go
// OAuth2Config OAuth2 配置
type OAuth2Config struct {
    Server       string // OAuth2 服务器地址
    ClientID     string // OAuth2 客户端 ID
    ClientSecret string // OAuth2 客户端密钥
    RedirectURI  string // OAuth2 重定向 URI
}
```

配置加载代码（第122-127行）：

```go
OAuth2: OAuth2Config{
    Server:       getEnv("OAUTH2_SERVER", "http://localhost:8080"),
    ClientID:     getEnv("OAUTH2_CLIENT_ID", ""),
    ClientSecret: getEnv("OAUTH2_CLIENT_SECRET", ""),
    RedirectURI:  getEnv("OAUTH2_REDIRECT_URI", "http://localhost:3000/callback"),
},
```

### 2.2 环境变量

| 环境变量 | 说明 | 示例值 |
|---------|------|-------|
| `OAUTH2_SERVER` | OAuth2 服务器地址 | `http://localhost:8080` |
| `OAUTH2_CLIENT_ID` | 客户端 ID | `your-client-id` |
| `OAUTH2_CLIENT_SECRET` | 客户端密钥 | `your-client-secret` |
| `OAUTH2_REDIRECT_URI` | 回调地址 | `http://localhost:3000/callback` |

### 2.3 环境配置示例

`.env.example` 文件中的配置：

```
# OAuth2 服务器地址
OAUTH2_SERVER=http://localhost:8080
# OAuth2 客户端 ID
OAUTH2_CLIENT_ID=your-client-id
# OAuth2 客户端密钥
OAUTH2_CLIENT_SECRET=your-client-secret
# OAuth2 重定向 URI
OAUTH2_REDIRECT_URI=http://localhost:3000/callback
```

---

## 三、数据类型定义

### 3.1 类型文件

`internal/types/oauth2/oauth2.go`

### 3.2 核心类型

| 类型 | 用途 |
|------|------|
| **TokenResponse** | OAuth2 令牌响应 |
| **UserInfo** | 用户信息 |
| **OAuth2Error** | 错误响应 |

### 3.3 TokenResponse 详解

```go
type TokenResponse struct {
    AccessToken      string `json:"access_token"`       // 访问令牌
    TokenType        string `json:"token_type"`         // 令牌类型（通常为 Bearer）
    ExpiresIn        int64  `json:"expires_in"`         // 访问令牌有效期（秒）
    RefreshToken     string `json:"refresh_token"`      // 刷新令牌
    RefreshExpiresIn int64  `json:"refresh_expires_in"` // 刷新令牌有效期（秒）
    Scope            string `json:"scope,omitempty"`    // 权限范围
}
```

### 3.4 UserInfo 详解

```go
type UserInfo struct {
    Sub       string  `json:"sub"`        // 用户唯一标识
    Username  string  `json:"username"`   // 用户名
    Status    int     `json:"status"`     // 用户状态
    ClientID  string  `json:"client_id"`  // 客户端ID
    ExpireAt  *string `json:"expireAt,omitempty"`  // 过期时间
    IsExpired *bool   `json:"isExpired,omitempty"` // 是否已过期
}
```

---

## 四、后端实现

### 4.1 服务层（Service）

**文件：** `internal/services/oauth2/oauth2.go`

#### 4.1.1 服务初始化

```go
type OAuth2Service struct {
    oauth2Server string
    clientID     string
    clientSecret string
    redirectURI  string
}

func NewOAuth2Service(cfg *config.OAuth2Config) *OAuth2Service {
    return &OAuth2Service{
        oauth2Server: cfg.Server,
        clientID:     cfg.ClientID,
        clientSecret: cfg.ClientSecret,
        redirectURI:  cfg.RedirectURI,
    }
}
```

#### 4.1.2 核心方法

| 方法名 | 功能 | 详细说明 |
|--------|------|----------|
| `ExchangeCodeForToken(code)` | 授权码换令牌 | 使用授权码从 OAuth2 服务器换取访问令牌和刷新令牌 |
| `GetUserInfo(accessToken)` | 获取用户信息 | 通过访问令牌获取用户基本信息 |
| `RefreshToken(refreshToken)` | 刷新令牌 | 使用刷新令牌获取新的访问令牌 |
| `GetConfig()` | 获取配置 | 返回 OAuth2 配置供前端使用 |

#### 4.1.3 ExchangeCodeForToken 详解

```go
// ExchangeCodeForToken 使用授权码换取访问令牌
func (s *OAuth2Service) ExchangeCodeForToken(code string) (*oauth2.TokenResponse, error) {
    tokenURL := s.oauth2Server + "/oauth2/token"

    // 使用表单格式发送请求（OAuth2 标准推荐使用表单格式）
    formData := url.Values{}
    formData.Set("grant_type", "authorization_code")
    formData.Set("code", code)
    formData.Set("client_id", s.clientID)
    formData.Set("client_secret", s.clientSecret)
    formData.Set("redirect_uri", s.redirectURI)

    // 发送 POST 请求并解析响应
    // ...
}
```

#### 4.1.4 GetUserInfo 详解

```go
// GetUserInfo 使用访问令牌获取用户信息
func (s *OAuth2Service) GetUserInfo(accessToken string) (*oauth2.UserInfo, error) {
    url := s.oauth2Server + "/oauth2/userinfo"

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }

    // 设置 Authorization 头
    req.Header.Set("Authorization", "Bearer "+accessToken)

    // 发送请求并解析响应
    // ...
}
```

#### 4.1.5 RefreshToken 详解

```go
// RefreshToken 使用刷新令牌获取新的访问令牌
func (s *OAuth2Service) RefreshToken(refreshToken string) (*oauth2.TokenResponse, error) {
    tokenURL := s.oauth2Server + "/oauth2/token"

    formData := url.Values{}
    formData.Set("grant_type", "refresh_token")
    formData.Set("refresh_token", refreshToken)
    formData.Set("client_id", s.clientID)
    formData.Set("client_secret", s.clientSecret)

    // 发送请求并解析响应
    // ...
}
```

---

### 4.2 处理器层（Handler）

**文件：** `internal/handlers/oauth2/oauth2.go`

#### 4.2.1 路由处理函数

| 路由 | HTTP方法 | 处理函数 | 功能 |
|------|---------|----------|------|
| `/api/oauth2/config` | GET | `GetConfig` | 获取 OAuth2 配置供前端使用 |
| `/api/oauth2/callback` | POST | `Callback` | 处理授权码回调 |
| `/api/oauth2/userinfo` | GET | `GetUserInfo` | 获取用户信息 |
| `/api/oauth2/refresh` | POST | `RefreshToken` | 刷新令牌 |

#### 4.2.2 Callback 处理函数详解

```go
// Callback OAuth2 回调处理
// 这个端点用于接收 OAuth2 授权码，然后由后端服务器端使用授权码换取令牌
func (h *OAuth2Handler) Callback(c *gin.Context) {
    var req struct {
        Code string `json:"code" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":             "invalid_request",
            "error_description": "缺少授权码: " + err.Error(),
        })
        return
    }

    code := req.Code

    // 使用授权码换取令牌
    tokenData, err := h.oauth2Service.ExchangeCodeForToken(code)
    // ... 返回令牌给前端
}
```

#### 4.2.3 GetUserInfo 处理函数详解

```go
// GetUserInfo 获取用户信息（通过访问令牌）
func (h *OAuth2Handler) GetUserInfo(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    // 提取 Bearer token
    token := ""
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        token = authHeader[7:]
    }
    // 从 OAuth2 服务器获取用户信息
    userInfo, err := h.oauth2Service.GetUserInfo(token)
    // ...
}
```

#### 4.2.4 RefreshToken 处理函数详解

```go
// RefreshToken 刷新访问令牌
func (h *OAuth2Handler) RefreshToken(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    // ... 验证请求
    // 使用刷新令牌获取新的访问令牌
    tokenData, err := h.oauth2Service.RefreshToken(req.RefreshToken)
    // ...
}
```

#### 4.2.5 GetConfig 处理函数详解

```go
// GetConfig 获取 OAuth2 配置（返回给前端）
func (h *OAuth2Handler) GetConfig(c *gin.Context) {
    config := h.oauth2Service.GetConfig()
    c.JSON(http.StatusOK, gin.H{
        "oauth_server":    config.Server,
        "client_id":       config.ClientID,
        "redirect_uri":    config.RedirectURI,
    })
}
```

---

### 4.3 路由配置

**文件：** `internal/api/router.go`

```go
// OAuth2 路由（在 /api 下，不在 /api/v1 下）
api := r.Group("/api")
{
    // OAuth2 相关路由
    api.GET("/oauth2/config", oauth2Handler.GetConfig)    // 获取 OAuth2 配置
    api.POST("/oauth2/callback", oauth2Handler.Callback)   // 授权码回调
    api.GET("/oauth2/userinfo", oauth2Handler.GetUserInfo)// 获取用户信息
    api.POST("/oauth2/refresh", oauth2Handler.RefreshToken)// 刷新令牌
}
```

---

### 4.4 服务初始化

**文件：** `main.go`

```go
// 初始化 OAuth2 服务
oauth2Service := oauth2Service.NewOAuth2Service(&cfg.OAuth2)

// ... 其他服务初始化 ...

// 初始化处理器层
oauth2Handler := oauth2.NewOAuth2Handler(oauth2Service)

// 设置路由
router := api.SetupRouter(
    // ...
    oauth2Handler,
)
```

---

## 五、前端实现

### 5.1 核心文件

| 文件 | 功能 |
|------|------|
| `web/src/api/oauth2.ts` | OAuth2 API 调用封装 |
| `web/src/stores/auth.ts` | 认证状态管理（Pinia） |
| `web/src/utils/auth.ts` | State 生成和验证工具 |
| `web/src/pages/Login.vue` | 登录页面 |
| `web/src/pages/Callback.vue` | 回调处理页面 |

---

### 5.2 前端 API 模块

**文件：** `web/src/api/oauth2.ts`

#### 5.2.1 核心函数

| 函数名 | 功能 |
|--------|------|
| `redirectToAuthorize(state)` | 重定向到OAuth2授权页面 |
| `exchangeCodeForToken(code)` | 使用授权码换取令牌（通过后端代理） |
| `refreshToken(refreshToken)` | 刷新令牌（通过后端代理） |
| `getUserInfo(accessToken)` | 获取用户信息（通过后端代理） |
| `introspectToken(token)` | 验证令牌 |

---

### 5.3 认证状态管理

**文件：** `web/src/stores/auth.ts`

#### 5.3.1 核心功能

| 功能 | 说明 |
|------|------|
| `saveTokens()` | 保存令牌到 localStorage |
| `clearTokens()` | 清除令牌 |
| `refreshAccessToken()` | 自动刷新即将过期的令牌 |
| `ensureValidToken()` | 确保令牌有效（提前5分钟自动刷新） |
| `login()` | 登录（重定向到OAuth2授权页面） |
| `logout()` | 登出 |

#### 5.3.2 令牌自动刷新机制

系统会在令牌过期前5分钟自动刷新访问令牌，确保用户操作不会被中断。

---

### 5.4 CSRF 防护工具

**文件：** `web/src/utils/auth.ts`

```typescript
// 生成随机 state 参数
export function generateState(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

// 保存 state 到 sessionStorage
export function saveState(state: string): void {
  sessionStorage.setItem('oauth2_state', state)
}

// 验证 state 参数（防止 CSRF 攻击）
export function validateState(state: string | null): boolean {
  // ... 验证并清除 state
}
```

---

### 5.5 页面组件

#### 5.5.1 登录页面

**文件：** `web/src/pages/Login.vue`

用户点击登录按钮后：
1. 生成 state 参数
2. 保存 state 到 sessionStorage
3. 重定向到 OAuth2 授权服务器

#### 5.5.2 回调页面

**文件：** `web/src/pages/Callback.vue`

OAuth2 授权服务器回调后：
1. 获取 URL 中的 code 和 state 参数
2. 验证 state 是否匹配
3. 将授权码发送给后端换取令牌
4. 保存令牌到 localStorage
5. 跳转到首页

---

## 六、完整认证流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           OAuth2 授权码模式流程                           │
└─────────────────────────────────────────────────────────────────────────┘

1. 用户点击登录
   ┌────────────────────────────────────────────────────────────────────┐
   │  Login.vue -> authStore.login()                                   │
   │    ↓                                                              │
   │  生成 state (CSRF token) → saveState(state)                       │
   │    ↓                                                              │
   │  redirectToAuthorize(state)                                       │
   │    ↓                                                              │
   │  重定向到: {OAuth2Server}/oauth2/authorize                        │
   │           ?client_id=xxx                                          │
   │           &redirect_uri=xxx                                       │
   │           &response_type=code                                     │
   │           &scope=read                                             │
   │           &state=xxx                                              │
   └────────────────────────────────────────────────────────────────────┘
                                      ↓
2. 用户在 OAuth2 服务器授权
   ┌────────────────────────────────────────────────────────────────────┐
   │  OAuth2 服务器显示登录/授权页面                                    │
   │  用户登录并点击"授权"                                             │
   │    ↓                                                              │
   │  OAuth2 服务器重定向到: {RedirectURI}?code=xxx&state=xxx         │
   └────────────────────────────────────────────────────────────────────┘
                                      ↓
3. 回调处理
   ┌────────────────────────────────────────────────────────────────────┐
   │  Callback.vue                                                     │
   │    ↓                                                              │
   │  验证 state (防止 CSRF 攻击)                                      │
   │    ↓                                                              │
   │  POST /api/oauth2/callback { code: "xxx" }                        │
   │    ↓                                                              │
   │  后端 Callback Handler                                            │
   │    ↓                                                              │
   │  POST {OAuth2Server}/oauth2/token                                │
   │       grant_type=authorization_code                               │
   │       code=xxx                                                    │
   │       client_id=xxx                                               │
   │       client_secret=xxx (后端保密)                                │
   │       redirect_uri=xxx                                            │
   │    ↓                                                              │
   │  返回: { access_token, token_type, expires_in,                    │
   │         refresh_token, refresh_expires_in }                       │
   │    ↓                                                              │
   │  前端保存令牌到 localStorage                                      │
   └────────────────────────────────────────────────────────────────────┘
                                      ↓
4. 后续请求
   ┌────────────────────────────────────────────────────────────────────┐
   │  前端每次请求携带: Authorization: Bearer {access_token}           │
   │    ↓                                                              │
   │  后端通过 GET /oauth2/userinfo 验证令牌                           │
   │  (或通过 POST /oauth2/introspect 验证)                           │
   │    ↓                                                              │
   │  返回用户信息                                                     │
   └────────────────────────────────────────────────────────────────────┘
                                      ↓
5. 令牌刷新
   ┌────────────────────────────────────────────────────────────────────┐
   │  当 access_token 即将过期（5分钟前）                              │
   │    ↓                                                              │
   │  POST /api/oauth2/refresh { refresh_token: "xxx" }                │
   │    ↓                                                              │
   │  后端 POST {OAuth2Server}/oauth2/token                           │
   │       grant_type=refresh_token                                   │
   │       refresh_token=xxx                                           │
   │       client_id=xxx                                               │
   │       client_secret=xxx                                           │
   │    ↓                                                              │
   │  返回新的 access_token 和 refresh_token                          │
   └────────────────────────────────────────────────────────────────────┘
```

---

## 七、关键文件汇总

| 类别 | 文件路径 | 描述 |
|------|---------|------|
| **配置** | `internal/config/config.go` | OAuth2 配置结构体定义和加载 |
| **类型** | `internal/types/oauth2/oauth2.go` | TokenResponse、UserInfo、OAuth2Error |
| **服务层** | `internal/services/oauth2/oauth2.go` | OAuth2 核心业务逻辑 |
| **处理器** | `internal/handlers/oauth2/oauth2.go` | HTTP 请求处理 |
| **路由** | `internal/api/router.go` | 路由注册 |
| **前端 API** | `web/src/api/oauth2.ts` | 前端 OAuth2 API 调用 |
| **前端状态** | `web/src/stores/auth.ts` | 前端认证状态管理 |
| **前端工具** | `web/src/utils/auth.ts` | State 生成和验证 |
| **登录页** | `web/src/pages/Login.vue` | 登录页面 |
| **回调页** | `web/src/pages/Callback.vue` | OAuth2 回调处理页面 |

---

## 八、重要说明

### 8.1 关于 JWT

本项目**不生成和使用 JWT**，完全依赖外部 OAuth2 服务器的令牌：

- 访问令牌（access_token）用于 API 请求的身份验证
- 刷新令牌（refresh_token）用于获取新的访问令牌
- 令牌验证通过 OAuth2 服务器的 `/oauth2/userinfo` 或 `/oauth2/introspect` 端点完成

### 8.2 Token 存储

- **访问令牌** 和 **刷新令牌** 存储在前端 `localStorage`
- `client_secret` 仅在后端使用，**不暴露给前端**

### 8.3 安全措施

| 安全措施 | 说明 |
|---------|------|
| **CSRF 防护** | 使用 state 参数防止跨站请求伪造 |
| **密钥保密** | 令牌交换通过后端代理，避免 client_secret 泄露 |
| **自动刷新** | 令牌过期前5分钟自动刷新，保证用户体验 |
| **HTTPS** | 生产环境必须使用 HTTPS 传输 |

### 8.4 与其他系统的关系

项目还有独立的**激活检查中间件**（`internal/api/middleware/activation.go`），用于软件授权验证，与 OAuth2 登录流程是**独立**的，两者可以同时使用。

---

## 九、常见问题

### 9.1 如何配置 OAuth2 服务器？

需要配置一个兼容 OAuth2 协议的授权服务器，支持以下端点：
- `/oauth2/authorize` - 授权端点
- `/oauth2/token` - 令牌端点
- `/oauth2/userinfo` - 用户信息端点
- `/oauth2/introspect` - 令牌验证端点（可选）

### 9.2 如何调试 OAuth2 流程？

1. 检查浏览器控制台的网络请求
2. 确认 OAuth2 服务器日志
3. 验证环境变量配置是否正确
4. 检查 redirect_uri 是否与 OAuth2 服务器配置匹配

### 9.3 令牌过期了怎么办？

前端会自动使用 refresh_token 刷新令牌，如果刷新失败，用户需要重新登录。

---

## 十、扩展阅读

- [OAuth 2.0 授权框架](https://tools.ietf.org/html/rfc6749)
- [OAuth 2.0 安全最佳实践](https://tools.ietf.org/html/rfc8252)
- [Go Gin 框架文档](https://gin-gonic.com/)
- [Vue 3 官方文档](https://vuejs.org/)
