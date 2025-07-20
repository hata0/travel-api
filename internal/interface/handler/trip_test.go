package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"travel-api/internal/domain"
	"travel-api/internal/interface/response"
	mock_handler "travel-api/internal/usecase/mock"
	"travel-api/internal/usecase/output"

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
	tripHandler.RegisterAPI(r)

	tripID := "00000000-0000-0000-0000-000000000001"
	now := time.Now()
	tripIDValue, _ := domain.NewTripID(tripID)
	expectedTrip := domain.NewTrip(tripIDValue, "Test Trip", now, now)
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
		assert.WithinDuration(t, expectedOutput.Trip.CreatedAt, resBody.Trip.CreatedAt, time.Second)
		assert.WithinDuration(t, expectedOutput.Trip.UpdatedAt, resBody.Trip.UpdatedAt, time.Second)
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
	tripHandler.RegisterAPI(r)

	now := time.Now()
	tripID1, _ := domain.NewTripID("00000000-0000-0000-0000-000000000001")
	tripID2, _ := domain.NewTripID("00000000-0000-0000-0000-000000000002")
	expectedTrips := []domain.Trip{
		domain.NewTrip(tripID1, "Trip 1", now, now),
		domain.NewTrip(tripID2, "Trip 2", now, now),
	}
	expectedOutput := output.NewListTripOutput(expectedTrips)

	t.Run("正常系: 複数のレコードが存在する", func(t *testing.T) {
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
		assert.WithinDuration(t, expectedOutput.Trips[0].CreatedAt, resBody.Trips[0].CreatedAt, time.Second)
		assert.WithinDuration(t, expectedOutput.Trips[0].UpdatedAt, resBody.Trips[0].UpdatedAt, time.Second)
	})

	t.Run("正常系: レコードが存在しない", func(t *testing.T) {
		mockUsecase.EXPECT().List(gomock.Any()).Return(output.ListTripOutput{Trips: []output.Trip{}}, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/trips", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resBody response.ListTripResponse
		err := json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.NoError(t, err)
		assert.Empty(t, resBody.Trips)
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
	tripHandler.RegisterAPI(r)

	tripName := "New Trip"

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Create(gomock.Any(), tripName).Return("new-id", nil)

		body, _ := json.Marshal(gin.H{"name": tripName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("異常系: Usecase error (Internal Server Error)", func(t *testing.T) {
		mockUsecase.EXPECT().Create(gomock.Any(), tripName).Return("", errors.New("some error"))

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
		var resBody map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", resBody["code"])
	})
}

func TestTripHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.RegisterAPI(r)

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

	t.Run("異常系: 無効なJSONボディ", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer([]byte(`{"name":`)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "VALIDATION_ERROR", resBody["code"])
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().Update(gomock.Any(), tripID, updatedName).Return(errors.New("some error"))

		body, _ := json.Marshal(gin.H{"name": updatedName})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/trips/"+tripID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTripHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockTripUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	tripHandler := NewTripHandler(mockUsecase)
	tripHandler.RegisterAPI(r)

	tripID := "00000000-0000-0000-0000-000000000001"

	t.Run("正常系", func(t *testing.T) {
		mockUsecase.EXPECT().Delete(gomock.Any(), tripID).Return(nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異常系: Internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().Delete(gomock.Any(), tripID).Return(errors.New("some error"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/trips/"+tripID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
