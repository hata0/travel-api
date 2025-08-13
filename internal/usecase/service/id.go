package service

//go:generate mockgen -destination mock/id.go github.com/hata0/travel-api/internal/usecase/service IDService
type IDService interface {
	Generate() string
}
