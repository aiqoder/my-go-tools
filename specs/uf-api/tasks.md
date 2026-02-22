# UF 用户反馈服务 - 新增 API 接口任务清单

## 任务列表

### 阶段一：类型定义

#### T1. 在 types.go 中新增请求/响应类型
- **描述**: 在 `uf/types.go` 中添加活跃度和激活检查相关的类型定义
- **验收标准**:
  - [ ] `ActivityRequest` 结构体定义完整
  - [ ] `ActivityResponse` 结构体定义完整
  - [ ] `ActivationCheckRequest` 结构体定义完整
  - [ ] `ActivationCheckResponse` 结构体定义完整
  - [ ] 包含完整的 go doc 注释
- **依赖**: 无
- **实现备注**: 待实现

### 阶段二：活跃度记录接口

#### T2. 创建 activity.go 实现文件
- **描述**: 实现活跃度记录服务
- **验收标准**:
  - [ ] `ActivityService` 结构体定义
  - [ ] `NewActivityService` 工厂函数
  - [ ] `CreateByGET` 方法实现
  - [ ] `CreateByPOST` 方法实现
  - [ ] 包含完整的 go doc 注释
- **依赖**: T1
- **实现备注**: 待实现

#### T3. 在 client.go 中添加便捷方法
- **描述**: 在 Client 上添加 Activity() 便捷方法
- **验收标准**:
  - [ ] `Activity() *ActivityService` 方法实现
  - [ ] 方法注释完整
- **依赖**: T2
- **实现备注**: 待实现

### 阶段三：激活检查接口

#### T4. 创建 activation.go 实现文件
- **描述**: 实现激活检查服务
- **验收标准**:
  - [ ] `ActivationService` 结构体定义
  - [ ] `NewActivationService` 工厂函数
  - [ ] `Check` 方法实现
  - [ ] 包含完整的 go doc 注释
- **依赖**: T1
- **实现备注**: 待实现

#### T5. 在 client.go 中添加便捷方法
- **描述**: 在 Client 上添加 Activation() 便捷方法
- **验收标准**:
  - [ ] `Activation() *ActivationService` 方法实现
  - [ ] 方法注释完整
- **依赖**: T4
- **实现备注**: 待实现

### 阶段四：测试

#### T6. 添加单元测试
- **描述**: 为新接口编写单元测试
- **验收标准**:
  - [ ] `TestActivityService_CreateByGET` 测试用例
  - [ ] `TestActivityService_CreateByPOST` 测试用例
  - [ ] `TestActivationService_Check` 测试用例
  - [ ] 使用 table-driven test 模式
- **依赖**: T2, T4
- **实现备注**: 待实现

#### T7. 添加示例代码
- **描述**: 在 examples_test.go 中添加使用示例
- **验收标准**:
  - [ ] 活跃度记录 GET 示例
  - [ ] 活跃度记录 POST 示例
  - [ ] 激活检查示例
- **依赖**: T2, T4
- **实现备注**: 待实现

### 阶段五：文档

#### T8. 更新 README 文档
- **描述**: 更新 uf/README.md 添加新 API 说明
- **验收标准**:
  - [ ] 添加 Activity API 说明
  - [ ] 添加 Activation API 说明
  - [ ] 包含使用示例
- **依赖**: T2, T4
- **实现备注**: 待实现

#### T9. 代码质量检查
- **描述**: 确保代码符合 Go 规范
- **验收标准**:
  - [ ] `gofmt -d` 无输出
  - [ ] `go vet` 无警告
- **依赖**: 所有任务
- **实现备注**: 待实现

## 执行顺序

```
T1 ─┬─> T2 ─> T3 ─> T7 ─> T8 ─> T9
     │
     └─> T4 ─> T5 ─> T7 ─> T8 ─> T9
              │
              └─> T6 ─┘
```

## 注意事项

1. 按依赖顺序执行任务
2. 每完成一个任务，在对应的验收标准上打勾
3. 保持代码风格与现有 uf 包一致（中文注释、go fmt）
4. 新增的类型放在 types.go 中，新增的服务实现放在独立文件中
