package bootstrap

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application and database configuration.
type Config struct {
	App    *AppConfig
	DB     *DBConfig
	Kafka  *KafkaConfig
	Twilio *TwilioConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	AppEnv         string
	MaxAttempts    int
	Port           string
	ContextTimeout time.Duration
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

// KafkaConfig defines Kafka broker addresses and topic names.
type KafkaConfig struct {
	KafkaAddrs    []string
	Topics        map[string]string
	ConsumerGroup string
}

type TwilioConfig struct {
	AccountSID             string
	AuthToken              string
	FromNumber             string
	StatusCallbackEndpoint string
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
			MaxAttempts:    getEnvAsInt("MAX_NOTIFICATION_ATTEMPTS", 5),
			Port:           getEnv("PORT", "8081"),
			ContextTimeout: getEnvAsDuration("CONTEXT_TIMEOUT_MS", 2000) * time.Millisecond,
		},
		DB: &DBConfig{
			Host:              getEnv("DB_HOST", "notification-service"),
			Port:              getEnv("DB_PORT", "5432"),
			Name:              getEnv("DB_NAME", "notification-service-postgres"),
			User:              getEnv("DB_USER", "user"),
			Password:          getEnv("DB_PASSWORD", "123456789admin"),
			ConnectionTimeout: getEnvAsDuration("DB_CONNECTION_TIMEOUT_MS", 10000) * time.Millisecond,
		},
		Kafka: &KafkaConfig{
			KafkaAddrs: getEnvAsSlice("KAFKA_ADDRS", []string{"kafka:9092"}, ","),
			Topics: map[string]string{
				"notification.requests": getEnv("KAFKA_TOPIC_NOTIFICATION_REQUESTS", "notification.requests"),
				"notification.tasks":    getEnv("KAFKA_TOPIC_NOTIFICATION_TASKS", "notification.tasks"),
			},
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "notification-requests-group"),
		},
		Twilio: &TwilioConfig{
			AuthToken:              getEnv("TWILIO_AUTH_TOKEN", "twilio-auth-token"),
			StatusCallbackEndpoint: getEnv("STATUS_CALLBACK_ENDPOINT", "https://some-digits.ngrok.io"),
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
