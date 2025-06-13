package bootstrap

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application and database configuration.
type Config struct {
	App AppConfig
	DB  DBConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	AppEnv         string
	Port           string
	ContextTimeout time.Duration
	FrontendOrigin string
	Jwt            JWTConfig
}

// JWTConfig holds JWT secret keys and expiry durations for access and refresh tokens.
type JWTConfig struct {
	AccessSecret  string
	AccessExpiry  time.Duration
	RefreshSecret string
	RefreshExpiry time.Duration
}

// DBConfig holds PostgreSQL database connection settings.
type DBConfig struct {
	Host              string
	Port              string
	Name              string
	User              string
	Password          string
	ConnectionTimeout time.Duration
}

// NewConfig loads configuration from environment variables with defaults.
func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file, using defaults: %v\n", err)
	}

	return &Config{
		App: AppConfig{
			AppEnv:         getEnv("APP_ENV", "development"),
			Port:           getEnv("PORT", "8080"),
			ContextTimeout: getEnvAsDuration("CONTEXT_TIMEOUT_MS", 2000) * time.Millisecond,
			FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
			Jwt: JWTConfig{
				AccessSecret:  getEnv("JWT_ACCESS_SECRET", "very_secret1"),
				AccessExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY_H", 2) * time.Hour,
				RefreshSecret: getEnv("JWT_REFRESH_SECRET", "very_secret2"),
				RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY_H", 720) * time.Hour,
			},
		},
		DB: DBConfig{
			Host:              getEnv("DB_HOST", "postgres"),
			Port:              getEnv("DB_PORT", "5432"),
			Name:              getEnv("DB_NAME", "devdb"),
			User:              getEnv("DB_USER", "user"),
			Password:          getEnv("DB_PASSWORD", "123456789admin"),
			ConnectionTimeout: getEnvAsDuration("DB_CONNECTION_TIMEOUT_MS", 10000) * time.Millisecond,
		},
	}
}

func getEnv(key string, defaultVal string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}

	return value
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return value
}

func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	value := getEnvAsInt(name, int(defaultVal))
	return time.Duration(value)
}
