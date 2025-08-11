package mapper

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// PostgreSQLTypeMapper はPostgreSQLの型変換を担当するマッパー
type PostgreSQLTypeMapper struct{}

// NewPostgreSQLTypeMapper は新しいPostgreSQLTypeMapperを作成する
func NewPostgreSQLTypeMapper() *PostgreSQLTypeMapper {
	return &PostgreSQLTypeMapper{}
}

// ToUUID は文字列をpgtype.UUIDに変換する
func (m *PostgreSQLTypeMapper) ToUUID(uuidStr string) (pgtype.UUID, error) {
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(uuidStr); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID format: %w", err)
	}
	return pgUUID, nil
}

// ToTimestamp は時刻をpgtype.Timestamptzに変換する
func (m *PostgreSQLTypeMapper) ToTimestamp(t time.Time) (pgtype.Timestamptz, error) {
	var pgTime pgtype.Timestamptz
	if err := pgTime.Scan(t); err != nil {
		return pgtype.Timestamptz{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	return pgTime, nil
}

// FromUUID はpgtype.UUIDを文字列に変換する
func (m *PostgreSQLTypeMapper) FromUUID(pgUUID pgtype.UUID) (string, error) {
	if !pgUUID.Valid {
		return "", fmt.Errorf("UUID is null")
	}
	return pgUUID.String(), nil
}

// FromTimestamp はpgtype.Timestamptzをtime.Timeに変換する
func (m *PostgreSQLTypeMapper) FromTimestamp(pgTime pgtype.Timestamptz) (time.Time, error) {
	if !pgTime.Valid {
		return time.Time{}, fmt.Errorf("timestamp is null")
	}
	return pgTime.Time, nil
}
