package controller

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

//go:generate mockgen -destination mock/trip.go travel-api/interface/controller TripUsecase
type TripUsecase interface {
	Get(ctx context.Context, id string) (output.GetTripOutput, error)
	List(ctx context.Context) (output.ListTripOutput, error)
	Create(ctx context.Context, name string) error
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripController struct {
	usecase TripUsecase
}

func NewTripController(usecase TripUsecase) *TripController {
	return &TripController{
		usecase: usecase,
	}
}

func (controller *TripController) Register(router *gin.Engine) {
	router.GET("/trips/:trip_id", controller.Get)
	router.GET("/trips", controller.List)
	router.POST("/trips", controller.Create)
	router.PUT("/trips/:trip_id", controller.Update)
	router.DELETE("/trips/:trip_id", controller.Delete)
}

func (controller *TripController) Get(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.BindUri(&uriParams); err != nil {
		fmt.Println(err)
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	tripOutput, err := controller.usecase.Get(c.Request.Context(), uriParams.TripID)
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

func (controller *TripController) List(c *gin.Context) {
	tripsOutput, err := controller.usecase.List(c.Request.Context())
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

func (controller *TripController) Create(c *gin.Context) {
	var body validator.CreateTripJSONBody
	if err := c.BindJSON(&body); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := controller.usecase.Create(c.Request.Context(), body.Name)
	if err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	response.NewSuccess(domain.SuccessMessage, http.StatusOK).JSON(c)
}

func (controller *TripController) Update(c *gin.Context) {
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

	err := controller.usecase.Update(c.Request.Context(), uriParams.TripID, body.Name)
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

func (controller *TripController) Delete(c *gin.Context) {
	var uriParams validator.TripURIParameters
	if err := c.BindUri(&uriParams); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := controller.usecase.Delete(c.Request.Context(), uriParams.TripID)
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
