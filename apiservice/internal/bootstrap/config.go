package bootstrap

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application and database configuration.
type Config struct {
	App   *AppConfig
	DB    *DBConfig
	S3    *S3Config
	Kafka *KafkaConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	AppEnv                  string
	Port                    string
	ContextTimeout          time.Duration
	FrontendOrigin          string
	Jwt                     *JWTConfig
	ContactsPerKafkaMessage int
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

// S3Config defines parameters for connecting to an S3-compatible service.
type S3Config struct {
	ID       string
	Key      string
	Region   string
	Endpoint string
	Buckets  map[string]string
}

// KafkaConfig defines Kafka broker addresses and topic names.
type KafkaConfig struct {
	KafkaAddrs                       []string
	Topics                           map[string]string
	NotificationRequestsBatchTimeout time.Duration
}

// NewConfig loads configuration from environment variables with defaults.
func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file, using defaults: %v\n", err)
	}

	return &Config{
		App: &AppConfig{
			AppEnv:         getEnv("APP_ENV", "development"),
			Port:           getEnv("PORT", "8080"),
			ContextTimeout: getEnvAsDuration("CONTEXT_TIMEOUT_MS", 2000) * time.Millisecond,
			FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
			Jwt: &JWTConfig{
				AccessSecret:  getEnv("JWT_ACCESS_SECRET", "very_secret1"),
				AccessExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY_H", 2) * time.Hour,
				RefreshSecret: getEnv("JWT_REFRESH_SECRET", "very_secret2"),
				RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY_H", 720) * time.Hour,
			},
			ContactsPerKafkaMessage: getEnvAsInt("CONTACTS_PER_KAFKA_MESSAGE", 10_000),
		},
		DB: &DBConfig{
			Host:              getEnv("DB_HOST", "apiservice"),
			Port:              getEnv("DB_PORT", "5432"),
			Name:              getEnv("DB_NAME", "api_service_postgres"),
			User:              getEnv("DB_USER", "user"),
			Password:          getEnv("DB_PASSWORD", "123456789admin"),
			ConnectionTimeout: getEnvAsDuration("DB_CONNECTION_TIMEOUT_MS", 10_000) * time.Millisecond,
		},
		S3: &S3Config{
			ID:       getEnv("S3_SECRET_ID", "miniouser"),
			Key:      getEnv("S3_SECRET_KEY", "miniopassword"),
			Region:   getEnv("S3_REGION", "us-east-1"),
			Endpoint: getEnv("S3_ENDPOINT", "http://minio:9000"),
			Buckets: map[string]string{
				"contacts": getEnv("S3_BUCKET_CONTACTS", "contacts-bucket"),
			},
		},
		Kafka: &KafkaConfig{
			KafkaAddrs: getEnvAsSlice("KAFKA_ADDRS", []string{"kafka:9092"}, ","),
			Topics: map[string]string{
				"contacts.loading.tasks": getEnv("KAFKA_TOPIC_CONTACTS_LOADING_TASKS", "contacts.loading.tasks"),
				"notification.requests":  getEnv("KAFKA_TOPIC_NOTIFICATION_REQUESTS", "notification.requests"),
			},
			NotificationRequestsBatchTimeout: getEnvAsDuration("KAFKA_NOTIFICATION_REQUESTS_BATCH_TIMEOUT_MS", 1) * time.Millisecond,
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

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultVal
	}

	split := strings.Split(valueStr, sep)
	result := make([]string, 0, len(split))
	for _, v := range split {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultVal
	}

	return result
}
