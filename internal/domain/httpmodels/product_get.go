package httpmodels

import "github.com/EwvwGeN/cataloger/internal/domain/models"

type ProductGetOneResponse struct {
	Product models.Product `json:"product"`
}

type ProductGetAllResponse struct {
	Products []models.Product `json:"products"`
}