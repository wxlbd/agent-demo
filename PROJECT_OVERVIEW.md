# IoT Agent Demo 项目总览

## 🎯 项目目标

基于 CloudWeGo Eino 框架，构建一个**生产级 IoT 设备管理 Agent**，展示：
- ✅ ReAct 智能体架构
- ✅ 工具调用（Tools）
- ✅ 流式输出（SSE）
- ✅ Callback 事件流
- ✅ 提示词工程解耦
- ✅ 上下文管理

---

## 📁 项目文件清单

```
iot-agent-demo/
├── 📄 go.mod                    # Go 模块定义
├── 📄 README.md                 # 项目说明
├── 📄 ARCHITECTURE.md           # 架构设计文档
├── 📄 .env.example              # 环境变量示例
├── 📄 .gitignore                # Git 忽略文件
├── 📄 test.sh                   # 测试脚本
│
├── 📂 cmd/
│   └── server/
│       └── main.go              # 服务入口（1.2KB）
│
├── 📂 internal/
│   ├── agent/
│   │   └── agent.go             # Agent 初始化（2.5KB）
│   ├── tools/
│   │   ├── device.go            # 设备工具（3.0KB）
│   │   └── weather.go           # 天气工具（1.7KB）
│   ├── config/
│   │   └── config.go            # 配置管理（0.5KB）
│   └── handlers/
│       └── chat.go              # HTTP 处理器（4.4KB）
│
├── 📂 prompts/
│   └── system.txt               # 系统提示词（0.4KB）
│
└── 📂 static/
    └── index.html               # Web 测试界面（11.5KB）
```

**总计：** ~26KB 代码 + 文档

---

## 🚀 快速开始

### 步骤 1: 复制环境变量

```bash
cd iot-agent-demo
cp .env.example .env
```

### 步骤 2: 配置 API Key

编辑 `.env` 文件：

```bash
OPENAI_API_KEY=sk-your-api-key-here
```

### 步骤 3: 安装依赖

```bash
go mod tidy
```

### 步骤 4: 运行服务

```bash
go run cmd/server/main.go
```

### 步骤 5: 访问测试

**浏览器：** http://localhost:8080

**命令行：**
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message":"查询设备 CPT-001 的状态"}'
```

---

## 💡 核心功能演示

### 1. 设备状态查询

**用户：** 查询设备 CPT-001 的状态

**Agent:**
```
1. 调用 query_device_status 工具
2. 返回设备数据：
   - 在线状态：在线
   - 锥尖阻力 (qc): 15.2 MPa
   - 摩阻力 (fs): 85.3 kPa
   - 报警状态：正常
```

### 2. 天气查询 + 维护建议

**用户：** 查询设备状态并结合天气给出维护建议

**Agent:**
```
1. 调用 query_device_status 工具
2. 调用 get_weather 工具
3. 综合分析：
   - 设备状态正常
   - 北京今天晴天，25°C
   - 建议：适合户外作业，注意防晒
```

### 3. 设备控制

**用户：** 重启设备 CPT-001

**Agent:**
```
1. 调用 control_device 工具
2. 确认操作风险
3. 执行重启指令
4. 返回执行结果
```

---

## 📊 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **框架** | Eino v0.7.0 | CloudWeGo Agent 框架 |
| **语言** | Go 1.21+ | 后端开发语言 |
| **Web** | Gin v1.9.1 | HTTP 框架 |
| **LLM** | OpenAI GPT-4o | 大模型（可替换） |
| **流式** | SSE | Server-Sent Events |
| **工具** | utils.InferTool | 自动 JSON Schema 生成 |

---

## 🔧 可扩展能力

### 添加新工具

在 `internal/tools/` 创建新文件：

```go
func NewCustomTool() (tool.InvokableTool, error) {
    return utils.InferTool(
        "custom_tool",
        "工具描述",
        func(ctx context.Context, req *Request) (string, error) {
            // 实现逻辑
        },
    )
}
```

### 添加记忆持久化

使用 Eino CheckPoint:

```go
compose.WithCheckPointStore(redisStore)
compose.WithCheckPointID(sessionID)
```

### 添加人工审批

使用 Eino Interrupt:

```go
compose.StatefulInterrupt(ctx, context, state)
compose.ResumeWithData(ctx, interruptID, data)
```

---

## 📝 学习路径

1. **阅读学习笔记** → `../learning/eino-agent-guide.md`
2. **理解架构设计** → `ARCHITECTURE.md`
3. **运行 Demo** → 按快速开始步骤操作
4. **修改扩展** → 添加工具、修改提示词
5. **生产部署** → 参考架构文档的部署建议

---

## 🐛 常见问题

### Q1: Agent 未初始化

**原因:** OPENAI_API_KEY 未配置或无效

**解决:** 检查 `.env` 文件，确保 API Key 正确

### Q2: 工具调用失败

**原因:** 工具参数校验失败

**解决:** 检查请求参数是否符合 JSON Schema

### Q3: SSE 连接中断

**原因:** 反向代理缓冲设置问题

**解决:** Nginx 配置 `proxy_buffering off`

---

## 📚 参考资料

- [Eino 官方文档](https://www.cloudwego.io/docs/eino/)
- [Eino GitHub](https://github.com/cloudwego/eino)
- [学习笔记](../learning/eino-agent-guide.md)
- [CloudWeGo 官网](https://www.cloudwego.io/)

---

## 👥 作者

- 整理时间：2026-03-11
- 基于：Eino 智能体开发实战指南
- 用途：IoT 平台 AI 集成 Demo

---

**🎉 祝使用愉快！有问题随时反馈。**
