package models

type User struct {
	Email    string `json:"email"`
	UUID     string `json:"uuid"`
	PassHash []byte `json:"-"`
}