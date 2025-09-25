package server

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/zainokta/item-sync/config"
	_ "github.com/zainokta/item-sync/docs"
	"github.com/zainokta/item-sync/internal/item/handler"
	"github.com/zainokta/item-sync/internal/item/repository"
	"github.com/zainokta/item-sync/internal/item/usecase"
	loggerPkg "github.com/zainokta/item-sync/pkg/logger"
)

func RegisterRoutes(e *echo.Echo, cfg *config.Config, logger loggerPkg.Logger, repoContainer *repository.RepositoryContainer) {
	// Create use cases with configured API client
	syncUseCase := usecase.NewSyncItemsUseCase(cfg, repoContainer.GetItemRepository(), repoContainer.GetJobRepository(), logger)
	listUseCase := usecase.NewListItemsUseCase(repoContainer.GetItemRepository(), repoContainer.GetItemCache(), logger)
	detailUseCase := usecase.NewFetchItemUseCase(cfg, repoContainer.GetItemRepository(), repoContainer.GetItemCache(), logger)

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
