package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"iot-agent-demo/internal/agent"
	"iot-agent-demo/internal/config"
	"iot-agent-demo/internal/handlers"
)

func main() {
	log.Println("🚀 IoT Agent Demo 启动中...")

	// 1. 加载配置
	cfg := config.Load()
	log.Printf("配置加载完成：模型=%s, 端口=%s", cfg.ModelName, cfg.ServerPort)

	// 2. 初始化 Agent
	ctx := context.Background()
	if err := agent.InitAgent(ctx, cfg); err != nil {
		log.Fatalf("Agent 初始化失败：%v", err)
	}
	log.Println("✅ Agent 初始化完成")

	// 3. 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 4. 注册路由
	r.GET("/health", handlers.HandleHealth)
	r.POST("/api/v1/chat", handlers.HandleChat)         // 普通聊天
	r.POST("/api/v1/chat/stream", handlers.HandleChatSSE) // 流式聊天
	r.GET("/", func(c *gin.Context) {
		c.File("static/index.html")
	})
	r.StaticFS("/static", http.Dir("static"))

	// 5. 启动服务
	addr := ":" + cfg.ServerPort
	log.Printf("📡 服务启动在 http://localhost%s", addr)
	log.Println("📝 测试命令:")
	log.Printf("   curl -X POST http://localhost%s/api/v1/chat/stream \\", addr)
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{\"message\":\"查询设备 CPT-001 的状态\"}'")
	log.Println("🌐 浏览器访问：http://localhost:" + cfg.ServerPort)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败：%v", err)
	}
}
