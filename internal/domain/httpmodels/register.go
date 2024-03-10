package httpmodels

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterReqsponse struct {
	Registered bool   `json:"registered"`
	NewUserId  string `json:"new_user_id"`
}