# IoT Agent Demo 测试脚本

## 1. 健康检查

```bash
curl http://localhost:8080/health
```

**预期响应:**
```json
{
  "status": "ok",
  "timestamp": "2026-03-11 22:30:00"
}
```

---

## 2. 普通聊天接口（非流式）

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "查询设备 CPT-001 的状态"}'
```

**预期响应:**
```json
{
  "message": "设备 CPT-001 状态报告:\n- 在线状态：在线\n- 锥尖阻力 (qc): 15.2 MPa\n...",
  "timestamp": "2026-03-11 22:30:00"
}
```

---

## 3. 流式聊天接口（SSE）

```bash
curl -N -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "查询设备 CPT-001 的状态，并结合天气给出维护建议"}'
```

**预期 SSE 事件流:**
```
event: model_start
data: 开始思考...

event: tool_start
data: query_device_status

event: tool_done
data: query_device_status

event: tool_start
data: get_weather

event: tool_done
data: get_weather

event: message
data: 正在为您查询...

event: message
data: 设备 CPT-001 状态正常...

event: done
data: [DONE]
```

---

## 4. 测试场景示例

### 场景 1: 简单设备查询
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "查询设备 CPT-001 的状态"}'
```

### 场景 2: 天气查询
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "北京今天天气如何？"}'
```

### 场景 3: 综合分析
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "查询设备状态并结合天气给出维护建议"}'
```

### 场景 4: 设备控制
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"message": "重启设备 CPT-001"}'
```

---

## 5. 浏览器测试

打开浏览器访问：http://localhost:8080

使用内置的 Web 界面进行交互式测试。

---

## 6. 常见问题排查

### 问题 1: Agent 未初始化
```
错误：Agent 未初始化
```
**解决:** 检查 OPENAI_API_KEY 是否正确配置

### 问题 2: 模型调用失败
```
错误：初始化模型失败
```
**解决:** 检查 OPENAI_BASE_URL 和 API Key 是否有效

### 问题 3: SSE 连接中断
```
错误：streaming unsupported
```
**解决:** 确保反向代理（如 Nginx）配置了正确的缓冲设置

---

## 7. 性能测试

### 并发测试（使用 ab 工具）
```bash
ab -n 100 -c 10 -p test.json -T application/json \
  http://localhost:8080/api/v1/chat
```

### 响应时间监控
查看日志中的请求处理时间。
