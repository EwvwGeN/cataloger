package httpmodels

import "github.com/EwvwGeN/cataloger/internal/domain/models"

type CategoryGetOneResponse struct {
	Category models.Category `json:"category"`
}

type CategoryGetAllResponse struct {
	Categories []models.Category `json:"categories"`
}