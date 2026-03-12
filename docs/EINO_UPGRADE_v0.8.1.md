# Eino v0.8.1 升级指南

> 文档版本：1.0  
> 更新时间：2026-03-12  
> 适用版本：Eino v0.8.1

---

## 📋 概述

本文档记录了从 Eino 早期版本升级到 v0.8.1 的关键变更和代码修正。

---

## 🔧 关键 API 变更

### 1. Callback 使用方式变更

**变更说明：** `react.WithCallbacks()` 已废弃，需要使用 `agent.WithComposeOptions(compose.WithCallbacks())`

**旧代码（❌ 错误）：**
```go
cb := react.BuildAgentCallback(modelHandler, toolHandler)

streamReader, err := reactAgent.Stream(
    ctx,
    []*schema.Message{userMessage},
    react.WithCallbacks(cb), // ❌ 此 API 不存在
)
```

**新代码（✅ 正确）：**
```go
import (
    agentbase "github.com/cloudwego/eino/flow/agent"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/flow/agent/react"
)

cb := react.BuildAgentCallback(modelHandler, toolHandler)

streamReader, err := reactAgent.Stream(
    ctx,
    []*schema.Message{userMessage},
    agentbase.WithComposeOptions(compose.WithCallbacks(cb)), // ✅ 正确方式
)
```

**原因：**
- `WithComposeOptions` 位于 `github.com/cloudwego/eino/flow/agent` 包中
- 需要通过此函数将 compose 层的选项传递给 agent 层

---

### 2. 包导入调整

**需要导入的包：**
```go
import (
    "github.com/cloudwego/eino/callbacks"
    agentbase "github.com/cloudwego/eino/flow/agent"  // 用于 WithComposeOptions
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/flow/agent/react"
)
```

---

## 📝 完整示例

### 流式聊天 Handler

```go
package handlers

import (
    "context"
    "errors"
    "io"
    "net/http"

    "github.com/cloudwego/eino/callbacks"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/tool"
    agentbase "github.com/cloudwego/eino/flow/agent"
    "github.com/cloudwego/eino/flow/agent/react"
    "github.com/cloudwego/eino/schema"
    ucb "github.com/cloudwego/eino/utils/callbacks"
    "github.com/gin-gonic/gin"
)

func HandleChatSSE(c *gin.Context) {
    // ... 省略请求解析代码 ...

    go func() {
        defer close(eventChan)

        ctx := c.Request.Context()
        reactAgent := agent.GetAgent()

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

        // 启动流式响应
        streamReader, err := reactAgent.Stream(
            ctx,
            []*schema.Message{userMessage},
            agentbase.WithComposeOptions(compose.WithCallbacks(cb)),
        )

        // ... 省略流式处理代码 ...
    }()

    // ... 省略 SSE 响应代码 ...
}
```

---

## 🚨 常见错误

### 错误 1：undefined: react.WithCallbacks

**错误信息：**
```
undefined: react.WithCallbacks
```

**解决方案：**
使用 `agentbase.WithComposeOptions(compose.WithCallbacks(cb))` 替代

---

### 错误 2：cannot use compose.WithCallbacks() as agent.AgentOption

**错误信息：**
```
cannot use compose.WithCallbacks(cb) (value of struct type compose.Option) 
as "github.com/cloudwego/eino/flow/agent".AgentOption value
```

**解决方案：**
需要外层包裹 `agentbase.WithComposeOptions()`

---

### 错误 3：undefined: callbacks

**错误信息：**
```
undefined: callbacks
```

**解决方案：**
确保导入 `"github.com/cloudwego/eino/callbacks"` 包

---

## 📚 相关资源

- Eino 官方文档：https://cloudwego.io/zh/docs/eino/
- GitHub 仓库：https://github.com/cloudwego/eino
- 本文档示例代码：`internal/handlers/chat.go`

---

## 🔄 版本兼容性

| Eino 版本 | Callback API | 状态 |
|-----------|-------------|------|
| v0.7.0 | `react.WithCallbacks()` | ✅ 支持 |
| v0.8.0 | `agent.WithComposeOptions()` | ✅ 支持 |
| v0.8.1 | `agent.WithComposeOptions()` | ✅ 支持 |

**建议：** 始终使用 `go.mod` 中指定的版本，并在升级前查阅官方变更日志。
