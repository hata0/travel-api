package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"travel-api/domain"
	"travel-api/interface/response"
	"travel-api/interface/validator"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -destination mock/trip.go travel-api/interface/controller TripService
type TripService interface {
	Get(ctx context.Context, id string) (domain.Trip, error)
	List(ctx context.Context) ([]domain.Trip, error)
	Create(ctx context.Context, name string) error
	Update(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type TripController struct {
	service TripService
}

func NewTripController(service TripService) *TripController {
	return &TripController{
		service: service,
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

	trip, err := controller.service.Get(c.Request.Context(), uriParams.TripID)
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
		Trip: controller.mapToTrip(trip),
	})
}

func (controller *TripController) List(c *gin.Context) {
	trips, err := controller.service.List(c.Request.Context())
	if err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	formattedTrips := make([]response.Trip, len(trips))
	for i, trip := range trips {
		formattedTrips[i] = controller.mapToTrip(trip)
	}

	c.JSON(http.StatusOK, response.ListTripResponse{
		Trips: formattedTrips,
	})
}

func (controller *TripController) mapToTrip(trip domain.Trip) response.Trip {
	return response.Trip{
		ID:        string(trip.ID),
		Name:      trip.Name,
		CreatedAt: trip.CreatedAt.Format(time.RFC3339),
		UpdatedAt: trip.UpdatedAt.Format(time.RFC3339),
	}
}

func (controller *TripController) Create(c *gin.Context) {
	var body validator.CreateTripJSONBody
	if err := c.BindJSON(&body); err != nil {
		response.NewError(domain.ErrInternalServerError, http.StatusInternalServerError).JSON(c)
		return
	}

	err := controller.service.Create(c.Request.Context(), body.Name)
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

	err := controller.service.Update(c.Request.Context(), uriParams.TripID, body.Name)
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

	err := controller.service.Delete(c.Request.Context(), uriParams.TripID)
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
