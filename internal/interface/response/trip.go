package response

import (
	"encoding/json"
	"time"
)

type (
	Trip struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	GetTripResponse struct {
		Trip Trip `json:"trip"`
	}

	ListTripResponse struct {
		Trips []Trip `json:"trips"`
	}

	CreateTripResponse struct {
		ID string `json:"id"`
	}
)

// MarshalJSON はTrip構造体をJSONにマーシャリングする際のカスタム処理を提供します。
// CreatedAtとUpdatedAtフィールドをRFC3339形式でフォーマットします。
func (t Trip) MarshalJSON() ([]byte, error) {
	type Alias Trip // 無限ループを防ぐためのエイリアス
	return json.Marshal(&struct {
		Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (Alias)(t),
		CreatedAt: t.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339Nano),
	})
}
