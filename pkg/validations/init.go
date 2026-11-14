package validations

import (
	"reflect"
	"strings"
	"sync"
	_ "time/tzdata"

	"github.com/go-playground/validator/v10"
)

var (
	validateI    *validator.Validate
	validateOnce sync.Once
)

var customValidationsI = map[string]validator.Func{
	"validate_timezone":         ValidateTimezone,
	"validate_linkedin_url":     ValidateLinkedinURL,
	"validate_external_url":     ValidateExternalURL,
	"validate_salary_range":     ValidateSalaryRange,
	"validate_job_title":        ValidateJobTitle,
	"validate_json_object":      ValidateJSONObject,
	"validate_resume_file":      ValidateResumeFile,
	"validate_role":             ValidateRole,
	"validate_api_client_grant": ValidateAPIClientGrant,
}

func GetValidator() *validator.Validate {
	validateOnce.Do(func() {
		validateI = validator.New(validator.WithRequiredStructEnabled())
		validateI.RegisterTagNameFunc(jsonFieldName)
		registerCustomValidations(validateI)
	})

	return validateI
}

func registerCustomValidations(validate *validator.Validate) {
	for tag, rule := range customValidationsI {
		if err := validate.RegisterValidation(tag, rule); err != nil {
			panic("register validation " + tag + ": " + err.Error())
		}
	}
}

func jsonFieldName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "" || name == "-" {
		name, _, _ = strings.Cut(field.Tag.Get("form"), ",")
	}

	if name == "" || name == "-" {
		return field.Name
	}

	return name
}
