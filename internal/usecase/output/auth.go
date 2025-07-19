package output

type RegisterOutput struct {
	UserID string
}

type LoginOutput struct {
	Token        string
	RefreshToken string
}
