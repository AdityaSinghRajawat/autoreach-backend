package validations

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

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
	if issue, found := config.GetValidationIssue(fieldError.Tag()); found {
		return issue
	}

	if prefix, found := config.GetParameterisedValidationIssue(fieldError.Tag()); found {
		return prefix + fieldError.Param()
	}

	return config.GetDefaultValidationIssue()
}
