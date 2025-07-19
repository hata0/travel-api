package handler

import (
	"context"
	"log/slog"
	"net/http"
	"travel-api/internal/domain"
	"travel-api/internal/interface/response"
	"travel-api/internal/interface/validator"
	"travel-api/internal/usecase/output"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination mock/auth.go travel-api/internal/interface/handler AuthUsecase
type AuthUsecase interface {
	Register(ctx context.Context, username, email, password string) (output.RegisterOutput, error)
	Login(ctx context.Context, email, password string) (output.LoginOutput, error)
}

type AuthHandler struct {
	usecase AuthUsecase
}

func NewAuthHandler(usecase AuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: usecase,
	}
}

func (handler *AuthHandler) RegisterAPI(router *gin.Engine) {
	router.POST("/register", handler.register)
	router.POST("/login", handler.login)
}

func (handler *AuthHandler) register(c *gin.Context) {
	var body validator.RegisterJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	output, err := handler.usecase.Register(c.Request.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		slog.Error("Failed to register user", "error", err)
		response.NewError(err).JSON(c)
		return
	}

	response.NewSuccessWithData(domain.SuccessMessage, http.StatusCreated, gin.H{"user_id": output.UserID}).JSON(c)
}

func (handler *AuthHandler) login(c *gin.Context) {
	var body validator.LoginJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	output, err := handler.usecase.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		slog.Error("Failed to login user", "error", err)
		response.NewError(err).JSON(c)
		return
	}

	response.NewSuccessWithData(domain.SuccessMessage, http.StatusOK, gin.H{"token": output.Token}).JSON(c)
}
