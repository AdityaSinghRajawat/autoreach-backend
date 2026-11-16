package validations

import (
	"net/mail"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var fqdnValidatorI = validator.New()

// ValidateSuppressionValue checks an EMAIL entry holds an address and a DOMAIN entry a domain.
func ValidateSuppressionValue(fieldLevel validator.FieldLevel) bool {
	value := strings.TrimSpace(fieldLevel.Field().String())

	switch parentField(fieldLevel, config.GetEntryTypeFieldName()).String() {
	case "EMAIL":
		address, err := mail.ParseAddress(value)

		return err == nil && address.Address == value
	case "DOMAIN":
		return fqdnValidatorI.Var(value, "fqdn") == nil
	default:
		return false
	}
}
