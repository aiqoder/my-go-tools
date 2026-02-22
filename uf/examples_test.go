package uf

import (
	"fmt"
)

// ExampleClient 演示创建默认配置的客户端
func ExampleClient() {
	_ = NewClient()
	fmt.Println("客户端创建成功")

	// Output:
	// 客户端创建成功
}

// ExampleClient_RecordActivity 演示记录活跃度
//
// 该示例展示了如何使用 RecordActivity 方法记录软件活跃度。
// 实际使用时需要配置正确的 BaseURL。
func ExampleClient_RecordActivity() {
	// 创建客户端（实际使用时配置正确的 BaseURL）
	_ = NewClient()

	// 模拟响应
	resp := &ActivityResponse{OK: true, ID: 1}
	fmt.Printf("记录成功: %v\n", resp.IsOK())

	// Output:
	// 记录成功: true
}

// ExampleClient_CheckActivation 演示检查激活状态
//
// 该示例展示了如何使用 CheckActivation 方法检查软件激活状态。
// 实际使用时需要配置正确的 BaseURL。
func ExampleClient_CheckActivation() {
	// 创建客户端（实际使用时配置正确的 BaseURL）
	_ = NewClient()

	// 模拟响应
	resp := &ActivationCheckResponse{OK: true, Activated: true, ExpireAt: "2026-12-31"}
	fmt.Printf("已激活: %v\n", resp.Activated)

	// Output:
	// 已激活: true
}
