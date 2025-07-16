package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestTripURIParameters_Validation(t *testing.T) {
	validate := validator.New()
	validate.SetTagName("binding")

	t.Run("正常系", func(t *testing.T) {
		params := TripURIParameters{
			TripID: "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d",
		}
		err := validate.Struct(params)
		assert.NoError(t, err)
	})

	t.Run("異常系: TripIDが空", func(t *testing.T) {
		params := TripURIParameters{
			TripID: "",
		}
		err := validate.Struct(params)
		assert.Error(t, err)
	})

	t.Run("異常系: TripIDがUUIDv4形式ではない", func(t *testing.T) {
		params := TripURIParameters{
			TripID: "not-a-valid-uuid",
		}
		err := validate.Struct(params)
		assert.Error(t, err)
	})
}
