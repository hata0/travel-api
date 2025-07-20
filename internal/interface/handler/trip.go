package handler

import (
	"net/http"
	"travel-api/internal/interface/response"
	"travel-api/internal/interface/validator"
	"travel-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type TripHandler struct {
	usecase usecase.TripUsecase
}

func NewTripHandler(usecase usecase.TripUsecase) *TripHandler {
	return &TripHandler{
		usecase: usecase,
	}
}

func (handler *TripHandler) RegisterAPI(router *gin.Engine) {
	router.GET("/trips/:trip_id", handler.get)
	router.GET("/trips", handler.list)
	router.POST("/trips", handler.create)
	router.PUT("/trips/:trip_id", handler.update)
	router.DELETE("/trips/:trip_id", handler.delete)
}

func (handler *TripHandler) get(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.ShouldBindUri(&uriParams); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	tripOutput, err := handler.usecase.Get(c.Request.Context(), uriParams.TripID)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.NewGetTripResponse(tripOutput))
}

func (handler *TripHandler) list(c *gin.Context) {
	tripsOutput, err := handler.usecase.List(c.Request.Context())
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.NewListTripResponse(tripsOutput))
}

func (handler *TripHandler) create(c *gin.Context) {
	var body validator.CreateTripJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	createdTripID, err := handler.usecase.Create(c.Request.Context(), body.Name)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusCreated, response.CreateTripResponse{ID: createdTripID})
}

func (handler *TripHandler) update(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.ShouldBindUri(&uriParams); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	var body validator.UpdateTripJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	err := handler.usecase.Update(c.Request.Context(), uriParams.TripID, body.Name)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "success"})
}

func (handler *TripHandler) delete(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.ShouldBindUri(&uriParams); err != nil {
		response.NewError(err).JSON(c)
		return
	}

	err := handler.usecase.Delete(c.Request.Context(), uriParams.TripID)
	if err != nil {
		response.NewError(err).JSON(c)
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{Message: "success"})
}
