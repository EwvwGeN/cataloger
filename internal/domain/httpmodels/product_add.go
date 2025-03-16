package httpmodels

import "github.com/EwvwGeN/cataloger/internal/domain/models"

type ProductAddRequest struct {
	Product models.Product `json:"product"`
}

type ProductAddResponse struct {
	ProductId string `json:"product_id"`
}