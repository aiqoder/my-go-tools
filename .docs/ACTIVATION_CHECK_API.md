# 激活状态检查接口对接手册

## 概述

激活状态检查接口 (`/api/activation/check`) 用于验证指定软件在指定机器上的激活状态。该接口常用于软件授权验证，可以作为中间件集成到其他系统中，实现统一的授权管理。

## 接口信息

- **接口地址**: `/api/activation/check`
- **请求方式**: `POST`
- **Content-Type**: `application/json` 或 `application/x-www-form-urlencoded`
- **认证方式**: 无需认证（公开接口）

## 对接规范

### 参数获取规范

为了统一对接标准，系统对接时必须遵循以下规范：

1. **机器码 (machineCode)**
   - **来源**: 必须从环境变量 `ykey` 中读取
   - **环境变量名**: `ykey`（固定，不可更改）
   - **说明**: 机器码用于标识设备，应通过环境变量配置，便于部署和管理

2. **软件ID (softwareId)**
   - **来源**: 在开发时在系统中固定配置
   - **说明**: 软件ID应在代码中作为常量或配置项固定，不同软件使用不同的ID

### 规范示例

```bash
# 环境变量配置示例
export ykey="ABC123-def_456"
```

```go
// 软件ID在代码中固定
const SOFTWARE_ID = 1  // 根据实际软件ID设置
```

## 请求参数

### 参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `softwareId` | `uint` | 是 | 软件ID，必须大于0，应在开发时固定配置 |
| `machineCode` | `string` | 是 | 机器码，格式要求见下方说明，必须从环境变量 `ykey` 读取 |

### 机器码格式要求

- 长度：4-128 个字符
- 允许字符：字母（大小写）、数字、下划线（`_`）、短横线（`-`）
- 正则表达式：`^[A-Za-z0-9_-]{4,128}$`
- 示例：`ABC123-def_456`、`machine-001`、`DEVICE_2024`

### 请求示例

#### JSON 格式

```json
{
  "softwareId": 1,
  "machineCode": "ABC123-def_456"
}
```

#### Form 格式

```
softwareId=1&machineCode=ABC123-def_456
```

## 响应格式

### 成功响应

**HTTP 状态码**: `200 OK`

```json
{
  "ok": true,
  "activated": true,
  "expireAt": "2024-12-31 23:59:59"
}
```

### 响应字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| `ok` | `bool` | 请求是否成功，`true` 表示成功，`false` 表示失败 |
| `activated` | `bool` | 是否已激活且未过期，`true` 表示已激活且有效 |
| `expireAt` | `string\|null` | 到期时间，格式：`YYYY-MM-DD HH:mm:ss`。如果未激活则为 `null` |
| `error` | `string` | 错误信息，仅在 `ok` 为 `false` 时存在 |

### 响应状态说明

| `ok` | `activated` | `expireAt` | 说明 |
|------|-------------|------------|------|
| `true` | `true` | `"2024-12-31 23:59:59"` | 已激活且未过期 |
| `true` | `false` | `"2024-12-31 23:59:59"` | 已激活但已过期 |
| `true` | `false` | `null` | 未激活 |
| `false` | - | - | 请求失败，查看 `error` 字段 |

### 错误响应

**HTTP 状态码**: `400 Bad Request`

```json
{
  "ok": false,
  "error": "软件ID和机器码不能为空"
}
```

### 常见错误信息

| 错误信息 | 说明 |
|----------|------|
| `参数错误` | 请求参数格式不正确 |
| `软件ID和机器码不能为空` | 缺少必需参数 |
| `机器码格式不正确` | 机器码不符合格式要求 |
| `软件不存在` | 指定的软件ID不存在 |
| `查询激活状态失败: ...` | 服务器内部错误 |

## 部署配置示例

### 环境变量配置

#### Linux/Mac

```bash
# 方式1: 在启动脚本中设置
export ykey="ABC123-def_456"
./your-app

# 方式2: 在 .bashrc 或 .zshrc 中设置（不推荐用于生产环境）
export ykey="ABC123-def_456"

# 方式3: 使用 .env 文件（如果框架支持）
echo 'ykey=ABC123-def_456' > .env
```

#### Windows

```cmd
REM 方式1: 在命令提示符中设置
set ykey=ABC123-def_456
your-app.exe

REM 方式2: 在系统环境变量中设置（控制面板 -> 系统 -> 高级系统设置 -> 环境变量）
```

#### Docker

```bash
# 方式1: 通过 -e 参数
docker run -e ykey="ABC123-def_456" your-image

# 方式2: 通过环境变量文件
echo 'ykey=ABC123-def_456' > .env
docker run --env-file .env your-image

# 方式3: 在 docker-compose.yml 中配置
# version: '3'
# services:
#   app:
#     image: your-image
#     environment:
#       - ykey=ABC123-def_456
```

