//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net/http"

	"github.com/aiqoder/my-go-tools/gin-static-server"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 引擎
	r := gin.New()
	r.Use(gin.Recovery())

	// 禁用默认的 NoRoute 处理
	r.NoRoute(func(c *gin.Context) {
		// 让请求继续传递，不做任何处理
		c.Next()
	})

	// 使用静态文件扩展名中间件（解决路由冲突问题）
	// 只拦截以 .js, .css, .html 等前端资源结尾的请求
	r.Use(ginstatic.StaticFileExtsMiddleware("./testdata/static"))

	// 定义 API 路由（不再被 /*path 拦截）
	r.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API response - users list",
			"data":    []string{"user1", "user2", "user3"},
		})
	})

	r.GET("/api/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "API response - products list",
			"data":    []string{"product1", "product2"},
		})
	})

	// 其他路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动服务器
	fmt.Println("服务器启动: http://localhost:8080")
	fmt.Println("")
	fmt.Println("测试路由:")
	fmt.Println("  - 静态文件: http://localhost:8080/app.js")
	fmt.Println("  - 静态文件: http://localhost:8080/style.css")
	fmt.Println("  - 静态文件: http://localhost:8080/index.html")
	fmt.Println("  - API 路由:  http://localhost:8080/api/users")
	fmt.Println("  - API 路由:  http://localhost:8080/api/products")
	fmt.Println("  - 健康检查:  http://localhost:8080/health")
	fmt.Println("")
	fmt.Println("按 Ctrl+C 停止服务器")

	r.Run(":8080")
}
