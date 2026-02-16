package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/aiqoder/my-go-tools/gin-static-server"
	"github.com/gin-gonic/gin"
)

//go:embed dist
var assets embed.FS

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 使用 embed.FS 服务静态文件
	ginstatic.NewEmbed(r, assets,
		ginstatic.WithPrefix(""),          // 根路径访问
		ginstatic.WithEmbedRoot("dist"),   // 指定 embed 的子目录
		ginstatic.WithSPA("index.html"),   // SPA 回退支持
		ginstatic.WithGzip(6),             // Gzip 压缩
		ginstatic.WithCache(50*1024*1024, 100), // 50MB 缓存
		ginstatic.WithCacheControl("public, max-age=31536000"),
		ginstatic.WithETag(),
	)

	// 打印配置信息
	fmt.Println("========================================")
	fmt.Println("  Vue + Gin Static Server (Embed Demo)")
	fmt.Println("========================================")
	fmt.Println("  静态文件: embed.FS")
	fmt.Println("  服务地址: http://localhost:8080")
	fmt.Println("  SPA 模式: 已启用")
	fmt.Println("  Gzip 压缩: 已启用")
	fmt.Println("  内存缓存: 50MB")
	fmt.Println("========================================")

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
