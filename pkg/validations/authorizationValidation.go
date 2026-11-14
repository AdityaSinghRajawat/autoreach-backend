package validations

import (
	"github.com/go-playground/validator/v10"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func ValidateRole(fieldLevel validator.FieldLevel) bool {
	return config.IsKnownRole(fieldLevel.Field().String())
}

func ValidateAPIClientGrant(fieldLevel validator.FieldLevel) bool {
	return config.IsAPIClientGrant(fieldLevel.Field().String())
}
