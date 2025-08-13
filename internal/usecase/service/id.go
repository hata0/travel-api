package service

//go:generate mockgen -destination mock/id.go github.com/hata0/travel-api/internal/usecase/services IDGenerator
type IDGenerator interface {
	Generate() string
}
