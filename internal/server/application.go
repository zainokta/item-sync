package server

import (
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/infrastructure/database"
	loggerPkg "github.com/zainokta/item-sync/pkg/logger"
)

type Application struct {
	config   *config.Config
	logger   loggerPkg.Logger
	database *sql.DB
	redis    *redis.Client
	server   Server
}

func NewApplication() (*Application, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	logger := loggerPkg.NewLogger(loggerPkg.LogLevel(cfg.LogLevel), cfg.Environment)
	db, err := database.NewMysqlDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Warn("Redis connection failed", "error", err)
		return nil, err
	}

	server, err := NewEchoServer(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &Application{
		config:   cfg,
		logger:   logger,
		database: db,
		redis:    redisClient,
		server:   server,
	}, nil
}

func (a *Application) GetServer() Server {
	return a.server
}

func (a *Application) GetLogger() loggerPkg.Logger {
	return a.logger
}

func (a *Application) Start() error {
	return a.server.Start()
}

func (a *Application) Stop() error {
	return a.server.Stop()
}

func (a *Application) Shutdown() {
	// Close database connection
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			a.logger.Error("Failed to close database connection", "error", err)
		}
	}

	// Close redis connection
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.logger.Error("Failed to close redis connection", "error", err)
		}
	}

	// Stop server
	if err := a.server.Stop(); err != nil {
		a.logger.Error("Failed to stop server", "error", err)
	}
}
