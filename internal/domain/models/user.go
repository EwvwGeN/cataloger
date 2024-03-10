package models

type User struct {
	Email       string `json:"email"`
	PassHash    []byte `json:"pass_hash"`
	RefreshHash string `json:"refresh_token"`
	ExpiresAt   int64  `json:"expires_at"`
}