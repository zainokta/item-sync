package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	pkgErrors "github.com/zainokta/item-sync/internal/errors"
	"github.com/zainokta/item-sync/internal/item/handler/dto"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

type ListHandler struct {
	listUseCase *usecase.ListItemsUseCase
	logger      logger.Logger
}

func NewListHandler(listUseCase *usecase.ListItemsUseCase, logger logger.Logger) *ListHandler {
	return &ListHandler{
		listUseCase: listUseCase,
		logger:      logger,
	}
}

func (h *ListHandler) ListItems(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	itemType := c.QueryParam("item_type")
	status := c.QueryParam("status")
	apiSource := c.QueryParam("api_source")

	ctx := c.Request().Context()
	response, err := h.listUseCase.Execute(ctx, usecase.ListItemsRequest{
		Limit:     limit,
		Offset:    offset,
		ItemType:  itemType,
		Status:    status,
		APISource: apiSource,
	})

	if err != nil {
		h.logger.Error("List items failed", "error", err.Error())

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

	h.logger.Info("List items completed",
		"total", response.TotalCount,
		"limit", limit,
		"offset", offset,
	)

	return c.JSON(http.StatusOK, dto.GetItemsResponse{
		Items: response.Items,
		Total: response.TotalCount,
	})
}
