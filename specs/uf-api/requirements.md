# UF 用户反馈服务 - 新增 API 接口需求规格

## 1. 需求概述

为 UF 用户反馈服务客户端增加两个新接口：
1. **活跃度记录接口** (`/api/activity`) - 用于创建或更新软件的活跃度记录
2. **激活检查接口** (`/api/activation/check`) - 用于检查软件是否已激活及激活状态

## 2. 背景与目标

### 背景
- 用户反馈服务需要支持软件活跃度追踪功能
- 需要提供激活状态检查能力，以支持软件授权验证场景

### 目标
- 新增接口与现有 UF 客户端无缝集成
- 遵循现有代码风格和错误处理模式
- 提供简洁易用的 Go API

## 3. 功能需求

### 3.1 活跃度记录接口 (/api/activity)

| 功能 | 描述 | 优先级 |
|------|------|--------|
| GET 请求支持 | 通过 Query 参数 `softwareId` 提交活跃度记录 | 必须 |
| POST 请求支持 | 通过 JSON Body 提交活跃度记录 | 必须 |
| 响应解析 | 解析通用响应 `{ ok: true, id: 1 }` | 必须 |

#### 请求参数
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| softwareId | uint | 是 | 软件 ID |

#### curl 示例
```bash
# GET 请求
curl -X GET "http://localhost:8080/api/activity?softwareId=1"

# POST 请求
curl -X POST "http://localhost:8080/api/activity" \
  -H "Content-Type: application/json" \
  -d '{"softwareId": 1}'
```

#### 响应示例
```json
{
  "ok": true,
  "id": 1
}
```

### 3.2 激活检查接口 (/api/activation/check)

| 功能 | 描述 | 优先级 |
|------|------|--------|
| POST 请求支持 | 通过 JSON Body 提交激活检查请求 | 必须 |
| 激活状态解析 | 解析响应中的 `activated` 和 `expireAt` 字段 | 必须 |

#### 请求参数
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| softwareId | uint | 是 | 软件 ID |
| machineCode | string | 是 | 机器码 |

#### curl 示例
```bash
curl -X POST "http://localhost:8080/api/activation/check" \
  -H "Content-Type: application/json" \
  -d '{"softwareId": 1, "machineCode": "ABC-123-XYZ"}'
```

#### 响应示例
```json
{
  "ok": true,
  "activated": true,
  "expireAt": "2026-12-31 23:59:59"
}
```

## 4. 非功能需求

### 4.1 性能要求
- 继承现有客户端的超时配置（默认 30 秒）
- 支持上下文取消（context）

### 4.2 可维护性
- 遵循现有代码风格（中文注释、go doc）
- 与现有类型定义保持一致

### 4.3 扩展性
- 使用与现有代码相同的错误处理模式
- 便于后续添加更多激活相关接口

## 5. 验收标准

### 5.1 功能验收
- [ ] 活跃度记录接口支持 GET 和 POST 两种方式
- [ ] 激活检查接口支持 POST 方式
- [ ] 正确解析响应中的 `id`、`activated`、`expireAt` 字段

### 5.2 代码质量验收
- [ ] 代码通过 gofmt 格式化
- [ ] 公共 API 包含完整的 go doc 注释
- [ ] 包含单元测试
- [ ] 包含示例代码

### 5.3 集成验收
- [ ] 新接口与现有 Client 无缝集成
- [ ] 错误处理与现有模式一致

## 6. 参考资料

- 现有 UF 客户端实现 (`uf/client.go`)
- 现有类型定义 (`uf/types.go`)
- 现有错误处理 (`uf/errors.go`)
