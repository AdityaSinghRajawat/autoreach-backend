package validations

import (
	"net"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

func ValidateExternalURL(fieldLevel validator.FieldLevel) bool {
	atsType := parentField(fieldLevel, config.GetATSTypeFieldName())
	if !atsType.IsValid() || atsType.Kind() != reflect.String {
		return false
	}

	return IsAllowedExternalURL(fieldLevel.Field().String(), atsType.String())
}

func ValidateJobTitle(fieldLevel validator.FieldLevel) bool {
	return modelHelpers.NormalizeJobTitle(fieldLevel.Field().String()) != ""
}

func ValidateSalaryRange(fieldLevel validator.FieldLevel) bool {
	salaryMin := parentField(fieldLevel, config.GetSalaryMinFieldName())
	if !salaryMin.IsValid() || salaryMin.Kind() != reflect.Float64 {
		return true
	}

	return fieldLevel.Field().Float() >= salaryMin.Float()
}

func parentField(fieldLevel validator.FieldLevel, name string) reflect.Value {
	parent := reflect.Indirect(fieldLevel.Parent())
	if parent.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	return reflect.Indirect(parent.FieldByName(name))
}

func IsAllowedExternalURL(rawURL, atsType string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if !isPublicHostname(host) {
		return false
	}

	suffixes, restricted := config.GetATSHostSuffixes(atsType)
	if !restricted {
		return true
	}

	for _, suffix := range suffixes {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}

	return false
}

func isPublicHostname(host string) bool {
	if host == "" || !strings.Contains(host, ".") || net.ParseIP(host) != nil {
		return false
	}

	for _, reserved := range []string{"localhost", ".local", ".internal", ".localhost"} {
		if host == reserved || strings.HasSuffix(host, reserved) {
			return false
		}
	}

	return true
}

func ValidateBoardToken(fieldLevel validator.FieldLevel) bool {
	boardType := parentField(fieldLevel, config.GetBoardTypeFieldName()).String()
	token := fieldLevel.Field().String()

	if boardType == config.GetCareerPageBoardType() {
		return IsAllowedExternalURL(token, config.GetOtherATSType())
	}

	pattern, found := config.GetBoardTokenPattern(boardType)

	return found && pattern.MatchString(token)
}

// ValidateBoardTypes accepts the field only on the board types listed in the tag's parameter.
func ValidateBoardTypes(fieldLevel validator.FieldLevel) bool {
	boardType := parentField(fieldLevel, config.GetBoardTypeFieldName()).String()

	return slices.Contains(strings.Fields(fieldLevel.Param()), boardType)
}
