package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"travel-api/internal/domain"
	mock_handler "travel-api/internal/interface/handler/mock"
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
	authHandler.RegisterAPI(r)

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
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "success", resBody["message"])
		assert.Equal(t, userID, resBody["data"].(map[string]interface{})["user_id"])
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

	t.Run("異常系: ユースケースエラー (ユーザーが既に存在する)", func(t *testing.T) {
		mockUsecase.EXPECT().Register(gomock.Any(), username, email, password).Return(output.RegisterOutput{}, domain.ErrUserAlreadyExists).Times(1)

		body, _ := json.Marshal(gin.H{
			"username": username,
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code) // 409 Conflict
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "USER_ALREADY_EXISTS", resBody["code"])
	})

	t.Run("異常系: ユースケースエラー (内部サーバーエラー)", func(t *testing.T) {
		mockUsecase.EXPECT().Register(gomock.Any(), username, email, password).Return(output.RegisterOutput{}, errors.New("some internal error")).Times(1)

		body, _ := json.Marshal(gin.H{
			"username": username,
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", resBody["code"])
	})
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_handler.NewMockAuthUsecase(ctrl)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	authHandler := NewAuthHandler(mockUsecase)
	authHandler.RegisterAPI(r)

	email := "test@example.com"
	password := "password123"
	token := "mock_jwt_token"

	t.Run("正常系: ユーザーログインが成功する", func(t *testing.T) {
		expectedOutput := output.LoginOutput{Token: token}
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
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "success", resBody["message"])
		assert.Equal(t, token, resBody["data"].(map[string]interface{})["token"])
	})

	t.Run("異常系: バリデーションエラー (メールアドレス形式不正)", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{
			"email":    "invalid-email",
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "VALIDATION_ERROR", resBody["code"])
	})

	t.Run("異常系: ユースケースエラー (認証情報が無効)", func(t *testing.T) {
		mockUsecase.EXPECT().Login(gomock.Any(), email, password).Return(output.LoginOutput{}, domain.ErrInvalidCredentials).Times(1)

		body, _ := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code) // 401 Unauthorized
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "INVALID_CREDENTIALS", resBody["code"])
	})

	t.Run("異常系: ユースケースエラー (内部サーバーエラー)", func(t *testing.T) {
		mockUsecase.EXPECT().Login(gomock.Any(), email, password).Return(output.LoginOutput{}, errors.New("some internal error")).Times(1)

		body, _ := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resBody map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resBody)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", resBody["code"])
	})
}
