package enrichmentManagement

import (
	"context"
	"fmt"
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// ResolveCompanyDomains maps company names to website domains. Names are looked up once and cached;
// failed lookups are retried after COMPANY_DOMAIN_RETRY_AFTER. Names missing from the result could
// not be resolved, or exceeded this call's lookup limit and wait for a later call.
func (s *EnrichmentService) ResolveCompanyDomains(
	ctx context.Context,
	companyNames []string,
) (map[string]string, error) {
	namesByKey := map[string][]string{}

	for _, name := range companyNames {
		if key := modelHelpers.NormalizeCompanyName(name); key != "" {
			namesByKey[key] = append(namesByKey[key], name)
		}
	}

	keys := make([]string, 0, len(namesByKey))
	for key := range namesByKey {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	cached, err := s.companyDomainRepo.ListByNormalizedNames(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("resolve company domains: %w", err)
	}

	entries := make(map[string]*enrichmentModel.CompanyDomain, len(cached))
	for _, entry := range cached {
		entries[entry.NormalizedName] = entry
	}

	resolver := s.domainResolver()
	retryBefore := utils.GetCurrTimeStamp().Add(-config.GetCompanyDomainRetryAfter())
	lookupsLeft := config.GetCompanyDomainLookupsPerPoll()
	domains := map[string]string{}

	for _, key := range keys {
		entry := entries[key]

		if entry == nil || (entry.Domain == nil && entry.UpdatedAt.Before(retryBefore)) {
			if resolver == nil || lookupsLeft == 0 {
				continue
			}

			lookupsLeft--

			entry, err = s.lookUpCompanyDomain(ctx, resolver, key, namesByKey[key][0], entry)
			if isProviderSkipped(err) {
				lookupsLeft = 0

				continue
			}

			if err != nil {
				utils.LogWarn(
					ctx,
					"company domain lookup failed",
					"company",
					key,
					"error",
					err.Error(),
				)

				continue
			}
		}

		if entry.Domain != nil {
			for _, name := range namesByKey[key] {
				domains[name] = *entry.Domain
			}
		}
	}

	return domains, nil
}

func (s *EnrichmentService) lookUpCompanyDomain(
	ctx context.Context,
	resolver providers.DomainResolver,
	normalizedName, companyName string,
	entry *enrichmentModel.CompanyDomain,
) (*enrichmentModel.CompanyDomain, error) {
	var domain string

	err := s.callProvider(
		ctx,
		resolver,
		config.GetDomainLookupCredits(resolver.Code()),
		func(ctx context.Context) error {
			var lookupErr error

			domain, lookupErr = resolver.ResolveCompanyDomain(ctx, companyName)

			return lookupErr
		},
	)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		entry = enrichmentModel.NewCompanyDomain()
	}

	provider := resolver.Code()
	entry.NormalizedName, entry.CompanyName, entry.Provider = normalizedName, companyName, &provider
	entry.Domain = utils.NilIfEmpty(domain)
	entry.UpdateTimestamp()

	if err := entry.Validate(); err != nil {
		entry.Domain = nil
	}

	return s.companyDomainRepo.Upsert(ctx, entry)
}

func (s *EnrichmentService) domainResolver() providers.DomainResolver {
	for _, provider := range s.providers {
		if resolver, supported := provider.(providers.DomainResolver); supported {
			return resolver
		}
	}

	return nil
}
