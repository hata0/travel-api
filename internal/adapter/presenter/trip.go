package presenter

import (
	"encoding/json"
	"time"

	"github.com/hata0/travel-api/internal/usecase/output"
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

func NewGetTripResponse(out output.GetTripOutput) GetTripResponse {
	return GetTripResponse{
		Trip: Trip{
			ID:        out.Trip.ID,
			Name:      out.Trip.Name,
			CreatedAt: out.Trip.CreatedAt,
			UpdatedAt: out.Trip.UpdatedAt,
		},
	}
}

func NewListTripResponse(out output.ListTripOutput) ListTripResponse {
	formattedTrips := make([]Trip, len(out.Trips))
	for i, trip := range out.Trips {
		formattedTrips[i] = Trip{
			ID:        trip.ID,
			Name:      trip.Name,
			CreatedAt: trip.CreatedAt,
			UpdatedAt: trip.UpdatedAt,
		}
	}
	return ListTripResponse{
		Trips: formattedTrips,
	}
}

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