#### Kubernetes

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: your-app
spec:
  containers:
  - name: app
    image: your-image
    env:
    - name: ykey
      value: "ABC123-def_456"
    # 或使用 Secret
    # env:
    # - name: ykey
    #   valueFrom:
    #     secretKeyRef:
    #       name: app-secrets
    #       key: ykey
```

### 代码配置示例

```go
// config.go
package config

import (
    "os"
    "fmt"
)

// 软件ID在开发时固定配置
const SOFTWARE_ID = 1  // 根据实际软件ID设置

// GetMachineCode 从环境变量 ykey 获取机器码
func GetMachineCode() (string, error) {
    machineCode := os.Getenv("ykey")
    if machineCode == "" {
        return "", fmt.Errorf("环境变量 ykey 未设置")
    }
    return machineCode, nil
}

// MustGetMachineCode 从环境变量 ykey 获取机器码，如果不存在则 panic
func MustGetMachineCode() string {
    machineCode, err := GetMachineCode()
    if err != nil {
        panic(err)
    }
    return machineCode
}
```

## 使用示例

### 1. 基础调用（Go）

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

// 软件ID在开发时固定配置
const SOFTWARE_ID = 1  // 根据实际软件ID设置

type CheckRequest struct {
    SoftwareID  uint   `json:"softwareId"`
    MachineCode string `json:"machineCode"`
}

type CheckResponse struct {
    OK        bool    `json:"ok"`
    Activated bool    `json:"activated"`
    ExpireAt  *string `json:"expireAt,omitempty"`
    Error     string  `json:"error,omitempty"`
}

func CheckActivation(baseURL string, softwareID uint, machineCode string) (*CheckResponse, error) {
    reqBody := CheckRequest{
        SoftwareID:  softwareID,
        MachineCode: machineCode,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    
    resp, err := http.Post(baseURL+"/api/activation/check", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result CheckResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

func main() {
    // 从环境变量 ykey 读取机器码
    machineCode := os.Getenv("ykey")
    if machineCode == "" {
        fmt.Println("错误: 环境变量 ykey 未设置")
        return
    }
    
    result, err := CheckActivation("https://api.example.com", SOFTWARE_ID, machineCode)
    if err != nil {
        fmt.Printf("请求失败: %v\n", err)
        return
    }
    
    if !result.OK {
        fmt.Printf("检查失败: %s\n", result.Error)
        return
    }
    
    if result.Activated {
        fmt.Printf("已激活，到期时间: %s\n", *result.ExpireAt)
    } else {
        fmt.Println("未激活或已过期")
    }
}
```

### 2. Gin 中间件示例

```go
package main

import (
    "net/http"
    "os"
    "github.com/gin-gonic/gin"
)

// 软件ID在开发时固定配置
const SOFTWARE_ID = 1  // 根据实际软件ID设置

// ActivationCheckMiddleware 激活状态检查中间件
func ActivationCheckMiddleware(baseURL string, softwareID uint) gin.HandlerFunc {
    // 从环境变量 ykey 读取机器码（在中间件初始化时读取，避免每次请求都读取）
    machineCode := os.Getenv("ykey")
    if machineCode == "" {
        panic("环境变量 ykey 未设置，无法初始化激活检查中间件")
    }
    
    return func(c *gin.Context) {
        // 调用激活检查接口
        result, err := CheckActivation(baseURL, softwareID, machineCode)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "激活检查服务不可用",
            })
            c.Abort()
            return
        }
        
        if !result.OK {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": result.Error,
            })
            c.Abort()
            return
        }
        
        if !result.Activated {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "软件未激活或已过期",
                "expireAt": result.ExpireAt,
            })
            c.Abort()
            return
        }
        
        // 将激活信息存储到上下文，供后续使用
        c.Set("activation_expire_at", result.ExpireAt)
        c.Next()
    }
}

func main() {
    r := gin.Default()
    
    // 使用激活检查中间件
    api := r.Group("/api")
    api.Use(ActivationCheckMiddleware("https://api.example.com", SOFTWARE_ID))
    {
        api.GET("/protected", func(c *gin.Context) {
            expireAt, _ := c.Get("activation_expire_at")
            c.JSON(http.StatusOK, gin.H{
                "message": "访问成功",
                "expireAt": expireAt,
            })
        })
    }
    
    r.Run(":8080")
}
```

### 3. 带缓存的中间件（推荐）

