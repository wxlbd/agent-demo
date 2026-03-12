# Eino Agent 开发指南

> 基于 CloudWeGo Eino v0.8.1  
> 最后更新：2026-03-12

---

## 📖 目录

1. [快速开始](#快速开始)
2. [核心概念](#核心概念)
3. [Agent 构建](#agent-构建)
4. [工具开发](#工具开发)
5. [流式处理](#流式处理)
6. [Callback 机制](#callback-机制)
7. [最佳实践](#最佳实践)

---

## 快速开始

### 安装依赖

```bash
go get github.com/cloudwego/eino@v0.8.1
go get github.com/cloudwego/eino-ext/components/model/openai@v0.1.0
```

### 最小化示例

```go
package main

import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/flow/agent/react"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()
    
    // 1. 初始化模型
    model, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL: "https://api.openai.com/v1",
        APIKey:  "your-api-key",
        Model:   "gpt-4o",
    })
    
    // 2. 创建 Agent
    agent, _ := react.NewAgent(ctx, &react.AgentConfig{
        ToolCallingModel: model,
        MaxStep: 10,
    })
    
    // 3. 调用 Agent
    resp, _ := agent.Generate(ctx, []*schema.Message{
        schema.UserMessage("你好，请介绍一下自己"),
    })
    
    println(resp.Content)
}
```

---

## 核心概念

### 1. Agent 类型

Eino 提供多种 Agent 构建方式：

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| **ReAct Agent** | 基于推理和行动的循环 | 需要工具调用的复杂任务 |
| **Workflow** | 有向无环图工作流 | 固定流程的多步骤任务 |
| **Graph** | 通用图执行引擎 | 需要分支、循环的复杂流程 |

### 2. 消息类型

```go
import "github.com/cloudwego/eino/schema"

// 系统消息
schema.SystemMessage("你是一个专业的助手")

// 用户消息
schema.UserMessage("查询天气")

// 助手消息
schema.AssistantMessage("好的，我来查询")

// 工具调用消息
schema.ToolMessage("工具执行结果", "tool_call_id")
```

### 3. 组件层次

```
Application
├── Agent (ReAct/Workflow)
│   ├── Model (LLM)
│   ├── Tools (工具集合)
│   └── MessageModifier (提示词修饰)
└── Callbacks (可观测性)
```

---

## Agent 构建

### 完整配置示例

```go
func InitAgent(ctx context.Context) (*react.Agent, error) {
    // 1. 初始化模型
    chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL: cfg.OpenAIBaseURL,
        APIKey:  cfg.OpenAIKey,
        Model:   cfg.ModelName,
    })
    if err != nil {
        return nil, err
    }

    // 2. 创建工具
    tools := []tool.BaseTool{
        NewDeviceStatusTool(),
        NewDeviceControlTool(),
        NewWeatherTool(),
    }

    // 3. 系统提示词
    systemPrompt := `你是一个专业的 IoT 设备管理助手...`

    // 4. 构建 Agent
    agent, err := react.NewAgent(ctx, &react.AgentConfig{
        ToolCallingModel: chatModel,
        ToolsConfig: compose.ToolsNodeConfig{
            Tools: tools,
        },
        MessageModifier: func(_ context.Context, input []*schema.Message) []*schema.Message {
            return append([]*schema.Message{
                schema.SystemMessage(systemPrompt),
            }, input...)
        },
        MaxStep: 20,
    })

    return agent, err
}
```

### 配置选项说明

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `ToolCallingModel` | `model.ToolCallingChatModel` | 支持工具调用的模型 | 必填 |
| `ToolsConfig` | `compose.ToolsNodeConfig` | 工具配置 | 可选 |
| `MessageModifier` | `MessageModifier` | 消息修饰函数 | 可选 |
| `MaxStep` | `int` | 最大执行步数 | 10 |
| `ToolReturnDirectly` | `bool` | 工具执行后是否直接返回 | false |

---

## 工具开发

### 工具接口

```go
type BaseTool interface {
    // 获取工具信息
    Info(ctx context.Context) (*ToolInfo, error)
    
    // 执行工具
    Run(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
}
```

### 工具实现示例

```go
type DeviceStatusTool struct {
    // 工具状态
}

func (t *DeviceStatusTool) Info(ctx context.Context) (*tool.ToolInfo, error) {
    return &tool.ToolInfo{
        Name:        "device_status",
        Description: "查询 IoT 设备的实时状态信息",
        Params: map[string]*schema.ParameterDefinition{
            "device_id": {
                Type: "string",
                Description: "设备 ID",
                Required: true,
            },
        },
    }, nil
}

func (t *DeviceStatusTool) Run(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    var params struct {
        DeviceID string `json:"device_id"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return "", err
    }
    
    // 执行业务逻辑
    status := queryDeviceStatus(params.DeviceID)
    
    return json.Marshal(status)
}

func NewDeviceStatusTool() (*DeviceStatusTool, error) {
    return &DeviceStatusTool{}, nil
}
```

### 工具注册

```go
toolsConfig := compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{
        deviceTool,
        weatherTool,
        controlTool,
    },
}
```

---

## 流式处理

### Stream 调用

```go
// 流式调用
streamReader, err := agent.Stream(ctx, messages, opts...)
if err != nil {
    return err
}
defer streamReader.Close()

// 读取流式输出
for {
    msg, err := streamReader.Recv()
    if errors.Is(err, io.EOF) {
        break
    }
    if err != nil {
        return err
    }
    
    // 处理消息
    fmt.Print(msg.Content)
}
```

### SSE 响应示例

```go
func HandleChatSSE(c *gin.Context) {
    // 设置 SSE 响应头
    c.Writer.Header().Set("Content-Type", "text/event-stream")
    c.Writer.Header().Set("Cache-Control", "no-cache")
    c.Writer.Header().Set("Connection", "keep-alive")
    
    flusher, _ := c.Writer.(http.Flusher)
    eventChan := make(chan SSEEvent, 100)
    
    go func() {
        defer close(eventChan)
        
        streamReader, _ := agent.Stream(ctx, messages, opts...)
        defer streamReader.Close()
        
        for {
            msg, err := streamReader.Recv()
            if errors.Is(err, io.EOF) {
                eventChan <- SSEEvent{Type: "done", Data: "[DONE]"}
                return
            }
            
            eventChan <- SSEEvent{Type: "message", Data: msg.Content}
        }
    }()
    
    // 发送 SSE 事件
    for event := range eventChan {
        c.SSEvent(event.Type, event.Data)
        flusher.Flush()
    }
}
```

---

## Callback 机制

### Callback 类型

Eino 提供多种 Callback 处理器：

| 类型 | 说明 | 用途 |
|------|------|------|
| `ModelCallbackHandler` | 模型调用回调 | 监控 LLM 调用、token 消耗 |
| `ToolCallbackHandler` | 工具调用回调 | 监控工具执行、参数记录 |
| `GraphCallbackHandler` | 图执行回调 | 监控工作流执行流程 |

### 构建 Callback

```go
// 1. 创建处理器
modelHandler := &ucb.ModelCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
        log.Printf("模型调用开始：%s", info.Name)
        return ctx
    },
    OnEndWithStreamOutput: func(ctx context.Context, info *callbacks.RunInfo,
        output *schema.StreamReader[*model.CallbackOutput]) context.Context {
        log.Printf("模型调用结束")
        return ctx
    },
}

toolHandler := &ucb.ToolCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
        log.Printf("工具调用：%s", info.Name)
        return ctx
    },
    OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
        log.Printf("工具执行完成：%s", info.Name)
        return ctx
    },
}

