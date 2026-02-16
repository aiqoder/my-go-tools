# OAuth2 登录模块需求规格

## 1. 概述

本规格定义了 Gin 框架下 OAuth2 授权码模式登录的后端实现需求。项目采用 Vue3 + Gin 全栈架构，完全依赖外部 OAuth2 服务器颁发的访问令牌进行身份验证，**不使用 JWT**。

## 2. 功能需求

### 2.1 OAuth2 配置管理

**需求 R1**: 系统必须支持通过环境变量配置 OAuth2 服务器参数

- `OAUTH2_SERVER`: OAuth2 服务器地址（默认 `http://localhost:8080`）
- `OAUTH2_CLIENT_ID`: 客户端应用 ID
- `OAUTH2_CLIENT_SECRET`: 客户端密钥（仅后端使用，不暴露）
- `OAUTH2_REDIRECT_URI`: 授权回调地址

**需求 R2**: 系统必须提供 API 端点供前端获取公开的 OAuth2 配置（不含密钥）

- 返回 `oauth_server`、`client_id`、`redirect_uri`
- 隐藏 `client_secret`

### 2.2 授权码换令牌

**需求 R3**: 系统必须实现授权码回调处理功能

- 接收前端传来的授权码（code）
- 使用授权码向 OAuth2 服务器换取访问令牌和刷新令牌
- 返回令牌信息给前端

**需求 R4**: 令牌交换必须通过后端代理，保护 client_secret 不泄露

### 2.3 用户信息获取

**需求 R5**: 系统必须提供用户信息查询接口

- 接收前端传来的访问令牌
- 向 OAuth2 服务器请求用户信息
- 返回用户基本信息（sub、username、status 等）

### 2.4 令牌刷新

**需求 R6**: 系统必须支持刷新令牌功能

- 接收前端传来的刷新令牌
- 向 OAuth2 服务器换取新的访问令牌
- 返回新令牌信息

### 2.5 路由规范

**需求 R7**: OAuth2 API 路由必须符合以下规范

| 路由 | HTTP 方法 | 功能 |
|------|-----------|------|
| `/api/oauth2/config` | GET | 获取 OAuth2 配置 |
| `/api/oauth2/callback` | POST | 处理授权码回调 |
| `/api/oauth2/userinfo` | GET | 获取用户信息 |
| `/api/oauth2/refresh` | POST | 刷新令牌 |

## 3. 数据类型需求

### 3.1 TokenResponse

系统必须支持以下令牌响应结构：

- `access_token`: 访问令牌（字符串）
- `token_type`: 令牌类型（通常为 Bearer）
- `expires_in`: 访问令牌有效期（秒，整数）
- `refresh_token`: 刷新令牌（字符串）
- `refresh_expires_in`: 刷新令牌有效期（秒，整数）
- `scope`: 权限范围（可选）

### 3.2 UserInfo

系统必须支持以下用户信息结构：

- `sub`: 用户唯一标识（字符串）
- `username`: 用户名（字符串）
- `status`: 用户状态（整数）
- `client_id`: 客户端 ID（字符串）
- `expire_at`: 过期时间（可选字符串指针）
- `is_expired`: 是否已过期（可选布尔指针）

## 4. 非功能需求

### 4.1 安全性

- **NFR1**: client_secret 必须在后端保密，不通过任何 API 暴露
- **NFR2**: 令牌交换必须使用 POST 方法和表单编码
- **NFR3**: 与 OAuth2 服务器通信建议使用 HTTPS

### 4.2 可用性

- **NFR4**: 错误响应必须包含有意义的错误描述
- **NFR5**: API 设计遵循 Go 惯用风格，简洁一致

### 4.3 可测试性

- **NFR6**: 所有公共函数必须编写单元测试
- **NFR7**: 关键性能路径必须编写基准测试

## 5. 验收标准

| 验收项 | 验收标准 |
|--------|----------|
| AC1 | 配置可通过环境变量加载，环境变量缺失时使用默认值 |
| AC2 | `/api/oauth2/config` 返回公开配置，不含 client_secret |
| AC3 | `/api/oauth2/callback` 正确处理授权码并返回令牌 |
| AC4 | `/api/oauth2/userinfo` 正确验证令牌并返回用户信息 |
| AC5 | `/api/oauth2/refresh` 正确刷新令牌并返回新令牌 |
| AC6 | 错误响应格式统一，包含 error 和 error_description |
| AC7 | 代码遵循 Go 官方规范（gofmt） |
| AC8 | 单元测试覆盖核心逻辑 |
