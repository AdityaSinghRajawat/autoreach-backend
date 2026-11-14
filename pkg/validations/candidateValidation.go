package validations

import (
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const linkedinHost = "linkedin.com"

func ValidateTimezone(fieldLevel validator.FieldLevel) bool {
	name := fieldLevel.Field().String()
	if name == "" || name == "Local" {
		return false
	}

	_, err := time.LoadLocation(name)

	return err == nil
}

func ValidateLinkedinURL(fieldLevel validator.FieldLevel) bool {
	raw := fieldLevel.Field().String()
	if raw == "" {
		return true
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	return host == linkedinHost || strings.HasSuffix(host, "."+linkedinHost)
}
