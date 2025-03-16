package httpmodels

import "github.com/EwvwGeN/cataloger/internal/domain/models"

type CategoryEditRequest struct {
	CategoryNewData models.CategoryForPatch `json:"category_new_data"`
}

type CategoryEditResponse struct {
	Edited bool `json:"edited"`
}