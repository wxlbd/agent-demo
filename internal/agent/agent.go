package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"iot-agent-demo/internal/config"
	"iot-agent-demo/internal/tools"
)

var globalAgent *react.Agent

// InitAgent 初始化全局 Agent 实例
func InitAgent(ctx context.Context, cfg *config.Config) error {
	// 1. 初始化大模型
	chatModel, err := initChatModel(ctx, cfg)
	if err != nil {
		return fmt.Errorf("初始化模型失败：%w", err)
	}

	// 2. 创建工具
	deviceTool, err := tools.NewDeviceStatusTool()
	if err != nil {
		return fmt.Errorf("创建设备工具失败：%w", err)
	}

	weatherTool, err := tools.NewWeatherTool()
	if err != nil {
		return fmt.Errorf("创建天气工具失败：%w", err)
	}

	controlTool, err := tools.NewDeviceControlTool()
	if err != nil {
		return fmt.Errorf("创建控制工具失败：%w", err)
	}

	// 3. 加载系统提示词
	systemPrompt, err := loadSystemPrompt()
	if err != nil {
		return fmt.Errorf("加载系统提示失败：%w", err)
	}

	// 4. 构建 Agent
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []compose.Tool{
				{Tool: deviceTool},
				{Tool: weatherTool},
				{Tool: controlTool},
			},
		},
		MessageModifier: func(_ context.Context, input []*schema.Message) []*schema.Message {
			return append([]*schema.Message{
				schema.SystemMessage(systemPrompt),
			}, input...)
		},
		MaxStep: 20, // 生产环境建议显式设置
	})

	if err != nil {
		return fmt.Errorf("构建 Agent 失败：%w", err)
	}

	globalAgent = agent
	return nil
}

// initChatModel 初始化聊天模型
func initChatModel(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.OpenAIBaseURL,
		APIKey:  cfg.OpenAIKey,
		Model:   cfg.ModelName,
	})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

// loadSystemPrompt 加载系统提示词
func loadSystemPrompt() (string, error) {
	promptBytes, err := os.ReadFile("prompts/system.txt")
	if err != nil {
		// 如果文件不存在，使用默认提示词
		return getDefaultSystemPrompt(), nil
	}
	return string(promptBytes), nil
}

// getDefaultSystemPrompt 获取默认系统提示词
func getDefaultSystemPrompt() string {
	return `你是一个专业的 IoT 设备管理助手，专门帮助用户查询和管理物联网设备。

【能力范围】
1. 查询设备实时状态（在线状态、传感器数据、报警信息）
2. 获取天气信息并结合设备状态给出维护建议
3. 执行设备控制操作（重启、校准、维护模式）

【输出规则】
1. 必须全程使用中文回答
2. 数据展示要清晰、结构化
3. 发现异常时要给出专业的排查建议
4. 涉及设备控制时要确认操作风险`
}

// GetAgent 获取全局 Agent 实例
func GetAgent() *react.Agent {
	return globalAgent
}
