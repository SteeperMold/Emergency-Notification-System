package bootstrap

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Application holds the core components of the app.
type Application struct {
	Config       *Config
	DB           *pgxpool.Pool
	Logger       *zap.Logger
	KafkaFactory *KafkaFactory
}

// NewApp initializes and returns a new Application instance.
func NewApp() *Application {
	app := &Application{}

	app.Config = NewConfig()
	app.DB = NewSQLDatabase(app.Config)
	app.Logger = NewLogger(app.Config.App.AppEnv)
	app.KafkaFactory = NewKafkaFactory(app.Config.Kafka)

	return app
}

// LoggerSync flushes any buffered log entries.
func (app *Application) LoggerSync() {
	LoggerSync(app.Logger)
}

// CloseDBConnection safely closes the database connection pool.
func (app *Application) CloseDBConnection() {
	CloseDBConnection(app.DB)
}
