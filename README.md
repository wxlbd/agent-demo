# IoT Agent Demo - 基于 Eino 框架的设备查询智能体

## 📦 项目结构

```
iot-agent-demo/
├── cmd/
│   └── server/
│       └── main.go          # 服务入口
├── internal/
│   ├── agent/
│   │   └── agent.go         # Agent 构建与初始化
│   ├── tools/
│   │   ├── device.go        # 设备查询工具
│   │   └── weather.go       # 天气查询工具
│   ├── config/
│   │   └── config.go        # 配置管理
│   └── handlers/
│       └── chat.go          # HTTP 处理器
├── prompts/
│   └── system.txt           # 系统提示词模板
├── static/
│   └── index.html           # Web 测试界面
├── .env.example             # 环境变量示例
├── .gitignore
├── go.mod
├── test.sh                  # 测试脚本
├── ARCHITECTURE.md          # 架构设计文档
└── README.md
```

## 🚀 快速开始

### 1. 安装依赖

```bash
cd iot-agent-demo
go mod tidy
```

### 2. 配置环境变量

```bash
# .env 文件
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=sk-your-api-key
MODEL_NAME=gpt-4o
SERVER_PORT=8080
```

### 3. 运行服务

```bash
go run cmd/server/main.go
```

### 4. 测试 API

```bash
# SSE 流式接口
curl -N http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "查询设备 CPT-001 的状态"}'
```

## 📡 API 接口

### POST /api/v1/chat/stream

**请求体：**
```json
{
  "message": "帮我查一下北京的设备状态",
  "session_id": "user-123"
}
```

**SSE 事件流：**
```
event: message
data: 正在查询设备状态...

event: tool_start
data: query_device_status

event: tool_done
data: query_device_status

event: message
data: 设备 CPT-001 状态正常...

event: done
data: [DONE]
```

## 🛠️ 核心功能

- ✅ ReAct Agent 架构（思考→行动→观察）
- ✅ 工具调用（设备查询、天气查询）
- ✅ 流式输出（SSE）
- ✅ Callback 事件流（工具状态实时推送）
- ✅ 提示词模板外部化
- ✅ 上下文管理（滑动窗口裁剪）
- ✅ 错误处理与降级

## 📝 示例对话

**用户：** 查询设备 CPT-001 的状态，并结合天气给出维护建议

**Agent：**
1. 调用 `query_device_status` 获取设备数据
2. 调用 `get_weather` 获取北京天气
3. 综合分析并给出建议

---

## 🔗 相关文档

- [Eino 官方文档](https://www.cloudwego.io/docs/eino/)
- [学习笔记](../learning/eino-agent-guide.md)
