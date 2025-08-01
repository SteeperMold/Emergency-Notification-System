package bootstrap

import (
	"log"

	"go.uber.org/zap"
)

// NewLogger creates a new zap.Logger instance based on the given application environment.
// Supported environments are "production" and "development".
// It terminates the program if the logger creation fails.
func NewLogger(appEnv string) *zap.Logger {
	var logger *zap.Logger
	var err error

	switch appEnv {
	case "production":
		logger, err = zap.NewProduction()
	case "development":
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	return logger
}

// LoggerSync flushes any buffered log entries.
func LoggerSync(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		log.Printf("logger sync error: %v\n", err)
	}
}
