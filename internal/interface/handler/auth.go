package handler

import (
	"net/http"
	"travel-api/internal/interface/response"
	"travel-api/internal/interface/validator"
	"travel-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	usecase usecase.AuthUsecase
}

func NewAuthHandler(usecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: usecase,
	}
}

func (handler *AuthHandler) RegisterAPI(router *gin.Engine) {
	router.POST("/register", handler.register)
	router.POST("/login", handler.login)
	router.POST("/refresh", handler.refresh)
}

func (handler *AuthHandler) register(c *gin.Context) {
	var body validator.RegisterJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	output, err := handler.usecase.Register(c.Request.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusCreated, response.RegisterResponse{UserID: output.UserID})
}

func (handler *AuthHandler) login(c *gin.Context) {
	var body validator.LoginJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	output, err := handler.usecase.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.AuthTokenResponse{
		Token:        output.Token,
		RefreshToken: output.RefreshToken,
	})
}

func (handler *AuthHandler) refresh(c *gin.Context) {
	var body validator.RefreshTokenJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	output, err := handler.usecase.VerifyRefreshToken(c.Request.Context(), body.RefreshToken)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.AuthTokenResponse{
		Token:        output.Token,
		RefreshToken: output.RefreshToken,
	})
}
