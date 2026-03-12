package tools

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// WeatherQuery 天气查询请求
type WeatherQuery struct {
	City string `json:"city" jsonschema:"description=需要查询天气的城市名称"`
}

// NewWeatherTool 创建天气查询工具
func NewWeatherTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_weather",
		"获取指定城市的当前天气信息，包括温度、天气状况、风力等",
		func(ctx context.Context, req *WeatherQuery) (string, error) {
			if req.City == "" {
				return "", fmt.Errorf("city is required")
			}

			// 模拟天气数据（实际项目中调用真实天气 API）
			weather := mockWeather(req.City)

			return fmt.Sprintf(
				"%s 天气信息:\n"+
					"- 天气状况：%s\n"+
					"- 温度：%d°C\n"+
					"- 湿度：%d%%\n"+
					"- 风力：%s\n"+
					"- 空气质量：%s\n"+
					"- 更新时间：%s",
				req.City,
				weather.Condition,
				weather.Temperature,
				weather.Humidity,
				weather.Wind,
				weather.AQI,
				time.Now().Format("2006-01-02 15:04:05"),
			), nil
		},
	)
}

// WeatherData 天气数据结构
type WeatherData struct {
	City        string
	Condition   string
	Temperature int
	Humidity    int
	Wind        string
	AQI         string
}

// mockWeather 模拟天气数据
func mockWeather(city string) *WeatherData {
	rand.Seed(time.Now().UnixNano())

	conditions := []string{"晴", "多云", "阴", "小雨", "大雨"}
	winds := []string{"微风", "1-2 级", "2-3 级", "3-4 级", "4-5 级"}
	aqis := []string{"优", "良", "轻度污染", "中度污染"}

	temp := 15 + rand.Intn(20) // 15-35°C
	humidity := 40 + rand.Intn(50)

	return &WeatherData{
		City:        city,
		Condition:   conditions[rand.Intn(len(conditions))],
		Temperature: temp,
		Humidity:    humidity,
		Wind:        winds[rand.Intn(len(winds))],
		AQI:         aqis[rand.Intn(len(aqis))],
	}
}
