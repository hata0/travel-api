package validator

type TripURIParameters struct {
	TripID string `uri:"trip_id" binding:"required,uuid4"`
}