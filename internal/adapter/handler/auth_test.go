package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"travel-api/internal/adapter/presenter"
	mock_handler "travel-api/internal/usecase/mock"
	"travel-api/internal/usecase/output"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockAuthUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	authHandler := NewAuthHandler(mockUsecase)
	authHandler.RegisterAPI(r.Group("/"))

	username := "testuser"
	email := "test@example.com"
	password := "password123"
	userID := "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"

	t.Run("正常系: ユーザー登録が成功する", func(t *testing.T) {
		expectedOutput := output.RegisterOutput{UserID: userID}
		mockUsecase.EXPECT().Register(gomock.Any(), username, email, password).Return(expectedOutput, nil).Times(1)

		body, _ := json.Marshal(gin.H{
			"username": username,
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resBody presenter.RegisterResponse
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, userID, resBody.UserID)
	})

	t.Run("異常系: バリデーションエラー (必須フィールド欠落)", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{
			"username": username,
			"email":    email,
			// password missing
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "VALIDATION_ERROR", resBody["code"])
	})
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockAuthUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	authHandler := NewAuthHandler(mockUsecase)
	authHandler.RegisterAPI(r.Group("/"))

	email := "test@example.com"
	password := "password123"
	token := "mock_jwt_token"

	t.Run("正常系: ユーザーログインが成功する", func(t *testing.T) {
		expectedOutput := output.TokenPairOutput{Token: token}
		mockUsecase.EXPECT().Login(gomock.Any(), email, password).Return(expectedOutput, nil).Times(1)

		body, _ := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resBody presenter.AuthTokenResponse
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, token, resBody.Token)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockAuthUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	authHandler := NewAuthHandler(mockUsecase)
	authHandler.RegisterAPI(r.Group("/"))

	refreshToken := "mock_refresh_token"
	newAccessToken := "new_mock_access_token"
	newRefreshToken := "new_mock_refresh_token"

	t.Run("正常系: リフレッシュトークンが有効で、新しいアクセストークンとリフレッシュトークンが返される", func(t *testing.T) {
		expectedOutput := output.TokenPairOutput{Token: newAccessToken, RefreshToken: newRefreshToken}
		mockUsecase.EXPECT().VerifyRefreshToken(gomock.Any(), refreshToken).Return(expectedOutput, nil).Times(1)

		body, _ := json.Marshal(gin.H{
			"refresh_token": refreshToken,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, newAccessToken, resBody["token"])
		assert.Equal(t, newRefreshToken, resBody["refresh_token"])
	})

	t.Run("異常系: バリデーションエラー (リフレッシュトークンが欠落している場合)", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{
			// refresh_token missing
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "VALIDATION_ERROR", resBody["code"])
	})
}
