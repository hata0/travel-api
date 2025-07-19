package response

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

type AuthTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
