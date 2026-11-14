package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

func BindAndValidate(c *gin.Context, payload any) *config.CustomError {
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(payload); err != nil {
		return decodeFailure(err)
	}

	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return config.NewCustomError(
			fmt.Errorf("decode request body: %w", config.ErrTrailingRequestData),
			config.GetErrCodeBadRequest(),
		)
	}

	return validatePayload(payload)
}

func BindQueryAndValidate(c *gin.Context, params any) *config.CustomError {
	if err := c.ShouldBindQuery(params); err != nil {
		return config.NewCustomError(
			fmt.Errorf("bind query parameters: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	return validatePayload(params)
}

type FileTarget interface {
	SetFile(fileName string, content []byte)
}

func BindFileAndValidate(
	c *gin.Context,
	field string,
	maxBytes int64,
	target FileTarget,
) *config.CustomError {
	header, err := c.FormFile(field)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return config.NewCustomError(
				fmt.Errorf("read uploaded file: %w", err),
				config.GetErrCodePayloadTooLarge(),
			)
		}

		return config.NewValidationError(
			fmt.Errorf("read uploaded file: %w", err),
			[]config.ErrorDetail{{Field: field, Issue: "is required"}},
		)
	}

	if header.Size > maxBytes {
		return config.NewCustomError(
			fmt.Errorf("read uploaded file: %w", config.ErrResumeFileTooLarge),
			config.GetErrCodePayloadTooLarge(),
		)
	}

	file, err := header.Open()
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("open uploaded file: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("read uploaded file: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	if int64(len(content)) > maxBytes {
		return config.NewCustomError(
			fmt.Errorf("read uploaded file: %w", config.ErrResumeFileTooLarge),
			config.GetErrCodePayloadTooLarge(),
		)
	}

	target.SetFile(header.Filename, content)

	return validatePayload(target)
}

func validatePayload(payload any) *config.CustomError {
	if err := validations.Validate(payload); err != nil {
		details := validations.BuildErrorDetails(err)
		if details == nil {
			return config.NewCustomError(
				fmt.Errorf("validate request: %w", err),
				config.GetErrCodeBadRequest(),
			)
		}

		return config.NewValidationError(fmt.Errorf("validate request: %w", err), details)
	}

	return nil
}

func decodeFailure(err error) *config.CustomError {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return config.NewCustomError(
			fmt.Errorf("decode request body: %w", err),
			config.GetErrCodePayloadTooLarge(),
		)
	}

	return config.NewCustomError(
		fmt.Errorf("decode request body: %w", err),
		config.GetErrCodeBadRequest(),
	)
}
