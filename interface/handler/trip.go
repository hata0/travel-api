package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"travel-api/domain"
	"travel-api/interface/response"
	"travel-api/interface/validator"
	"travel-api/usecase/output"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination mock/trip.go travel-api/interface/handler TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (output.GetTripOutput, error)
	List(ctx context.Context) (output.ListTripOutput, error)
	Create(ctx context.Context, name string) error
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripHandler struct {
	usecase TripUsecase
}

func NewTripHandler(usecase TripUsecase) *TripHandler {
	return &TripHandler{
		usecase: usecase,
	}
}

func (handler *TripHandler) Register(router *gin.Engine) {
	router.GET("/trips/:trip_id", handler.Get)
	router.GET("/trips", handler.List)
	router.POST("/trips", handler.Create)
	router.PUT("/trips/:trip_id", handler.Update)
	router.DELETE("/trips/:trip_id", handler.Delete)
}

func (handler *TripHandler) Get(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.BindUri(&uriParams); err != nil {
		fmt.Println(err)
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	tripOutput, err := handler.usecase.Get(c.Request.Context(), uriParams.TripID)
	if err != nil {
		switch err {
		case domain.ErrTripNotFound:
			response.NewError(err, http.StatusNotFound).JSON(c)
		default:
			response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.GetTripResponse{
		Trip: response.Trip{
			ID:        tripOutput.Trip.ID,
			Name:      tripOutput.Trip.Name,
			CreatedAt: tripOutput.Trip.CreatedAt.Format(time.RFC3339),
			UpdatedAt: tripOutput.Trip.UpdatedAt.Format(time.RFC3339),
		},
	})
}

func (handler *TripHandler) List(c *gin.Context) {
	tripsOutput, err := handler.usecase.List(c.Request.Context())
	if err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	formattedTrips := make([]response.Trip, len(tripsOutput.Trips))
	for i, trip := range tripsOutput.Trips {
		formattedTrips[i] = response.Trip{
			ID:        trip.ID,
			Name:      trip.Name,
			CreatedAt: trip.CreatedAt.Format(time.RFC3339),
			UpdatedAt: trip.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, response.ListTripResponse{
		Trips: formattedTrips,
	})
}

func (handler *TripHandler) Create(c *gin.Context) {
	var body validator.CreateTripJSONBody
	if err := c.BindJSON(&body); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := handler.usecase.Create(c.Request.Context(), body.Name)
	if err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	response.NewSuccess(domain.SuccessMessage, http.StatusOK).JSON(c)
}

func (handler *TripHandler) Update(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.BindUri(&uriParams); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	var body validator.UpdateTripJSONBody
	if err := c.BindJSON(&body); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := handler.usecase.Update(c.Request.Context(), uriParams.TripID, body.Name)
	if err != nil {
		switch err {
		case domain.ErrTripNotFound:
			response.NewError(err, http.StatusNotFound).JSON(c)
		default:
			response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		}
		return
	}

	response.NewSuccess(domain.SuccessMessage, http.StatusOK).JSON(c)
}

func (handler *TripHandler) Delete(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.BindUri(&uriParams); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := handler.usecase.Delete(c.Request.Context(), uriParams.TripID)
	if err != nil {
		switch err {
		case domain.ErrTripNotFound:
			response.NewError(err, http.StatusNotFound).JSON(c)
		default:
			response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		}
		return
	}

	response.NewSuccess(domain.SuccessMessage, http.StatusOK).JSON(c)
}
