package validations

import (
	"bytes"
	"encoding/json"

	"github.com/go-playground/validator/v10"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

func ValidateJSONObject(fieldLevel validator.FieldLevel) bool {
	raw := bytes.TrimSpace(fieldLevel.Field().Bytes())
	if len(raw) == 0 || raw[0] != '{' {
		return false
	}

	return json.Valid(raw)
}

func ValidateResumeFile(fieldLevel validator.FieldLevel) bool {
	return modelHelpers.DetectResumeContentType(fieldLevel.Field().Bytes()) != ""
}
