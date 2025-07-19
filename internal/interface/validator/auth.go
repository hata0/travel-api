package validator

type RegisterJSONBody struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginJSONBody struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenJSONBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
