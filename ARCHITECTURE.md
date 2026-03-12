# IoT Agent Demo 架构设计文档

## 1. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Web 界面    │  │  API 客户端  │  │  移动端     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                        API 网关层                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Gin HTTP Server (8080)                 │   │
│  │  GET  /              - Web 界面                     │   │
│  │  GET  /health        - 健康检查                     │   │
│  │  POST /api/v1/chat   - 普通聊天                     │   │
│  │  POST /api/v1/chat/stream - 流式聊天 (SSE)          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Agent 核心层                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Eino ReAct Agent                       │   │
│  │  ┌──────────────┐  ┌──────────────┐                │   │
│  │  │ ChatModel    │  │ ToolsNode    │                │   │
│  │  │ (GPT-4o)     │  │ (工具调度)    │                │   │
│  │  └──────────────┘  └──────────────┘                │   │
│  │                                                     │   │
│  │  ┌──────────────────────────────────────────────┐  │   │
│  │  │  MessageModifier / MessageRewriter           │  │   │
│  │  │  - 系统提示注入                              │  │   │
│  │  │  - 上下文管理                                │  │   │
│  │  └──────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       工具层                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Device Tool  │  │ Weather Tool │  │ Control Tool │     │
│  │ (设备查询)    │  │ (天气查询)    │  │ (设备控制)    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      外部服务                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  IoT 平台 API │  │  天气 API    │  │  LLM API     │     │
│  │  (模拟)      │  │  (模拟)      │  │  (OpenAI)    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 核心流程

### 2.1 ReAct 执行流程

```
用户请求
   │
   ▼
┌─────────────────┐
│ 1. 感知输入     │  接收用户消息
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ 2. 思考 (Thought)│  LLM 分析是否需要调用工具
└─────────────────┘
   │
   ├───────┐
   │       │ 需要工具
   ▼       ▼
┌─────────────────┐
│ 3. 行动 (Action) │  执行工具调用
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ 4. 观察 (Observation)│  获取工具返回结果
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ 5. 继续推理     │  基于新信息继续或结束
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ 6. 输出回答     │  返回最终结果
└─────────────────┘
```

### 2.2 流式输出流程 (SSE)

```
Client                              Server
  │                                   │
  │──── POST /chat/stream ───────────▶│
  │                                   │
  │                                   │ [创建 SSE 连接]
  │                                   │
  │◄─── event: model_start ───────────│
  │     data: 开始思考...              │
  │                                   │
  │◄─── event: tool_start ────────────│
  │     data: query_device_status     │
  │                                   │ [执行工具]
  │◄─── event: tool_done ─────────────│
  │     data: query_device_status     │
  │                                   │
  │◄─── event: message ───────────────│
  │     data: 正在查询...             │ [流式输出]
  │◄─── event: message ───────────────│
  │     data: 设备状态正常...         │
  │◄─── event: message ───────────────│
  │     data: 建议穿轻薄外套...       │
  │                                   │
  │◄─── event: done ──────────────────│
  │     data: [DONE]                  │
  │                                   │ [关闭连接]
```

---

## 3. 数据模型

### 3.1 消息结构

```go
type Message struct {
    Role    string `json:"role"`    // user/assistant/system
    Content string `json:"content"` // 消息内容
}
```

### 3.2 设备数据结构

```go
type DeviceData struct {
    DeviceID    string    `json:"device_id"`
    TenantID    string    `json:"tenant_id"`
    Status      string    `json:"status"`      // 在线/离线
    Qc          float64   `json:"qc"`          // 锥尖阻力
    Fs          float64   `json:"fs"`          // 摩阻力
    U2          float64   `json:"u2"`          // 孔隙水压力
    Temperature float64   `json:"temperature"` // 温度
    LastUpdate  time.Time `json:"last_update"`
    AlarmStatus string    `json:"alarm_status"`
}
```

### 3.3 天气数据结构

