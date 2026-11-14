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
	"validate_timezone":     ValidateTimezone,
	"validate_linkedin_url": ValidateLinkedinURL,
	"validate_external_url": ValidateExternalURL,
	"validate_salary_range": ValidateSalaryRange,
	"validate_job_title":    ValidateJobTitle,
	"validate_json_object":  ValidateJSONObject,
	"validate_resume_file":  ValidateResumeFile,
}

var customIssuesI = map[string]string{
	"validate_timezone":     "must be a valid IANA time zone",
	"validate_linkedin_url": "must be an https linkedin.com URL",
	"validate_external_url": "must be a public https URL on a host that matches atsType",
	"validate_salary_range": "must be greater than or equal to salaryMin",
	"validate_job_title":    "must contain letters or digits",
	"validate_json_object":  "must be a JSON object",
	"validate_resume_file":  "must be a PDF or DOCX document",
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
