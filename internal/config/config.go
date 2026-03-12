package config

import (
	"os"
)

type Config struct {
	OpenAIBaseURL string
	OpenAIKey     string
	ModelName     string
	ServerPort    string
}

func Load() *Config {
	cfg := &Config{
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIKey:     getEnv("OPENAI_API_KEY", ""),
		ModelName:     getEnv("MODEL_NAME", "gpt-4o"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
