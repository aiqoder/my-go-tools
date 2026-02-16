package uf

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

// ExampleClient 演示创建默认配置的客户端
func ExampleClient() {
	// 创建默认配置的客户端
	// BaseURL 默认为 https://uf.yigechengzi.com/
	// 超时时间默认为 30 秒
	client := NewClient()

	fmt.Println("客户端创建成功")
	fmt.Printf("BaseURL: %s\n", client.GetBaseURL())

	// Output:
	// 客户端创建成功
	// BaseURL: https://uf.yigechengzi.com/
}

// ExampleClient_withOptions 演示使用选项自定义客户端配置
func ExampleClient_withOptions() {
	// 使用选项自定义配置
	client := NewClient(
		WithTimeout(60*time.Second),
		WithHTTPClient(&http.Client{
			Timeout: 60 * time.Second,
		}),
	)

	fmt.Printf("客户端创建成功，超时时间: %v\n", client.GetHTTPClient().Timeout)

	// Output:
	// 客户端创建成功，超时时间: 1m0s
}

// Example_client_Get 演示 GET 请求
func Example_client_Get() {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "data": "test"}`))
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(WithBaseURL(server.URL))

	// 发起 GET 请求
	data, err := client.Get("/api/test")
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	// 解析响应
	var resp map[string]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("响应: %v\n", resp["data"])

	// Output:
	// 响应: test
}

// Example_client_Post 演示 POST 请求
func Example_client_Post() {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(WithBaseURL(server.URL))

	// 发起 POST 请求
	data, err := client.Post("/api/submit", map[string]string{
		"name": "test",
	})
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("请求成功: %v\n", resp.IsOK())

	// Output:
	// 请求成功: true
}

// Example_client_PostForm 演示表单 POST 请求
func Example_client_PostForm() {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(WithBaseURL(server.URL))

	// 发起表单 POST 请求
	data, err := client.PostForm("/api/form", map[string][]string{
		"username": {"testuser"},
		"password": {"password123"},
	})
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("表单提交成功: %v\n", resp.IsOK())

	// Output:
	// 表单提交成功: true
}

// Example_client_DoJSONRequest 演示通用 JSON 请求
func Example_client_DoJSONRequest() {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "result": "success"}`))
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(WithBaseURL(server.URL))

	// 定义请求和响应结构
	type Request struct {
		Action string `json:"action"`
	}
	type ResponseWithResult struct {
		OK     bool   `json:"ok"`
		Result string `json:"result"`
	}

	// 发起请求
	var resp ResponseWithResult
	err := client.DoJSONRequest(http.MethodPost, "/api/action", &Request{Action: "test"}, &resp)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	fmt.Printf("操作结果: %s\n", resp.Result)

	// Output:
	// 操作结果: success
}

// Example_client_GetWithContext 演示带上下文的 GET 请求
func Example_client_GetWithContext() {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(WithBaseURL(server.URL))

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发起带上下文的请求
	data, err := client.GetWithContext(ctx, "/api/test")
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("请求成功: %v\n", resp.IsOK())

	// Output:
	// 请求成功: true
}

// Example_client_SetHTTPClient 演示设置自定义 HTTP 客户端
func Example_client_SetHTTPClient() {
	// 创建自定义 HTTP 客户端
	customClient := &http.Client{
		Timeout: 60 * time.Second,
		// 可以配置 Transport、CheckRedirect 等
	}

	// 创建默认客户端
	client := NewClient()

	// 设置自定义客户端
	client.SetHTTPClient(customClient)

	fmt.Printf("HTTP 客户端超时时间: %v\n", client.GetHTTPClient().Timeout)

	// Output:
	// HTTP 客户端超时时间: 1m0s
}

// Example_error_handling 演示错误处理
func Example_error_handling() {
	// 测试服务器返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok": false, "error": "参数错误"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/test")

	if err != nil {
		fmt.Printf("捕获到错误: %s\n", err.Error())
	}

	// Output:
	// 捕获到错误: [SERVER_ERROR] 参数错误
}
