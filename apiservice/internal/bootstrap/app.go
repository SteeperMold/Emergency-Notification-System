package bootstrap

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Application holds the core components of the app.
type Application struct {
	Config *Config
	Logger *zap.Logger
	DB     *pgxpool.Pool
}

// NewApp initializes and returns a new Application instance.
func NewApp() *Application {
	app := &Application{}
	app.Config = NewConfig()
	app.Logger = NewLogger(app.Config.App.AppEnv)
	app.DB = NewSQLDatabase(app.Config)
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
