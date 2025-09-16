package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zainokta/item-sync/internal/item/usecase"
	"github.com/zainokta/item-sync/pkg/logger"
)

type ItemDetailHandler struct {
	fetchItemUseCase *usecase.FetchItemUseCase
	logger           logger.Logger
}

func NewItemDetailHandler(fetchItemUseCase *usecase.FetchItemUseCase, logger logger.Logger) *ItemDetailHandler {
	return &ItemDetailHandler{
		fetchItemUseCase: fetchItemUseCase,
		logger:           logger,
	}
}

func (h *ItemDetailHandler) GetItemDetail(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid ID format",
		})
	}

	apiSource := c.QueryParam("api_source")
	if apiSource == "" {
		apiSource = "unknown" 
	}

	req := usecase.FetchItemRequest{
		ID:        id,
		APISource: apiSource,
	}

	response, err := h.fetchItemUseCase.Execute(c.Request().Context(), req)
	if err != nil {
		h.logger.Error("Failed to get item detail", "error", err, "id", id, "api_source", apiSource)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"item":       response.Item,
		"from_cache": response.FromCache,
	})
}
