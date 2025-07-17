package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-api/domain"
	mock_handler "travel-api/interface/handler/mock"
	"travel-api/interface/response"
	"travel-api/usecase/output"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTripHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"
	now := time.Now()
	expectedTrip := domain.NewTrip(domain.TripID(tripID), "Test Trip", now, now)
	expectedOutput := output.NewGetTripOutput(expectedTrip)

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Get(gomock.Any(), tripID).Return(expectedOutput, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resBody response.GetTripResponse
		err := json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput.Trip.ID, resBody.Trip.ID)
		assert.Equal(t, expectedOutput.Trip.Name, resBody.Trip.Name)
		assert.Equal(t, expectedOutput.Trip.CreatedAt.Format(time.RFC3339), resBody.Trip.CreatedAt)
		assert.Equal(t, expectedOutput.Trip.UpdatedAt.Format(time.RFC3339), resBody.Trip.UpdatedAt)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockUsecase.EXPECT().Get(gomock.Any(), tripID).Return(output.GetTripOutput{}, domain.ErrTripNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().Get(gomock.Any(), tripID).Return(output.GetTripOutput{}, errors.New("some error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTripHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.Register(r)

	now := time.Now()
	expectedTrips := []domain.Trip{
		domain.NewTrip("1", "Trip 1", now, now),
		domain.NewTrip("2", "Trip 2", now, now),
	}
	expectedOutput := output.NewListTripOutput(expectedTrips)

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().List(gomock.Any()).Return(expectedOutput, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resBody response.ListTripResponse
		err := json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.NoError(t, err)
		assert.Len(t, resBody.Trips, 2)
		assert.Equal(t, expectedOutput.Trips[0].ID, resBody.Trips[0].ID)
		assert.Equal(t, expectedOutput.Trips[0].Name, resBody.Trips[0].Name)
		assert.Equal(t, expectedOutput.Trips[0].CreatedAt.Format(time.RFC3339), resBody.Trips[0].CreatedAt)
		assert.Equal(t, expectedOutput.Trips[0].UpdatedAt.Format(time.RFC3339), resBody.Trips[0].UpdatedAt)
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().List(gomock.Any()).Return(output.ListTripOutput{}, errors.New("some error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTripHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.Register(r)

	tripName := "New Trip"

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Create(gomock.Any(), tripName).Return(nil)

		body, _ := json.Marshal(gin.H{"name": tripName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Usecase error", func(t *testing.T) {
		mockUsecase.EXPECT().Create(gomock.Any(), tripName).Return(errors.New("some error"))

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

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTripHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"
	updatedName := "Updated Trip"

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Update(gomock.Any(), tripID, updatedName).Return(nil)

		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockUsecase.EXPECT().Update(gomock.Any(), tripID, updatedName).Return(domain.ErrTripNotFound)

		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestTripHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.Register(r)

	tripID := "00000000-0000-0000-0000-000000000001"

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Delete(gomock.Any(), tripID).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Trip not found", func(t *testing.T) {
		mockUsecase.EXPECT().Delete(gomock.Any(), tripID).Return(domain.ErrTripNotFound)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
