package handler

import (
	"net/http"
	"travel-api/internal/interface/presenter"
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

func (handler *AuthHandler) RegisterAPI(router *gin.RouterGroup) {
	router.POST("/register", handler.register)
	router.POST("/login", handler.login)
	router.POST("/refresh", handler.refresh)
}

func (handler *AuthHandler) register(c *gin.Context) {
	var body validator.RegisterJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	output, err := handler.usecase.Register(c.Request.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	c.JSON(http.StatusCreated, presenter.RegisterResponse{UserID: output.UserID})
}

func (handler *AuthHandler) login(c *gin.Context) {
	var body validator.LoginJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	output, err := handler.usecase.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	c.JSON(http.StatusOK, presenter.AuthTokenResponse{
		Token:        output.Token,
		RefreshToken: output.RefreshToken,
	})
}

func (handler *AuthHandler) refresh(c *gin.Context) {
	var body validator.RefreshTokenJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	output, err := handler.usecase.VerifyRefreshToken(c.Request.Context(), body.RefreshToken)
	if err != nil {
		c.JSON(presenter.ConvertToHTTPError(err))
		return
	}

	c.JSON(http.StatusOK, presenter.AuthTokenResponse{
		Token:        output.Token,
		RefreshToken: output.RefreshToken,
	})
}
