package server

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/infrastructure/database"
	"github.com/zainokta/item-sync/internal/infrastructure/worker"
	"github.com/zainokta/item-sync/internal/item/jobs"
	"github.com/zainokta/item-sync/internal/item/repository"
	"github.com/zainokta/item-sync/pkg/api"
	loggerPkg "github.com/zainokta/item-sync/pkg/logger"
	"github.com/zainokta/item-sync/pkg/migration"
)

type Application struct {
	config    *config.Config
	logger    loggerPkg.Logger
	database  *sql.DB
	redis     *redis.Client
	server    Server
	scheduler *worker.Scheduler
	ctx       context.Context
	cancel    context.CancelFunc
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

	// Run database migrations automatically if enabled
	if cfg.Migration.Enabled {
		logger.Info("Running database migrations...", "path", cfg.Migration.MigrationsPath)

		databaseURL := buildDatabaseURL(cfg.Database)
		migrator, err := migration.NewMigrator(migration.Config{
			DatabaseURL:    databaseURL,
			MigrationsPath: cfg.Migration.MigrationsPath,
			Logger:         logger,
		})
		if err != nil {
			if cfg.Migration.FailOnError {
				return nil, fmt.Errorf("failed to create migrator: %w", err)
			}
			logger.Warn("Failed to create migrator, continuing...", "error", err)
		} else {
			defer migrator.Close()

			if err := migrator.Up(); err != nil {
				if cfg.Migration.FailOnError {
					return nil, fmt.Errorf("migration failed: %w", err)
				}
				logger.Warn("Migration failed, continuing...", "error", err)
			} else {
				logger.Info("Database migrations completed successfully")
			}
		}
	} else {
		logger.Info("Database migrations disabled")
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

	// Create repository container
	repoContainer := repository.NewRepositoryContainer(db, redisClient, cfg.Cache.DefaultTTL, logger)

	RegisterRoutes(server.GetEcho(), cfg, logger, repoContainer)

	// Create worker scheduler
	ctx, cancel := context.WithCancel(context.Background())
	scheduler := worker.NewScheduler(cfg.Worker, logger)

	// Create and register sync jobs if worker is enabled
	if cfg.Worker.Enabled {
		availableAPIs := []string{"pokemon", "openweather"}

		// Create API client
		for _, availableAPI := range availableAPIs {
			apiClient, err := api.NewAPIClient(availableAPI, cfg.API, cfg.Retry, logger)
			if err != nil {
				logger.Warn("Failed to create API client for worker", "error", err)
			} else {
				// Register sync job
				syncJob := jobs.NewSyncJob(
					"background-sync",
					repoContainer.GetItemRepository(),
					repoContainer.GetJobRepository(),
					apiClient,
					availableAPI,
					logger,
					*cfg,
					nil,
				)
				scheduler.RegisterJob(syncJob)
			}
		}

	}

	return &Application{
		config:    cfg,
		logger:    logger,
		database:  db,
		redis:     redisClient,
		server:    server,
		scheduler: scheduler,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (a *Application) GetServer() Server {
	return a.server
}

func (a *Application) GetLogger() loggerPkg.Logger {
	return a.logger
}

func (a *Application) Start() error {
	// Start worker scheduler in background if enabled
	if a.config.Worker.Enabled && a.scheduler != nil {
		go func() {
			a.logger.Info("Starting background worker scheduler")
			if err := a.scheduler.Start(a.ctx); err != nil {
				a.logger.Error("Worker scheduler failed", "error", err)
			}
		}()
	}

	// Start HTTP server
	return a.server.Start()
}

func (a *Application) Stop() error {
	return a.server.Stop()
}

func (a *Application) Shutdown() {
	// Stop worker scheduler first
	if a.scheduler != nil {
		a.logger.Info("Stopping worker scheduler")
		a.cancel() // Cancel context
		a.scheduler.Stop()
	}

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

func buildDatabaseURL(dbConfig config.DatabaseConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
	)
}
