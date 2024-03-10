package httpmodels

import "github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"

type RefreshRequest struct {
	TokenPair models.TokenPair `json:"token_pair"`
}

type RefreshResponse struct {
	TokenPair models.TokenPair `json:"token_pair"`
}