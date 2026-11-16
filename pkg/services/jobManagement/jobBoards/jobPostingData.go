package jobBoards

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// jobPostingsOnPage reads schema.org JobPosting structured data, the format career sites publish for
// Google's job search.
func jobPostingsOnPage(pageURL string, page []byte) []jobTypes.BoardJob {
	jobs := []jobTypes.BoardJob{}

	for _, posting := range utils.ExtractJSONLDObjects(page, "JobPosting") {
		jobURL := resolveURL(pageURL, text(posting["url"]))
		job := jobTypes.BoardJob{
			ExternalID:  utils.FirstNonEmpty(identifier(posting["identifier"]), jobURL),
			Title:       html.UnescapeString(text(posting["title"])),
			Description: utils.HTMLToText(text(posting["description"])),
			URL:         jobURL,
			Location:    utils.NilIfEmpty(jobLocation(posting)),
			PostedAt:    postedAt(text(posting["datePosted"])),
		}

		if salary, found := posting["baseSalary"].(map[string]any); found {
			job.SalaryMin, job.SalaryMax, job.SalaryCurrency = salaryRange(salary)
		}

		jobs = append(jobs, job)
	}

	return jobs
}

func text(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return fmt.Sprint(typed)
	case map[string]any:
		return utils.FirstNonEmpty(text(typed["name"]), text(typed["@id"]))
	}

	return ""
}

func identifier(value any) string {
	if object, isObject := value.(map[string]any); isObject {
		return text(object["value"])
	}

	return text(value)
}

func resolveURL(pageURL, link string) string {
	base, err := url.Parse(pageURL)
	if err != nil || link == "" {
		return pageURL
	}

	resolved, err := base.Parse(link)
	if err != nil {
		return pageURL
	}

	return resolved.String()
}

func jobLocation(posting map[string]any) string {
	if text(posting["jobLocationType"]) == "TELECOMMUTE" {
		return "Remote"
	}

	locations, isList := posting["jobLocation"].([]any)
	if !isList {
		locations = []any{posting["jobLocation"]}
	}

	names := []string{}

	for _, location := range locations {
		place, isPlace := location.(map[string]any)
		if !isPlace {
			continue
		}

		address, isAddress := place["address"].(map[string]any)
		if !isAddress {
			continue
		}

		parts := []string{}

		for _, key := range []string{"addressLocality", "addressRegion", "addressCountry"} {
			if part := text(address[key]); part != "" {
				parts = append(parts, part)
			}
		}

		if len(parts) > 0 {
			names = append(names, strings.Join(parts, ", "))
		}
	}

	return strings.Join(names, "; ")
}

func postedAt(value string) *time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}

	return nil
}

func salaryRange(salary map[string]any) (*float64, *float64, *string) {
	currency := strings.ToUpper(text(salary["currency"]))
	if len(currency) != 3 {
		return nil, nil, nil
	}

	value, isObject := salary["value"].(map[string]any)
	if !isObject {
		return nil, nil, nil
	}

	minimum, maximum := number(value["minValue"]), number(value["maxValue"])
	if minimum == nil && maximum == nil {
		minimum = number(value["value"])
		maximum = minimum
	}

	if minimum == nil && maximum == nil {
		return nil, nil, nil
	}

	return minimum, maximum, &currency
}

func number(value any) *float64 {
	if typed, isNumber := value.(float64); isNumber {
		return &typed
	}

	return nil
}
