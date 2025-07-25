package injector

import (
	"travel-api/internal/interface/handler"
)

// Handlers はハンドラーを提供する
type Handlers struct {
	usecases *Usecases

	tripHandler *handler.TripHandler
	authHandler *handler.AuthHandler
}

// NewHandlers はハンドラーを初期化する
func NewHandlers(usecases *Usecases) *Handlers {
	return &Handlers{
		usecases: usecases,
	}
}

func (h *Handlers) TripHandler() *handler.TripHandler {
	if h.tripHandler == nil {
		h.tripHandler = handler.NewTripHandler(h.usecases.TripUsecase())
	}
	return h.tripHandler
}

func (h *Handlers) AuthHandler() *handler.AuthHandler {
	if h.authHandler == nil {
		h.authHandler = handler.NewAuthHandler(h.usecases.AuthUsecase())
	}
	return h.authHandler
}
