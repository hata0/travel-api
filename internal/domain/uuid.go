package domain

import "github.com/google/uuid"

func NewUUID() string {
	return uuid.New().String()
}

func IsValidUUID(id string) bool {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	return parsedUUID.Version() == 4
}
