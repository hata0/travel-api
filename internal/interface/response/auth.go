package response

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
