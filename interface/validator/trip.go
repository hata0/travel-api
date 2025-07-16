package validator

type TripURIParameters struct {
	TripID string `uri:"trip_id" binding:"required"`
}

type CreateTripJSONBody struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTripJSONBody struct {
	Name string `json:"name" binding:"required"`
}
