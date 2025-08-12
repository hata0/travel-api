package services

import "time"

type TimeProvider interface {
	Now() time.Time
}
