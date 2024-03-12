package httpmodels

import "github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"

type CategoryGetOneResponse struct {
	Category models.Category `json:"category"`
}

type CategoryGetAllResponse struct {
	Categories []models.Category `json:"categories"`
}