// 2. 构建 Agent Callback
cb := react.BuildAgentCallback(modelHandler, toolHandler)

// 3. 传递给 Agent
streamReader, err := agent.Stream(ctx, messages,
    agentbase.WithComposeOptions(compose.WithCallbacks(cb)),
)
```

### Callback 事件类型

**Model 事件：**
- `OnStart` - 模型调用开始
- `OnEnd` - 模型调用结束（非流式）
- `OnEndWithStreamOutput` - 模型调用结束（流式）

**Tool 事件：**
- `OnStart` - 工具调用开始
- `OnEnd` - 工具调用结束

---

## 最佳实践

### 1. 错误处理

```go
streamReader, err := agent.Stream(ctx, messages, opts...)
if err != nil {
    // 区分不同类型的错误
    if errors.Is(err, context.DeadlineExceeded) {
        return "请求超时，请重试"
    }
    return fmt.Sprintf("服务错误：%v", err)
}
defer streamReader.Close()
```

### 2. 超时控制

```go
// 设置上下文超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := agent.Generate(ctx, messages)
```

### 3. 日志记录

```go
// 使用 Callback 记录详细日志
logHandler := &ucb.ModelCallbackHandler{
    OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
        logrus.WithFields(logrus.Fields{
            "model": info.Model,
            "input": input,
        }).Info("LLM Call Start")
        return ctx
    },
}
```

### 4. Token 计数

```go
toolHandler := &ucb.ToolCallbackHandler{
    OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
        // 记录工具调用的 token 消耗
        metrics.ToolCallCounter.WithLabelValues(info.Name).Inc()
        return ctx
    },
}
```

### 5. 重试机制

```go
func callWithRetry(ctx context.Context, agent *react.Agent, messages []*schema.Message) (*schema.Message, error) {
    var lastErr error
    for i := 0; i < 3; i++ {
        resp, err := agent.Generate(ctx, messages)
        if err == nil {
            return resp, nil
        }
        lastErr = err
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return nil, lastErr
}
```

---

## 常见问题

### Q1: 工具调用失败怎么办？

**A:** 检查以下几点：
1. 工具的 `Info()` 方法返回的参数定义是否正确
2. 工具的 `Run()` 方法是否正确解析 JSON 参数
3. 确保工具已正确注册到 `ToolsConfig` 中

### Q2: 流式输出乱序？

**A:** 确保：
1. 按顺序调用 `Recv()` 读取流
2. 不要在多个 goroutine 中同时读取同一个 StreamReader
3. 使用 channel 缓冲事件

### Q3: Callback 不触发？

**A:** 检查：
1. 是否使用了正确的 `WithComposeOptions(compose.WithCallbacks(...))`
2. Callback 处理器是否正确构建
3. 导入的包路径是否正确

---

## 参考资源

- 官方文档：https://cloudwego.io/zh/docs/eino/
- GitHub: https://github.com/cloudwego/eino
- 示例项目：https://github.com/cloudwego/eino-examples
- 本文档代码：`iot-agent-demo/` 项目
