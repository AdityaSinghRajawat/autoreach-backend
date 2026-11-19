package hunter

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

// Only a company whose name matches exactly is accepted, because Hunter also suggests look-alikes.
func (p *HunterProvider) ResolveCompanyDomain(
	ctx context.Context,
	companyName string,
) (string, error) {
	response, err := p.integration.FindDomains(
		ctx,
		companyName,
		config.GetCompanyDomainCandidates(),
	)
	if err != nil {
		return "", err
	}

	wanted := modelHelpers.NormalizeCompanyName(companyName)

	for _, candidate := range response.Data {
		if candidate.Domain != "" &&
			modelHelpers.NormalizeCompanyName(candidate.CompanyName) == wanted {
			return modelHelpers.NormalizeDomain(candidate.Domain), nil
		}
	}

	return "", nil
}
