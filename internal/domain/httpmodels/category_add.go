package httpmodels

import "github.com/EwvwGeN/cataloger/internal/domain/models"

type CategoryAddRequest struct {
	Category models.Category `json:"category"`
}

type CategoryAddResponse struct {
	Added bool `json:"added"`
}