```go
type WeatherData struct {
    City        string `json:"city"`
    Condition   string `json:"condition"`   // 晴/多云/雨
    Temperature int    `json:"temperature"` // °C
    Humidity    int    `json:"humidity"`    // %
    Wind        string `json:"wind"`        // 风力等级
    AQI         string `json:"aqi"`         // 空气质量
}
```

---

## 4. 配置管理

### 4.1 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `OPENAI_BASE_URL` | LLM API 地址 | `https://api.openai.com/v1` |
| `OPENAI_API_KEY` | API Key | (必填) |
| `MODEL_NAME` | 模型名称 | `gpt-4o` |
| `SERVER_PORT` | 服务端口 | `8080` |

### 4.2 提示词配置

外部文件：`prompts/system.txt`

支持动态加载，修改后无需重启服务。

---

## 5. 扩展点

### 5.1 添加工具

1. 在 `internal/tools/` 创建新工具文件
2. 实现 `tool.InvokableTool` 接口
3. 在 `internal/agent/agent.go` 中注册

```go
// 示例：添加数据库查询工具
func NewDatabaseTool() (tool.InvokableTool, error) {
    return utils.InferTool(
        "query_database",
        "查询业务数据库",
        func(ctx context.Context, req *QueryRequest) (string, error) {
            // 实现查询逻辑
        },
    )
}
```

### 5.2 添加记忆持久化

使用 Eino CheckPoint:

```go
import "github.com/cloudwego/eino/compose"

// 编译时绑定存储
runnable, err := g.Compile(ctx,
    compose.WithCheckPointStore(redisStore),
)

// 运行时关联会话
output, err := runnable.Invoke(ctx, input,
    compose.WithCheckPointID(sessionID),
)
```

### 5.3 添加人工审批

使用 Eino Interrupt:

```go
// 在敏感操作前中断
if requiresApproval {
    return "", compose.StatefulInterrupt(ctx,
        map[string]any{"reason": "需要审批"},
        &ApprovalState{Operation: op},
    )
}
```

---

## 6. 生产部署建议

### 6.1 容器化

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o iot-agent ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/iot-agent .
COPY --from=builder /app/prompts ./prompts
EXPOSE 8080
CMD ["./iot-agent"]
```

### 6.2 负载均衡

使用 Nginx 反向代理：

```nginx
upstream iot_agent {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://iot_agent;
        proxy_buffering off;  # SSE 需要关闭缓冲
        proxy_cache off;
    }
}
```

### 6.3 监控指标

- 请求响应时间 (P95/P99)
- 工具调用成功率
- LLM API 调用次数
- SSE 连接数
- 错误率

---

## 7. 安全考虑

### 7.1 API 鉴权

生产环境建议添加：

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if !validateToken(token) {
            c.AbortWithStatus(401)
            return
        }
        c.Next()
    }
}
```

### 7.2 工具权限控制

```go
// 白名单控制
allowedTools := map[string]bool{
    "query_device_status": true,
    "get_weather": true,
}

if !allowedTools[toolName] {
    return nil, fmt.Errorf("tool not allowed")
}
```

### 7.3 输入校验

```go
// 限制消息长度
if len(req.Message) > 2000 {
    return errors.New("message too long")
}

// 过滤敏感词
if containsSensitiveWords(req.Message) {
    return errors.New("invalid message")
}
```

---

## 8. 性能优化

### 8.1 连接池

```go
// LLM Client 复用
var globalClient *http.Client = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}
```

### 8.2 缓存

```go
// 工具结果缓存（相同参数短时间内不重复调用）
var cache = sync.Map{}

func getCachedResult(key string) (string, bool) {
    if val, ok := cache.Load(key); ok {
        return val.(string), true
    }
    return "", false
}
```

### 8.3 并发控制

```go
// 限制并发请求数
var semaphore = make(chan struct{}, 10)

func handleRequest() {
    semaphore <- struct{}{}
    defer func() { <-semaphore }()
    // 处理请求
}
```
