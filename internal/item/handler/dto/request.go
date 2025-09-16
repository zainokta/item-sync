package dto

import (
	"github.com/go-playground/validator/v10"
)

// SyncItemsRequest represents the request body for syncing items
type SyncItemsRequest struct {
	ForceSync bool                   `json:"force_sync" validate:"boolean" example:"false" description:"Force sync even if data already exists"`
	APISource string                 `json:"api_source" example:"pokemon" description:"API source to sync from (pokemon, openweather)"`
	Operation string                 `json:"operation" example:"list" description:"Operation to perform (list, get)"`
	Params    map[string]interface{} `json:"params" example:"{\"limit\":20,\"offset\":0}" description:"Additional parameters for the API call"`
}

func (r SyncItemsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// GetItemsRequest represents the query parameters for listing items
type GetItemsRequest struct {
	Status string `json:"status" query:"status" validate:"omitempty,oneof=pending completed failed" example:"completed" description:"Filter by status"`
	Limit  int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100" example:"20" description:"Number of items to return (max 100)"`
	Offset int    `json:"offset" query:"offset" validate:"omitempty,min=0" example:"0" description:"Number of items to skip for pagination"`
}

func (r GetItemsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
