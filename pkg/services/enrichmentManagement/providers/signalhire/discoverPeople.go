package signalhire

import (
	"context"
	"slices"
	"strings"
	"unicode"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	signalhireIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/signalhire"
)

// Finds people by company name and title, so contacts can be found without Apollo.
func (p *SignalHireProvider) DiscoverPeople(
	ctx context.Context,
	query enrichmentTypes.DiscoveryQuery,
) ([]enrichmentTypes.Person, error) {
	companyWords := companySearchWords(query.CompanyName)
	if len(companyWords) == 0 {
		return nil, nil
	}

	response, err := p.integration.SearchPeople(
		ctx,
		strings.Join(companyWords, " AND "),
		titleQuery(query.Titles),
		min(max(query.Limit, 1), config.GetSignalHireMaxSearchSize()),
	)
	if err != nil {
		return nil, err
	}

	people := make([]enrichmentTypes.Person, 0, len(response.Profiles))

	for _, profile := range response.Profiles {
		title := currentTitle(profile, companyWords)
		if profile.UID == "" || title == "" {
			continue
		}

		people = append(people, enrichmentTypes.Person{
			Source:           config.ProviderSignalHire,
			ProviderPersonID: profile.UID,
			FullName:         strings.TrimSpace(profile.FullName),
			Title:            title,
			CompanyName:      query.CompanyName,
		})
	}

	return people, nil
}

// "Acme Labs Pvt Ltd" searches as Acme AND Labs.
func companySearchWords(companyName string) []string {
	words := queryWords(companyName)

	significant := slices.DeleteFunc(slices.Clone(words), config.IsCompanyLegalSuffix)
	if len(significant) > 0 {
		return significant
	}

	return words
}

// ["engineering manager", "recruiter"] searches as (engineering AND manager) OR recruiter.
func titleQuery(titles []string) string {
	terms := make([]string, 0, len(titles))

	for _, title := range titles {
		words := queryWords(title)

		switch {
		case len(words) == 0:
			continue
		case len(words) == 1:
			terms = append(terms, words[0])
		default:
			terms = append(terms, "("+strings.Join(words, " AND ")+")")
		}
	}

	return strings.Join(slices.Compact(terms), " OR ")
}

// Keeps letters and digits only, so names can't inject boolean syntax.
func queryWords(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	return slices.DeleteFunc(words, config.IsBooleanOperator)
}

// The title held at the searched company, or the most recent one.
func currentTitle(profile signalhireIntegration.Profile, companyWords []string) string {
	for _, job := range profile.Experience {
		company := strings.ToLower(job.Company)
		if strings.TrimSpace(job.Title) != "" &&
			strings.Contains(company, strings.ToLower(companyWords[0])) {
			return strings.TrimSpace(job.Title)
		}
	}

	if len(profile.Experience) > 0 {
		return strings.TrimSpace(profile.Experience[0].Title)
	}

	return ""
}
