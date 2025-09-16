package dto

import (
	"github.com/go-playground/validator/v10"
)

type SyncItemsRequest struct {
	ForceSync bool `json:"force_sync" validate:"boolean"`
}

func (r SyncItemsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type GetItemsRequest struct {
	Status string `json:"status" query:"status" validate:"omitempty,oneof=pending completed failed"`
	Limit  int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset" query:"offset" validate:"omitempty,min=0"`
}

func (r GetItemsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
