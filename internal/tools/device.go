package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// DeviceQueryRequest 设备查询请求
type DeviceQueryRequest struct {
	DeviceID string `json:"device_id" jsonschema:"description=设备 ID，例如 CPT-001"`
	TenantID string `json:"tenant_id" jsonschema:"description=租户 ID"`
}

// NewDeviceStatusTool 创建设备状态查询工具
func NewDeviceStatusTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"query_device_status",
		"查询 IoT 设备的实时运行状态，包括设备在线状态、传感器数据、报警信息等",
		func(ctx context.Context, req *DeviceQueryRequest) (string, error) {
			if req.DeviceID == "" {
				return "", fmt.Errorf("device_id is required")
			}

			// 模拟业务服务调用（实际项目中替换为真实数据库/API 调用）
			deviceData := mockDeviceData(req.DeviceID)

			return fmt.Sprintf(
				"设备 %s 状态报告:\n"+
					"- 在线状态：%s\n"+
					"- 租户：%s\n"+
					"- 锥尖阻力 (qc): %.2f MPa\n"+
					"- 摩阻力 (fs): %.2f kPa\n"+
					"- 孔隙水压力 (u2): %.2f kPa\n"+
					"- 温度：%.1f °C\n"+
					"- 最后更新：%s\n"+
					"- 报警状态：%s",
				req.DeviceID,
				deviceData.Status,
				req.TenantID,
				deviceData.Qc,
				deviceData.Fs,
				deviceData.U2,
				deviceData.Temperature,
				deviceData.LastUpdate.Format("2006-01-02 15:04:05"),
				deviceData.AlarmStatus,
			), nil
		},
	)
}

// DeviceData 设备数据结构
type DeviceData struct {
	Status      string
	Qc          float64
	Fs          float64
	U2          float64
	Temperature float64
	LastUpdate  time.Time
	AlarmStatus string
}

// mockDeviceData 模拟设备数据（实际项目中替换为真实数据源）
func mockDeviceData(deviceID string) *DeviceData {
	rand.Seed(time.Now().UnixNano())

	// 模拟正常数据
	qc := 10.0 + rand.Float64()*20.0  // 10-30 MPa
	fs := 50.0 + rand.Float64()*100.0  // 50-150 kPa
	u2 := 100.0 + rand.Float64()*200.0 // 100-300 kPa
	temp := 15.0 + rand.Float64()*20.0 // 15-35 °C

	status := "在线"
	alarm := "正常"

	// 10% 概率模拟异常
	if rand.Float64() < 0.1 {
		status = "离线"
		alarm = "通信中断"
	} else if rand.Float64() < 0.15 {
		alarm = "数据异常"
	}

	return &DeviceData{
		Status:      status,
		Qc:          qc,
		Fs:          fs,
		U2:          u2,
		Temperature: temp,
		LastUpdate:  time.Now(),
		AlarmStatus: alarm,
	}
}

// DeviceControlRequest 设备控制请求
type DeviceControlRequest struct {
	DeviceID string `json:"device_id" jsonschema:"description=设备 ID"`
	Action   string `json:"action" jsonschema:"description=操作类型：restart/calibrate/maintenance"`
	TenantID string `json:"tenant_id" jsonschema:"description=租户 ID"`
}

// NewDeviceControlTool 创建设备控制工具（示例：重启设备）
func NewDeviceControlTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"control_device",
		"控制 IoT 设备执行操作，如重启、校准、进入维护模式等",
		func(ctx context.Context, req *DeviceControlRequest) (string, error) {
			if req.DeviceID == "" || req.Action == "" {
				return "", fmt.Errorf("device_id and action are required")
			}

			// 模拟控制操作
			return fmt.Sprintf(
				"设备控制指令已执行:\n"+
					"- 设备：%s\n"+
					"- 操作：%s\n"+
					"- 状态：成功\n"+
					"- 执行时间：%s",
				req.DeviceID,
				req.Action,
				time.Now().Format("2006-01-02 15:04:05"),
			), nil
		},
	)
}
