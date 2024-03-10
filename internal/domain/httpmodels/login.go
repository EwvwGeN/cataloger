package httpmodels

import "github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	TokenPair models.TokenPair `json:"token_pair"`
}