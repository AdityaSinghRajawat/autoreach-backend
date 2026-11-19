package hunter

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

// Only a company whose name matches is accepted, because Hunter also suggests look-alikes.
func (p *HunterProvider) ResolveCompanyDomain(
	ctx context.Context,
	companyName string,
) (string, error) {
	if utf8.RuneCountInString(
		strings.TrimSpace(companyName),
	) < config.GetHunterMinCompanyNameLength() {
		return "", nil
	}

	response, err := p.integration.FindDomains(
		ctx,
		companyName,
		config.GetCompanyDomainCandidates(),
	)
	if err != nil {
		return "", err
	}

	for _, candidate := range response.Data {
		if candidate.Domain != "" &&
			modelHelpers.CompanyNamesMatch(candidate.CompanyName, companyName) {
			return modelHelpers.NormalizeDomain(candidate.Domain), nil
		}
	}

	return "", nil
}
