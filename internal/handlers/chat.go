package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	ucb "github.com/cloudwego/eino/utils/callbacks"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"iot-agent-demo/internal/agent"
)

// ChatRequest 聊天请求体
type ChatRequest struct {
	Message  string `json:"message" binding:"required"`
	SessionID string `json:"session_id"`
}

// HandleChatSSE 处理 SSE 流式聊天请求
func HandleChatSSE(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	// 创建事件通道
	eventChan := make(chan SSEEvent, 100)
	defer close(eventChan)

	// 启动后台 goroutine 处理 Agent 响应
	go func() {
		defer close(eventChan)

		ctx := c.Request.Context()
		reactAgent := agent.GetAgent()

		if reactAgent == nil {
			eventChan <- SSEEvent{Type: "error", Data: "Agent 未初始化"}
			return
		}

		// 构建 Callback 处理器
		modelHandler := &ucb.ModelCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
				eventChan <- SSEEvent{Type: "model_start", Data: "开始思考..."}
				return ctx
			},
			OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo,
				output *schema.StreamReader[*model.CallbackOutput]) context.Context {
				output.Close()
				eventChan <- SSEEvent{Type: "model_end", Data: "思考完成"}
				return ctx
			},
		}

		toolHandler := &ucb.ToolCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
				eventChan <- SSEEvent{Type: "tool_start", Data: info.Name}
				return ctx
			},
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
				eventChan <- SSEEvent{Type: "tool_done", Data: info.Name}
				return ctx
			},
		}

		cb := react.BuildAgentCallback(modelHandler, toolHandler)

		// 创建用户消息
		userMessage := schema.UserMessage(req.Message)

		// 启动流式响应
		streamReader, err := reactAgent.Stream(
			ctx,
			[]*schema.Message{userMessage},
			react.WithComposeOptions(compose.WithCallbacks(cb)),
		)

		if err != nil {
			eventChan <- SSEEvent{Type: "error", Data: "Agent 启动失败：" + err.Error()}
			return
		}
		defer streamReader.Close()

		// 读取流式输出
		for {
			msg, err := streamReader.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					eventChan <- SSEEvent{Type: "done", Data: "[DONE]"}
					return
				}
				eventChan <- SSEEvent{Type: "error", Data: "读取失败：" + err.Error()}
				return
			}

			if msg.Content != "" {
				eventChan <- SSEEvent{Type: "message", Data: msg.Content}
			}
		}
	}()

	// 发送 SSE 事件
	for {
		event, ok := <-eventChan
		if !ok {
			return
		}

		c.SSEvent(event.Type, event.Data)
		flusher.Flush()

		// 客户端断开检测
		if c.Writer.Status() == http.StatusNoContent {
			return
		}
	}
}

// SSEEvent SSE 事件结构
type SSEEvent struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// HandleChat 处理普通聊天请求（非流式）
func HandleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	reactAgent := agent.GetAgent()

	if reactAgent == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent 未初始化"})
		return
	}

	userMessage := schema.UserMessage(req.Message)

	response, err := reactAgent.Generate(ctx, []*schema.Message{userMessage})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   response.Content,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// HandleHealth 健康检查接口
func HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	})
}
