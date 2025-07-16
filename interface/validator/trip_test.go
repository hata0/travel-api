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
			TripID: "00000000-0000-0000-0000-000000000001",
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
}

func TestCreateTripJSONBody_Validation(t *testing.T) {
	validate := validator.New()
	validate.SetTagName("binding")

	t.Run("正常系", func(t *testing.T) {
		params := CreateTripJSONBody{
			Name: "test name",
		}
		err := validate.Struct(params)
		assert.NoError(t, err)
	})

	t.Run("異常系: Nameが空", func(t *testing.T) {
		params := CreateTripJSONBody{
			Name: "",
		}
		err := validate.Struct(params)
		assert.Error(t, err)
	})
}
