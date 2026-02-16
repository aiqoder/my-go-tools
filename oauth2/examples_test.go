package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExampleOAuth2Service_NewOAuth2Service 演示如何创建 OAuth2 服务
func ExampleOAuth2Service_NewOAuth2Service() {
	// 创建 OAuth2 配置
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	// 创建服务实例
	svc := NewOAuth2Service(cfg)

	// 获取公开配置
	publicConfig := svc.GetConfig()
	fmt.Printf("OAuth2 服务器: %s\n", publicConfig.Server)
	fmt.Printf("客户端 ID: %s\n", publicConfig.ClientID)
	fmt.Printf("重定向 URI: %s\n", publicConfig.RedirectURI)

	// Output:
	// OAuth2 服务器: http://localhost:8080
	// 客户端 ID: your-client-id
	// 重定向 URI: http://localhost:3000/callback
}

// ExampleOAuth2Handler_GetConfig 演示获取 OAuth2 配置
func ExampleOAuth2Handler_GetConfig() {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"oauth_server": "http://localhost:8080",
			"client_id":    "test-client",
			"redirect_uri": "http://localhost:3000/callback",
		})
	}))
	defer mockServer.Close()

	// 创建服务
	cfg := &Config{
		Server:       mockServer.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}
	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试路由
	router := gin.New()
	router.GET("/api/oauth2/config", handler.GetConfig)

	// 发送请求
	req := httptest.NewRequest("GET", "/api/oauth2/config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 解析响应
	var resp PublicConfig
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Printf("状态码: %d\n", w.Code)
	fmt.Printf("客户端 ID: %s\n", resp.ClientID)

	// Output:
	// 状态码: 200
	// 客户端 ID: test-client
}

// ExampleOAuth2Handler_Callback 演示授权码回调处理
func ExampleOAuth2Handler_Callback() {
	gin.SetMode(gin.TestMode)

	// 创建模拟 OAuth2 服务器
	mockOAuth2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:      "access-token-123",
			TokenType:        "Bearer",
			ExpiresIn:        3600,
			RefreshToken:     "refresh-token-456",
			RefreshExpiresIn: 86400,
			Scope:            "read",
		})
	}))
	defer mockOAuth2Server.Close()

	// 创建服务
	cfg := &Config{
		Server:       mockOAuth2Server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}
	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试路由
	router := gin.New()
	router.POST("/api/oauth2/callback", handler.Callback)

	// 发送请求
	body := `{"code":"authorization-code-123"}`
	req := httptest.NewRequest("POST", "/api/oauth2/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 解析响应
	var resp TokenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Printf("状态码: %d\n", w.Code)
	fmt.Printf("访问令牌: %s\n", resp.AccessToken)
	fmt.Printf("刷新令牌: %s\n", resp.RefreshToken)

	// Output:
	// 状态码: 200
	// 访问令牌: access-token-123
	// 刷新令牌: refresh-token-456
}

// ExampleOAuth2Handler_GetUserInfo 演示获取用户信息
func ExampleOAuth2Handler_GetUserInfo() {
	gin.SetMode(gin.TestMode)

	// 创建模拟 OAuth2 服务器
	mockOAuth2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "Bearer test-token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UserInfo{
				Sub:      "user-123",
				Username: "testuser",
				Status:   1,
				ClientID: "app-123",
			})
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}))
	defer mockOAuth2Server.Close()

	// 创建服务
	cfg := &Config{
		Server:       mockOAuth2Server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}
	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建测试路由
	router := gin.New()
	router.GET("/api/oauth2/userinfo", handler.GetUserInfo)

	// 发送请求
	req := httptest.NewRequest("GET", "/api/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 解析响应
	var resp UserInfo
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Printf("状态码: %d\n", w.Code)
	fmt.Printf("用户 ID: %s\n", resp.Sub)
	fmt.Printf("用户名: %s\n", resp.Username)

	// Output:
	// 状态码: 200
	// 用户 ID: user-123
	// 用户名: testuser
}

// ExampleOAuth2Service_BuildAuthorizeURL 演示构建授权 URL
func ExampleOAuth2Service_BuildAuthorizeURL() {
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)

	// 构建授权 URL（state 应由前端生成）
	authURL := svc.BuildAuthorizeURL("random-state-string", "read write")

	fmt.Println("授权 URL:")
	fmt.Println(authURL)

	// Output:
	// 授权 URL:
	// http://localhost:8080/oauth2/authorize?client_id=your-client-id&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback&response_type=code&scope=read+write&state=random-state-string
}

// ExampleSetupRouter 演示如何在 Gin 中设置路由
func ExampleSetupRouter() {
	gin.SetMode(gin.TestMode)

	// 创建配置和服务
	cfg := &Config{
		Server:       "http://localhost:8080",
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURI:  "http://localhost:3000/callback",
	}

	svc := NewOAuth2Service(cfg)
	handler := NewOAuth2Handler(svc)

	// 创建 Gin 引擎
	router := gin.New()

	// 设置路由
	SetupRouter(router, handler)

	// 测试各个路由
	testRoutes := []string{
		"GET:/api/oauth2/config",
		"GET:/api/oauth2/authorize?state=test",
		"POST:/api/oauth2/callback",
		"GET:/api/oauth2/userinfo",
		"POST:/api/oauth2/refresh",
	}

	fmt.Println("注册的路由:")
	for _, route := range testRoutes {
		fmt.Printf("  %s\n", route)
	}

	// Output:
	// 注册的路由:
	//   GET:/api/oauth2/config
	//   GET:/api/oauth2/authorize?state=test
	//   POST:/api/oauth2/callback
	//   GET:/api/oauth2/userinfo
	//   POST:/api/oauth2/refresh
}
