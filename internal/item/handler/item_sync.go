package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/handler/dto"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

type SyncHandler struct {
	syncUseCase *usecase.SyncItemsUseCase
	logger      logger.Logger
}

func NewSyncHandler(syncUseCase *usecase.SyncItemsUseCase, logger logger.Logger) *SyncHandler {
	return &SyncHandler{
		syncUseCase: syncUseCase,
		logger:      logger,
	}
}

func (h *SyncHandler) SyncItems(c echo.Context) error {
	var req dto.SyncItemsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: err.Error(),
		})
	}

	ctx := c.Request().Context()
	response, err := h.syncUseCase.Execute(ctx, usecase.SyncItemsRequest{
		ForceSync: req.ForceSync,
	})

	if err != nil {
		h.logger.Error("Sync failed", "error", err.Error())

		var domainErr *pkgErrors.DomainError
		if errors.As(err, &domainErr) {
			return c.JSON(getHTTPStatusFromError(domainErr), dto.ErrorResponse{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Details: domainErr.Details,
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		})
	}

	h.logger.Info("Sync completed",
		"success_count", response.SuccessCount,
		"failed_count", response.FailedCount,
	)

	return c.JSON(http.StatusOK, dto.SyncItemsResponse{
		SuccessCount: response.SuccessCount,
		FailedCount:  response.FailedCount,
		Items:        response.Items,
		Errors:       response.Errors,
	})
}

func getHTTPStatusFromError(err *pkgErrors.DomainError) int {
	switch err.Category {
	case pkgErrors.CategoryValidation:
		return http.StatusBadRequest
	case pkgErrors.CategoryNotFound:
		return http.StatusNotFound
	case pkgErrors.CategoryExternalAPI:
		return http.StatusBadGateway
	case pkgErrors.CategoryCache:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
