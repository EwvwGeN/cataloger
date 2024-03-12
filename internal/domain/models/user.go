package models

type User struct {
	Email       string `json:"email"`
	PassHash    string `json:"pass_hash"`
	RefreshHash string `json:"refresh_hash"`
	ExpiresAt   int64  `json:"expires_at"`
}