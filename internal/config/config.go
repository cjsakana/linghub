package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	Port          int
	DBDSN         string
	JWTSecret     string
	JWTExpiration time.Duration
	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
	SystemPrompt  string
}

// Load reads configuration from environment variables.
func Load() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	jwtExpiration := 7 * 24 * time.Hour
	if exp := os.Getenv("JWT_EXPIRATION_HOURS"); exp != "" {
		if hours, err := strconv.ParseInt(exp, 10, 64); err == nil {
			jwtExpiration = time.Duration(hours) * time.Hour
		}
	}

	return &Config{
		Port:          port,
		DBDSN:         getEnvOrPanic("DB_DSN"),
		JWTSecret:     getEnvOrPanic("JWT_SECRET"),
		JWTExpiration: jwtExpiration,
		OpenAIAPIKey:  getEnvOrPanic("OPENAI_API_KEY"),
		OpenAIBaseURL: getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:   getEnvOrPanic("OPENAI_MODEL"),
		SystemPrompt:  getEnvOrDefault("SYSTEM_PROMPT", "你是一个有帮助的中文AI助手。"),
	}
}

func getEnvOrPanic(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	panic("missing required environment variable: " + key)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
