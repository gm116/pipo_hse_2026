package config

import (
	"os"
	"strconv"
	"time"
)

type AuthService struct {
	AppName     string
	Port        string
	LogLevel    string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
}

type TaskService struct {
	AppName     string
	Port        string
	LogLevel    string
	DatabaseURL string
	JWTSecret   string
}

type Gateway struct {
	AppName        string
	Port           string
	LogLevel       string
	AuthServiceURL string
	TaskServiceURL string
}

func LoadAuthService() AuthService {
	return AuthService{
		AppName:     getenv("APP_NAME", "auth-service"),
		Port:        getenv("PORT", "8081"),
		LogLevel:    getenv("LOG_LEVEL", "INFO"),
		DatabaseURL: getenv("DATABASE_URL", ""),
		JWTSecret:   getenv("JWT_SECRET", ""),
		TokenTTL:    time.Duration(getenvInt("TOKEN_TTL_HOURS", 24)) * time.Hour,
	}
}

func LoadTaskService() TaskService {
	return TaskService{
		AppName:     getenv("APP_NAME", "task-service"),
		Port:        getenv("PORT", "8082"),
		LogLevel:    getenv("LOG_LEVEL", "INFO"),
		DatabaseURL: getenv("DATABASE_URL", ""),
		JWTSecret:   getenv("JWT_SECRET", ""),
	}
}

func LoadGateway() Gateway {
	return Gateway{
		AppName:        getenv("APP_NAME", "gateway"),
		Port:           getenv("PORT", "8080"),
		LogLevel:       getenv("LOG_LEVEL", "INFO"),
		AuthServiceURL: getenv("AUTH_SERVICE_URL", "http://localhost:8081"),
		TaskServiceURL: getenv("TASK_SERVICE_URL", "http://localhost:8082"),
	}
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}
