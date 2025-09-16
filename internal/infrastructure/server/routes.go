package server

import (
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/zainokta/item-sync/config"
	_ "github.com/zainokta/item-sync/docs"
	"github.com/zainokta/item-sync/internal/item/handler"
	"github.com/zainokta/item-sync/internal/item/repository"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/api"
	loggerPkg "github.com/zainokta/item-sync/pkg/logger"
)

func RegisterRoutes(e *echo.Echo, cfg *config.Config, logger loggerPkg.Logger, db *sql.DB, redis *redis.Client) {
	// Create dependencies
	itemRepo := repository.NewItemRepository(db, logger)
	itemCache := repository.NewItemCache(redis, cfg.Cache.DefaultTTL, logger)

	// Create API client based on configuration
	apiClient, err := api.NewAPIClient(cfg.API.APIType, cfg.API, cfg.Retry, logger)
	if err != nil {
		logger.Error("Failed to create API client", "error", err, "api_type", cfg.API.APIType)
		apiClient = nil
	}

	// Create use cases with configured API client
	syncUseCase := usecase.NewSyncItemsUseCase(itemRepo, apiClient, itemCache, logger, cfg.API.APIType)
	listUseCase := usecase.NewListItemsUseCase(itemRepo, itemCache, logger)
	detailUseCase := usecase.NewFetchItemUseCase(itemRepo, apiClient, itemCache, logger)

	// Create handlers
	syncHandler := handler.NewSyncHandler(syncUseCase, logger)
	listHandler := handler.NewListHandler(listUseCase, logger)
	detailHandler := handler.NewItemDetailHandler(detailUseCase, logger)

	// Health check endpoint
	// @Summary      Health check
	// @Description  Check if the service is healthy and operational
	// @Tags         health
	// @Accept       json
	// @Produce      json
	// @Success      200 {object} map[string]string "Service is healthy"
	// @Router       /health [get]
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	e.POST("/sync", syncHandler.SyncItems)
	e.GET("/items", listHandler.ListItems)
	e.GET("/items/:id", detailHandler.GetItemDetail)

	// Swagger documentation endpoints
	// Only serve Swagger UI in development and staging environments
	if cfg.Environment != "production" {
		// @Summary      Swagger API Documentation
		// @Description  Interactive API documentation and testing interface
		// @Tags         docs
		// @Produce      text/html
		// @Success      200 {string} html "Swagger UI HTML page"
		// @Router       /swagger/ [get]
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}
