# UF 用户反馈服务 - 新增 API 接口设计文档

## 1. 架构设计

### 1.1 文件结构

在现有 `uf` 包基础上新增以下内容：

```
uf/
├── client.go           # [已存在] 客户端核心实现
├── config.go           # [已存在] 配置管理
├── types.go           # [已存在] 类型定义 → 新增 Activity/Activation 类型
├── errors.go          # [已存在] 错误定义
├── activity.go        # [新增] 活跃度记录接口
├── activation.go      # [新增] 激活检查接口
├── examples_test.go   # [已存在] 使用示例 → 新增示例
├── client_test.go     # [已存在] 单元测试 → 新增测试
└── README.md          # [已存在] 包文档 → 新增 API 说明
```

### 1.2 设计模式

沿用现有设计模式：
- **工厂模式**：通过 `ActivityService(client *Client)` 创建服务实例
- **方法链式调用**：服务方法返回响应结构，支持链式处理

## 2. 数据结构设计

### 2.1 活跃度记录相关类型

```go
// ActivityRequest 活跃度记录请求
type ActivityRequest struct {
    // SoftwareID 软件 ID
    SoftwareID uint `json:"softwareId"`
}

// ActivityResponse 活跃度记录响应
type ActivityResponse struct {
    // OK 请求是否成功
    OK bool `json:"ok"`
    // ID 活跃度记录 ID
    ID uint `json:"id,omitempty"`
    // Error 错误信息（仅在 OK 为 false 时存在）
    Error string `json:"error,omitempty"`
}
```

### 2.2 激活检查相关类型

```go
// ActivationCheckRequest 激活检查请求
type ActivationCheckRequest struct {
    // SoftwareID 软件 ID
    SoftwareID uint `json:"softwareId"`
    // MachineCode 机器码
    MachineCode string `json:"machineCode"`
}

// ActivationCheckResponse 激活检查响应
type ActivationCheckResponse struct {
    // OK 请求是否成功
    OK bool `json:"ok"`
    // Activated 是否已激活
    Activated bool `json:"activated"`
    // ExpireAt 过期时间
    ExpireAt string `json:"expireAt,omitempty"`
    // Error 错误信息（仅在 OK 为 false 时存在）
    Error string `json:"error,omitempty"`
}
```

## 3. 接口设计

### 3.1 活跃度记录服务

```go
// ActivityService 活跃度记录服务
//
// 提供软件活跃度记录的创建和更新功能。
type ActivityService struct {
    client *Client
}

// NewActivityService 创建活跃度记录服务
func NewActivityService(client *Client) *ActivityService

// CreateByGET 通过 GET 请求创建活跃度记录
//
// 参数 softwareId 为软件 ID。
// 返回响应和错误。
func (s *ActivityService) CreateByGET(softwareId uint) (*ActivityResponse, error)

// CreateByPOST 通过 POST 请求创建活跃度记录
//
// 参数 softwareId 为软件 ID。
// 返回响应和错误。
func (s *ActivityService) CreateByPOST(softwareId uint) (*ActivityResponse, error)
```

### 3.2 激活检查服务

```go
// ActivationService 激活检查服务
//
// 提供软件激活状态检查功能。
type ActivationService struct {
    client *Client
}

// NewActivationService 创建激活检查服务
func NewActivationService(client *Client) *ActivationService

// Check 检查软件激活状态
//
// 参数 softwareId 为软件 ID，machineCode 为机器码。
// 返回响应和错误。
func (s *ActivationService) Check(softwareId uint, machineCode string) (*ActivationCheckResponse, error)
```

### 3.3 客户端集成方法

在 `Client` 结构体上添加便捷方法：

```go
// Activity 返回活跃度记录服务
func (c *Client) Activity() *ActivityService

// Activation 返回激活检查服务
func (c *Client) Activation() *ActivationService
```

## 4. 实现细节

### 4.1 活跃度记录实现

**GET 请求**：
```go
func (s *ActivityService) CreateByGET(softwareId uint) (*ActivityResponse, error) {
    path := fmt.Sprintf("/api/activity?softwareId=%d", softwareId)
    resp := &ActivityResponse{}
    err := s.client.DoJSONRequest("GET", path, nil, resp)
    return resp, err
}
```

**POST 请求**：
```go
func (s *ActivityService) CreateByPOST(softwareId uint) (*ActivityResponse, error) {
    req := &ActivityRequest{SoftwareID: softwareId}
    resp := &ActivityResponse{}
    err := s.client.DoJSONRequest("POST", "/api/activity", req, resp)
    return resp, err
}
```

### 4.2 激活检查实现

```go
func (s *ActivationService) Check(softwareId uint, machineCode string) (*ActivationCheckResponse, error) {
    req := &ActivationCheckRequest{
        SoftwareID:  softwareId,
        MachineCode: machineCode,
    }
    resp := &ActivationCheckResponse{}
    err := s.client.DoJSONRequest("POST", "/api/activation/check", req, resp)
    return resp, err
}
```

## 5. 错误处理

沿用现有错误处理模式：
- 网络错误：通过 `errors.go` 中的错误创建函数处理
- API 错误：解析响应中的 `Error` 字段
- 参数错误：使用 `NewParamsError`

## 6. 测试设计

### 6.1 单元测试

```go
// 测试 ActivityService
func TestActivityService_CreateByGET(t *testing.T)
func TestActivityService_CreateByPOST(t *testing.T)

// 测试 ActivationService
func TestActivationService_Check(t *testing.T)

// 测试客户端便捷方法
func TestClient_Activity(t *testing.T)
func TestClient_Activation(t *testing.T)
```

### 6.2 示例代码

```go
// 活跃度记录示例
func ExampleActivityService_CreateByGET()
func ExampleActivityService_CreateByPOST()

// 激活检查示例
func ExampleActivationService_Check()
```

## 7. 技术决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 服务创建方式 | 工厂函数 | 与现有选项模式保持一致 |
| 响应解析 | 结构体标签 | 利用现有 JSON 解析能力 |
| 错误传播 | 直接返回错误 | 保持简洁，不做过度封装 |
| GET 参数传递 | Query String | 符合接口规范 |

## 8. 向后兼容性

- 现有 `Client` 的公共方法保持不变
- 新增方法均为增量添加
- 不修改现有类型定义
