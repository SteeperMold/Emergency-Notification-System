package bootstrap

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
)

// Application holds the core components of the app.
type Application struct {
	Config       *Config
	Logger       *zap.Logger
	S3Client     *s3.S3
	KafkaFactory *KafkaFactory
}

// NewApp initializes and returns a new Application instance.
func NewApp() *Application {
	app := &Application{}

	app.Config = NewConfig()
	app.Logger = NewLogger(app.Config.App.AppEnv)
	app.S3Client = NewS3Client(app.Config.S3)
	app.KafkaFactory = NewKafkaFactory(app.Config.Kafka)

	return app
}

// LoggerSync flushes any buffered log entries.
func (app *Application) LoggerSync() {
	LoggerSync(app.Logger)
}
