# Eino API 参考手册

> 基于 Eino v0.8.1  
> 最后更新：2026-03-12

---

## 📋 目录

1. [Agent API](#agent-api)
2. [Model API](#model-api)
3. [Tool API](#tool-api)
4. [Schema API](#schema-api)
5. [Callback API](#callback-api)
6. [Compose API](#compose-api)

---

## Agent API

### react.Agent

#### 创建 Agent

```go
func NewAgent(ctx context.Context, config *AgentConfig) (*Agent, error)
```

**配置参数：**

```go
type AgentConfig struct {
    ToolCallingModel model.ToolCallingChatModel  // 必填：支持工具调用的模型
    ToolsConfig      compose.ToolsNodeConfig     // 可选：工具配置
    MessageModifier  MessageModifier             // 可选：消息修饰器
    MaxStep          int                         // 可选：最大执行步数，默认 10
}
```

#### 调用方法

**Generate（非流式）：**
```go
func (r *Agent) Generate(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error)
```

**Stream（流式）：**
```go
func (r *Agent) Stream(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error)
```

**Collect（收集式）：**
```go
func (r *Agent) Collect(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*MessageFuture, error)
```

#### Agent 选项

```go
// 使用 compose 选项
func WithComposeOptions(opts ...compose.Option) AgentOption

// 使用模型选项
func WithChatModelOptions(opts ...model.Option) AgentOption

// 使用工具选项
func WithToolOptions(opts ...tool.Option) AgentOption

// 动态添加工具
func WithTools(ctx context.Context, tools ...tool.BaseTool) ([]AgentOption, error)

// 指定工具列表
func WithToolList(tools ...tool.BaseTool) AgentOption

// 消息 Future（用于 Collect 模式）
func WithMessageFuture() (AgentOption, MessageFuture)
```

---

## Model API

### openai.ChatModel

#### 创建模型

```go
func NewChatModel(ctx context.Context, config *ChatModelConfig) (*ChatModel, error)
```

**配置参数：**

```go
type ChatModelConfig struct {
    BaseURL    string  // API 基础 URL
    APIKey     string  // API 密钥
    Model      string  // 模型名称
    Type       string  // 模型类型（可选）
}
```

#### 模型接口

```go
type ChatModel interface {
    // 生成响应
    Generate(ctx context.Context, input []*schema.Message, opts ...Option) (*schema.Message, error)
    
    // 流式生成
    Stream(ctx context.Context, input []*schema.Message, opts ...Option) (*schema.StreamReader[*schema.Message], error)
    
    // 绑定工具
    BindTools(tools []tool.BaseTool) error
}
```

#### 模型选项

```go
// 设置温度
func WithTemperature(temp float64) Option

// 设置最大 token 数
func WithMaxTokens(tokens int) Option

// 设置 TopP
func WithTopP(p float64) Option

// 设置停止词
func WithStopWords(words []string) Option
```

---

## Tool API

### 工具接口

```go
type BaseTool interface {
    // 获取工具信息
    Info(ctx context.Context) (*ToolInfo, error)
    
    // 执行工具
    Run(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
}
```

### ToolInfo 结构

```go
type ToolInfo struct {
    Name        string                          `json:"name"`
    Description string                          `json:"description"`
    Params      map[string]*ParameterDefinition `json:"parameters"`
}
```

### ParameterDefinition 结构

```go
type ParameterDefinition struct {
    Type        string      `json:"type"`        // string, number, boolean, array, object
    Description string      `json:"description"` // 参数描述
    Required    bool        `json:"required"`    // 是否必填
    Enum        []string    `json:"enum"`        // 枚举值（可选）
    Default     interface{} `json:"default"`     // 默认值（可选）
}
```

### 工具实现模板

```go
type MyTool struct {
    // 工具状态
}

func (t *MyTool) Info(ctx context.Context) (*tool.ToolInfo, error) {
    return &tool.ToolInfo{
        Name:        "my_tool",
        Description: "工具描述",
        Params: map[string]*schema.ParameterDefinition{
            "param1": {
                Type:        "string",
                Description: "参数 1 描述",
                Required:    true,
            },
        },
    }, nil
}

func (t *MyTool) Run(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    var params struct {
        Param1 string `json:"param1"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return "", err
    }
    
    // 执行逻辑
    result := doSomething(params.Param1)
    
    return json.Marshal(result)
}
```

---

## Schema API

### 消息类型

```go
// 系统消息
func SystemMessage(content string) *Message

// 用户消息
func UserMessage(content string) *Message

// 助手消息
func AssistantMessage(content string) *Message

// 工具调用消息
func ToolMessage(content string, toolCallID string) *Message
```

### Message 结构

```go
type Message struct {
    Role       string      `json:"role"`       // system, user, assistant, tool
    Content    string      `json:"content"`    // 消息内容
    Name       string      `json:"name"`       // 消息名称（可选）
    ToolCalls  []ToolCall  `json:"tool_calls"` // 工具调用列表（助手消息）
    ToolCallID string      `json:"tool_call_id"` // 工具调用 ID（工具消息）
}
```

### ToolCall 结构

```go
type ToolCall struct {
    ID       string       `json:"id"`        // 工具调用 ID
    Type     string       `json:"type"`      // 调用类型（function）
    Function FunctionCall `json:"function"`  // 函数调用信息
}

type FunctionCall struct {
    Name      string `json:"name"`      // 函数名称
    Arguments string `json:"arguments"` // JSON 格式的参数
}
```

### StreamReader

```go
type StreamReader[T any] interface {
    // 接收下一个元素
    Recv() (T, error)
    
    // 关闭流
    Close()
    
    // 检查是否已完成
    IsDone() bool
}
```

---

## Callback API

### Callback 处理器

#### ModelCallbackHandler

```go
type ModelCallbackHandler struct {
    OnStart                 func(ctx context.Context, info *RunInfo, input *model.CallbackInput) context.Context
    OnEnd                   func(ctx context.Context, info *RunInfo, output *model.CallbackOutput) context.Context
    OnEndWithStreamOutput   func(ctx context.Context, info *RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context
}
```

#### ToolCallbackHandler

```go
type ToolCallbackHandler struct {
    OnStart   func(ctx context.Context, info *RunInfo, input *tool.CallbackInput) context.Context
    OnEnd     func(ctx context.Context, info *RunInfo, output *tool.CallbackOutput) context.Context
}
```

### RunInfo 结构

```go
type RunInfo struct {
    Name      string `json:"name"`      // 组件名称
    Type      string `json:"type"`      // 组件类型
    Component string `json:"component"` // 组件类型标识
    Model     string `json:"model"`     // 模型名称（如果是模型调用）
}
```

### 构建 Agent Callback

```go
func BuildAgentCallback(
    modelHandler *template.ModelCallbackHandler,
    toolHandler *template.ToolCallbackHandler,
) callbacks.Handler
```

### 使用 Callback

```go
// 1. 创建处理器
modelHandler := &ucb.ModelCallbackHandler{
    OnStart: func(ctx context.Context, info *RunInfo, input *model.CallbackInput) context.Context {
        log.Printf("模型调用：%s", info.Name)
        return ctx
    },
}

toolHandler := &ucb.ToolCallbackHandler{
    OnStart: func(ctx context.Context, info *RunInfo, input *tool.CallbackInput) context.Context {
        log.Printf("工具调用：%s", info.Name)
        return ctx
    },
}

// 2. 构建 Callback
cb := react.BuildAgentCallback(modelHandler, toolHandler)

// 3. 传递给 Agent
streamReader, err := agent.Stream(ctx, messages,
    agentbase.WithComposeOptions(compose.WithCallbacks(cb)),
)
```

---

## Compose API

### ToolsNodeConfig

```go
type ToolsNodeConfig struct {
    Tools []tool.BaseTool  // 工具列表
}
```

### Callback 选项

```go
// 添加 Callback 处理器
func WithCallbacks(cbs ...callbacks.Handler) Option
```

### Graph 编译选项

```go
// 设置图名称
func WithGraphName(graphName string) GraphCompileOption

// 添加编译回调
func WithGraphCompileCallbacks(cbs ...GraphCompileCallback) GraphCompileOption

// 设置 FanIn 合并配置
func WithFanInMergeConfig(confs map[string]FanInMergeConfig) GraphCompileOption
```

---

## 完整示例

### 创建并调用 Agent

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/cloudwego/eino-ext/components/model/openai"
    agentbase "github.com/cloudwego/eino/flow/agent"
    "github.com/cloudwego/eino/flow/agent/react"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    ucb "github.com/cloudwego/eino/utils/callbacks"
)

func main() {
    ctx := context.Background()
    
    // 1. 创建模型
    model, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        BaseURL: "https://api.openai.com/v1",
        APIKey:  "sk-xxx",
        Model:   "gpt-4o",
    })
    
    // 2. 创建工具
    tools := []tool.BaseTool{
        NewWeatherTool(),
    }
    
    // 3. 创建 Callback
    cb := react.BuildAgentCallback(
        &ucb.ModelCallbackHandler{},
        &ucb.ToolCallbackHandler{},
    )
    
    // 4. 创建 Agent
    agent, _ := react.NewAgent(ctx, &react.AgentConfig{
        ToolCallingModel: model,
        ToolsConfig: compose.ToolsNodeConfig{
            Tools: tools,
        },
        MaxStep: 10,
    })
    
    // 5. 调用 Agent
    messages := []*schema.Message{
        schema.UserMessage("北京今天天气怎么样？"),
    }
    
    resp, err := agent.Generate(ctx, messages,
        agentbase.WithComposeOptions(compose.WithCallbacks(cb)),
    )
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(resp.Content)
}
```

---

## 错误处理

### 常见错误类型

```go
// 上下文超时
if errors.Is(err, context.DeadlineExceeded) {
    // 处理超时
}

// 上下文取消
if errors.Is(err, context.Canceled) {
    // 处理取消
}

// EOF（流式读取结束）
if errors.Is(err, io.EOF) {
    // 处理流结束
}
```

### 错误包装

```go
if err != nil {
    return fmt.Errorf("agent generate failed: %w", err)
}
```

---

## 参考资源

- 官方文档：https://cloudwego.io/zh/docs/eino/
- GitHub: https://github.com/cloudwego/eino
- 示例代码：https://github.com/cloudwego/eino-examples
