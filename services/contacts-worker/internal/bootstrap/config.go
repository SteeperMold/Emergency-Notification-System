package bootstrap

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all top-level configuration sections.
type Config struct {
	App   *AppConfig
	DB    *DBConfig
	S3    *S3Config
	Kafka *KafkaConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	AppEnv         string
	Port           string
	ContextTimeout time.Duration
	BatchSize      int
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

// S3Config holds credentials and endpoint for S3 storage.
type S3Config struct {
	ID       string
	Key      string
	Region   string
	Endpoint string
	Bucket   string
}

// KafkaConfig holds broker addresses and topic configuration.
type KafkaConfig struct {
	KafkaAddrs    []string
	Topics        map[string]string
	ConsumerGroup string
}

// NewConfig reads environment variables (optionally from a .env file) and
// returns a Config populated with defaults where variables are unset.
func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load .env file, using default: %v\n", err)
	}

	return &Config{
		App: &AppConfig{
			AppEnv:         getEnv("APP_ENV", "development"),
			Port:           getEnv("PORT", "8080"),
			ContextTimeout: getEnvAsDuration("CONTEXT_TIMEOUT_MS", 2000) * time.Millisecond,
			BatchSize:      getEnvAsInt("BATCH_SIZE", 100000),
		},
		DB: &DBConfig{
			Host:              getEnv("DB_HOST", "apiservice"),
			Port:              getEnv("DB_PORT", "5432"),
			Name:              getEnv("DB_NAME", "api_service_postgres"),
			User:              getEnv("DB_USER", "user"),
			Password:          getEnv("DB_PASSWORD", "123456789admin"),
			ConnectionTimeout: getEnvAsDuration("DB_CONNECTION_TIMEOUT_MS", 10000) * time.Millisecond,
		},
		S3: &S3Config{
			ID:       getEnv("S3_SECRET_ID", "miniouser"),
			Key:      getEnv("S3_SECRET_KEY", "miniopassword"),
			Region:   getEnv("S3_REGION", "us-east-1"),
			Endpoint: getEnv("S3_ENDPOINT", "http://minio:9000"),
			Bucket:   getEnv("S3_BUCKET", "contacts-bucket"),
		},
		Kafka: &KafkaConfig{
			KafkaAddrs: getEnvAsSlice("KAFKA_ADDRS", []string{"kafka:9092"}, ","),
			Topics: map[string]string{
				"contacts.loading.tasks":   getEnv("KAFKA_TOPIC_CONTACTS_LOADING_TASKS", "contacts.loading.tasks"),
				"contacts.loading.results": getEnv("KAFKA_TOPIC_CONTACTS_LOADING_RESULTS", "contacts.loading.results"),
			},
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "contacts-worker-group"),
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