```go
package main

import (
    "fmt"
    "os"
    "sync"
    "time"
    "net/http"
    "github.com/gin-gonic/gin"
)

// 软件ID在开发时固定配置
const SOFTWARE_ID = 1  // 根据实际软件ID设置

type ActivationCache struct {
    mu          sync.RWMutex
    cache       map[string]*CacheItem
    ttl         time.Duration
    checkFunc   func(uint, string) (*CheckResponse, error)
}

type CacheItem struct {
    Result    *CheckResponse
    ExpiresAt time.Time
}

func NewActivationCache(ttl time.Duration, checkFunc func(uint, string) (*CheckResponse, error)) *ActivationCache {
    cache := &ActivationCache{
        cache:     make(map[string]*CacheItem),
        ttl:       ttl,
        checkFunc: checkFunc,
    }
    
    // 定期清理过期缓存
    go cache.cleanup()
    
    return cache
}

func (c *ActivationCache) Get(softwareID uint, machineCode string) (*CheckResponse, error) {
    key := fmt.Sprintf("%d:%s", softwareID, machineCode)
    
    c.mu.RLock()
    item, exists := c.cache[key]
    c.mu.RUnlock()
    
    if exists && time.Now().Before(item.ExpiresAt) {
        return item.Result, nil
    }
    
    // 缓存未命中或已过期，重新检查
    result, err := c.checkFunc(softwareID, machineCode)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    c.mu.Lock()
    c.cache[key] = &CacheItem{
        Result:    result,
        ExpiresAt: time.Now().Add(c.ttl),
    }
    c.mu.Unlock()
    
    return result, nil
}

func (c *ActivationCache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, item := range c.cache {
            if now.After(item.ExpiresAt) {
                delete(c.cache, key)
            }
        }
        c.mu.Unlock()
    }
}

// 使用缓存的中间件
func CachedActivationCheckMiddleware(cache *ActivationCache, softwareID uint, machineCode string) gin.HandlerFunc {
    return func(c *gin.Context) {
        result, err := cache.Get(softwareID, machineCode)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "激活检查失败"})
            c.Abort()
            return
        }
        
        if !result.OK || !result.Activated {
            c.JSON(http.StatusForbidden, gin.H{"error": "软件未激活或已过期"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func main() {
    // 从环境变量 ykey 读取机器码
    machineCode := os.Getenv("ykey")
    if machineCode == "" {
        panic("环境变量 ykey 未设置")
    }
    
    // 创建缓存
    cache := NewActivationCache(5*time.Minute, CheckActivation)
    
    r := gin.Default()
    
    // 使用带缓存的激活检查中间件
    api := r.Group("/api")
    api.Use(CachedActivationCheckMiddleware(cache, SOFTWARE_ID, machineCode))
    {
        api.GET("/protected", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"message": "访问成功"})
        })
    }
    
    r.Run(":8080")
}
```

### 4. JavaScript/TypeScript 示例

```javascript
// 软件ID在开发时固定配置
const SOFTWARE_ID = 1;  // 根据实际软件ID设置

async function checkActivation(baseURL, softwareId, machineCode) {
    const response = await fetch(`${baseURL}/api/activation/check`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            softwareId: softwareId,
            machineCode: machineCode,
        }),
    });
    
    const data = await response.json();
    
    if (!data.ok) {
        throw new Error(data.error || '激活检查失败');
    }
    
    return {
        activated: data.activated,
        expireAt: data.expireAt,
    };
}

// 使用示例
// 从环境变量 ykey 读取机器码（Node.js 环境）
const machineCode = process.env.ykey;
if (!machineCode) {
    console.error('错误: 环境变量 ykey 未设置');
    process.exit(1);
}

try {
    const result = await checkActivation('https://api.example.com', SOFTWARE_ID, machineCode);
    if (result.activated) {
        console.log('已激活，到期时间:', result.expireAt);
    } else {
        console.log('未激活或已过期');
    }
} catch (error) {
    console.error('检查失败:', error.message);
}
```

### 5. Python 示例

```python
import os
import requests
from typing import Dict

# 软件ID在开发时固定配置
SOFTWARE_ID = 1  # 根据实际软件ID设置

def check_activation(base_url: str, software_id: int, machine_code: str) -> Dict:
    """
    检查激活状态
    
    Args:
        base_url: API 基础地址
        software_id: 软件ID
        machine_code: 机器码
    
    Returns:
        包含激活状态的字典
    """
    url = f"{base_url}/api/activation/check"
    payload = {
        "softwareId": software_id,
        "machineCode": machine_code
    }
    
    try:
        response = requests.post(url, json=payload, timeout=5)
        response.raise_for_status()
        data = response.json()
        
        if not data.get("ok"):
            raise Exception(data.get("error", "激活检查失败"))
        
        return {
            "activated": data.get("activated", False),
            "expire_at": data.get("expireAt")
        }
    except requests.RequestException as e:
        raise Exception(f"请求失败: {str(e)}")

# 使用示例
# 从环境变量 ykey 读取机器码
machine_code = os.getenv("ykey")
if not machine_code:
    print("错误: 环境变量 ykey 未设置")
    exit(1)

try:
    result = check_activation("https://api.example.com", SOFTWARE_ID, machine_code)
    if result["activated"]:
        print(f"已激活，到期时间: {result['expire_at']}")
    else:
        print("未激活或已过期")
except Exception as e:
    print(f"检查失败: {e}")
```

