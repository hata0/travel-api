package uuid

import "github.com/google/uuid"

// UUIDGenerator はUUIDを生成するためのインターフェースです。
//
//go:generate mockgen -destination mock/uuid_generator.go travel-api/internal/domain/shared/uuid UUIDGenerator
type UUIDGenerator interface {
	NewUUID() string
}

// DefaultUUIDGenerator はUUIDGeneratorのデフォルト実装です。
type DefaultUUIDGenerator struct{}

// NewUUID は新しいUUIDを生成します。
func (g *DefaultUUIDGenerator) NewUUID() string {
	return uuid.New().String()
}

func IsValidUUID(id string) bool {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	return parsedUUID.Version() == 4
}
