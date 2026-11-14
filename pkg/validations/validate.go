package validations

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var issueByTagI = map[string]string{
	"required": "is required",
	"email":    "must be a valid email address",
	"uuid":     "must be a valid UUID",
	"url":      "must be a valid URL",
	"fqdn":     "must be a valid domain name",
}

var parameterisedIssueByTagI = map[string]string{
	"min":   "must be at least ",
	"max":   "must be at most ",
	"gte":   "must be greater than or equal to ",
	"lte":   "must be less than or equal to ",
	"gt":    "must be greater than ",
	"lt":    "must be less than ",
	"len":   "must have length ",
	"oneof": "must be one of: ",
}

func Validate(payload any) error {
	return GetValidator().Struct(payload)
}

func BuildErrorDetails(err error) []config.ErrorDetail {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	details := make([]config.ErrorDetail, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		details = append(details, config.ErrorDetail{
			Field: fieldPath(fieldError),
			Issue: issueFor(fieldError),
		})
	}

	return details
}

func fieldPath(fieldError validator.FieldError) string {
	namespace := fieldError.Namespace()
	if separator := strings.Index(namespace, "."); separator >= 0 {
		return namespace[separator+1:]
	}

	return fieldError.Field()
}

func issueFor(fieldError validator.FieldError) string {
	if issue, found := issueByTagI[fieldError.Tag()]; found {
		return issue
	}

	if issue, found := customIssuesI[fieldError.Tag()]; found {
		return issue
	}

	if prefix, found := parameterisedIssueByTagI[fieldError.Tag()]; found {
		return prefix + fieldError.Param()
	}

	return "is invalid"
}