## 最佳实践

### 1. 参数获取规范

- **机器码**: 必须从环境变量 `ykey` 读取，启动时检查环境变量是否存在
- **软件ID**: 在代码中作为常量固定配置，不要从配置文件或环境变量读取
- **环境变量检查**: 应用启动时验证 `ykey` 环境变量是否存在，不存在则拒绝启动

### 2. 错误处理

- **网络错误**: 建议设置合理的超时时间（如 5 秒），网络错误时根据业务需求决定是否允许访问
- **服务不可用**: 可以考虑降级策略，如允许访问或限制功能
- **参数错误**: 严格验证参数格式，避免无效请求
- **环境变量缺失**: 如果 `ykey` 未设置，应拒绝启动或返回明确的错误信息

### 3. 性能优化

- **缓存机制**: 对于频繁调用的场景，建议实现缓存机制（如示例 3），减少对服务器的压力
- **缓存时间**: 建议缓存时间设置为 1-5 分钟，平衡实时性和性能
- **并发控制**: 对于高并发场景，考虑使用连接池和请求限流

### 4. 安全性

- **HTTPS**: 生产环境必须使用 HTTPS 传输
- **机器码保护**: 机器码通过环境变量 `ykey` 配置，避免硬编码在代码中
- **环境变量安全**: 确保环境变量文件权限正确，避免泄露
- **日志记录**: 记录激活检查的日志，但不要记录完整的机器码（可记录部分或哈希值）

### 5. 中间件设计

- **初始化时读取**: 在中间件初始化时从环境变量读取机器码，避免每次请求都读取
- **上下文传递**: 将激活信息存储到上下文，供后续处理使用
- **错误响应**: 提供清晰的错误信息，便于客户端处理

## 注意事项

1. **机器码唯一性**: 机器码应在同一软件内唯一标识一台设备
2. **环境变量规范**: 
   - 机器码必须从环境变量 `ykey` 读取，环境变量名固定为 `ykey`
   - 应用启动时应检查环境变量是否存在，不存在则拒绝启动
   - 不要将机器码硬编码在代码中
3. **软件ID固定**: 
   - 软件ID应在开发时在代码中固定配置
   - 不同软件使用不同的ID，不要从配置文件或环境变量读取
4. **时区处理**: 到期时间使用服务器时区，客户端需要根据实际情况处理
5. **并发激活**: 同一机器码在同一软件下只能有一个有效激活记录
6. **过期判断**: 接口返回的 `activated` 字段已经考虑了过期时间，无需客户端再次判断
7. **软件ID**: 确保使用正确的软件ID，不同软件之间的激活记录是独立的

## 常见问题

### Q1: 机器码格式验证失败怎么办？

**A**: 检查机器码是否符合格式要求：
- 长度在 4-128 字符之间
- 只包含字母、数字、下划线和短横线
- 去除首尾空格
- 确保环境变量 `ykey` 已正确设置

### Q1.1: 环境变量 ykey 未设置怎么办？

**A**: 
- 检查环境变量是否已正确配置：`echo $ykey`（Linux/Mac）或 `echo %ykey%`（Windows）
- 在启动脚本中设置：`export ykey="your-machine-code"`
- 使用 `.env` 文件（如果框架支持）
- 在 Docker 容器中通过 `-e ykey=xxx` 或环境变量文件设置
- 应用启动时应检查环境变量，不存在则拒绝启动并提示错误信息

### Q2: 如何获取软件ID？

**A**: 
- 软件ID由系统管理员在后台配置，请联系管理员获取
- 获取后应在代码中作为常量固定配置，不要从配置文件读取
- 不同软件使用不同的ID，确保使用正确的ID

### Q3: 激活检查失败是否应该阻止访问？

**A**: 根据业务需求决定：
- 严格模式：未激活或已过期直接拒绝访问
- 宽松模式：允许访问但限制功能
- 降级模式：服务不可用时允许访问

### Q4: 如何处理时区问题？

**A**: 服务器返回的时间为服务器时区，客户端需要：
- 解析时间字符串时考虑时区
- 与本地时间比较时进行时区转换
- 或直接使用服务器返回的 `activated` 字段

## 更新日志

- **v1.0.0** (2024-XX-XX): 初始版本

## 技术支持

如有问题，请联系技术支持团队或查看相关文档。
