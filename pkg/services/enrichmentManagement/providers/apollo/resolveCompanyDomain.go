package apollo

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

// ResolveCompanyDomain accepts only an organisation whose name matches exactly, because Apollo's
// name search also returns partial matches.
func (p *ApolloProvider) ResolveCompanyDomain(
	ctx context.Context,
	companyName string,
) (string, error) {
	response, err := p.integration.SearchOrganizations(
		ctx,
		companyName,
		config.GetCompanyDomainCandidates(),
	)
	if err != nil {
		return "", err
	}

	for _, organization := range append(response.Organizations, response.Accounts...) {
		if organization.PrimaryDomain != "" &&
			modelHelpers.CompanyNamesMatch(organization.Name, companyName) {
			return modelHelpers.NormalizeDomain(organization.PrimaryDomain), nil
		}
	}

	return "", nil
}
