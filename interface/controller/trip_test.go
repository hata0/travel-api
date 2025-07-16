package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-api/domain"
	mock_controller "travel-api/interface/controller/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTripController_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_controller.NewMockTripService(ctrl)
	// ginのvalidatorを有効にするため、NewTripControllerの前にSetdefaultsする
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripController := NewTripController(mockService)
	tripController.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"
	now := time.Now()
	expectedTrip := domain.NewTrip(domain.TripID(tripID), "Test Trip", now, now)

	t.Run("正常系", func(t *testing.T) {
		mockService.EXPECT().Get(gomock.Any(), tripID).Return(expectedTrip, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody map[string]domain.Trip
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedTrip.ID, responseBody["trip"].ID)
		assert.Equal(t, expectedTrip.Name, responseBody["trip"].Name)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockService.EXPECT().Get(gomock.Any(), tripID).Return(domain.Trip{}, domain.ErrTripNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockService.EXPECT().Get(gomock.Any(), tripID).Return(domain.Trip{}, errors.New("some error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("異常系: Invalid UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/invalid-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTripController_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_controller.NewMockTripService(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripController := NewTripController(mockService)
	tripController.Register(r)

	now := time.Now()
	expectedTrips := []domain.Trip{
		domain.NewTrip("1", "Trip 1", now, now),
		domain.NewTrip("2", "Trip 2", now, now),
	}

	t.Run("正常系", func(t *testing.T) {
		mockService.EXPECT().List(gomock.Any()).Return(expectedTrips, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var responseBody map[string][]domain.Trip
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Len(t, responseBody["trips"], 2)
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockService.EXPECT().List(gomock.Any()).Return(nil, errors.New("some error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTripController_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_controller.NewMockTripService(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripController := NewTripController(mockService)
	tripController.Register(r)

	tripName := "New Trip"

	t.Run("正常系", func(t *testing.T) {
		mockService.EXPECT().Create(gomock.Any(), tripName).Return(nil)

		body, _ := json.Marshal(gin.H{"name": tripName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Service error", func(t *testing.T) {
		mockService.EXPECT().Create(gomock.Any(), tripName).Return(errors.New("some error"))

		body, _ := json.Marshal(gin.H{"name": tripName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("異常系: Invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer([]byte(`{"name":`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTripController_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_controller.NewMockTripService(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripController := NewTripController(mockService)
	tripController.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"
	updatedName := "Updated Trip"

	t.Run("正常系", func(t *testing.T) {
		mockService.EXPECT().Update(gomock.Any(), tripID, updatedName).Return(nil)

		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockService.EXPECT().Update(gomock.Any(), tripID, updatedName).Return(domain.ErrTripNotFound)

		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("異常系: Invalid UUID", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTripController_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_controller.NewMockTripService(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripController := NewTripController(mockService)
	tripController.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"

	t.Run("正常系", func(t *testing.T) {
		mockService.EXPECT().Delete(gomock.Any(), tripID).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockService.EXPECT().Delete(gomock.Any(), tripID).Return(domain.ErrTripNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("異常系: Invalid UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/invalid-uuid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
