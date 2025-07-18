package domain

import "time"

// Clock は現在時刻を取得するためのインターフェースです。
//go:generate mockgen -destination mock/clock.go travel-api/internal/domain Clock
type Clock interface {
	Now() time.Time
}

// SystemClock はClockのデフォルト実装です。
type SystemClock struct{}

// Now は現在時刻を返します。
func (c *SystemClock) Now() time.Time {
	return time.Now()
}
