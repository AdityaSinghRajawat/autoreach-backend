package enrichmentManagement

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *EnrichmentService) EnrichJobContacts(
	ctx context.Context,
	jobID string,
	targetRoles []string,
) (*enrichmentTypes.EnrichJobContactsResult, *config.CustomError) {
	job, err := s.jobPostingRepo.FindByID(ctx, jobID)

	job, customErr := config.RequireFound("enrich job contacts", job, err, config.ErrJobNotFound)
	if customErr != nil {
		return nil, customErr
	}

	result := &enrichmentTypes.EnrichJobContactsResult{JobID: jobID}

	reused, customErr := s.reuseFreshContacts(ctx, job)
	if customErr != nil {
		return nil, customErr
	}

	if reused > 0 {
		result.Reused = reused

		s.publishOutreachPlans(ctx, jobID)

		return result, nil
	}

	roles := enrichmentTypes.ResolveTargetRoles(targetRoles, job)

	people, sourceCode, customErr := s.discoverPeople(ctx, job, roles)
	if customErr != nil {
		return nil, customErr
	}

	result.Discovered = len(people)
	selected := selectPeople(people, roles, sourceCode)
	contacts := s.resolveContacts(ctx, job, selected)

	for index, contact := range contacts {
		if contact == nil {
			continue
		}

		saved, customErr := s.saveContact(ctx, job, contact)
		if customErr != nil {
			return nil, customErr
		}

		if saved == nil {
			continue
		}

		result.Saved++

		if saved.WorkEmail == nil &&
			s.requestAsyncContact(ctx, job, saved, selected[index].Person) {
			result.PendingAsync++
		}
	}

	if result.Saved > 0 {
		s.publishOutreachPlans(ctx, jobID)
	}

	return result, nil
}

func (s *EnrichmentService) reuseFreshContacts(
	ctx context.Context,
	job *jobModel.JobPosting,
) (int, *config.CustomError) {
	since := utils.GetCurrTimeStamp().Add(-config.GetEnrichmentContactFreshness())

	fresh, err := s.companyContactRepo.ListFreshByCompany(ctx, job.CompanyDomain, since)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("enrich job contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	contactIDs := make([]string, 0, config.GetEnrichmentMaxContactsPerJob())

	for _, contact := range fresh {
		if contact.IsReachable() && len(contactIDs) < config.GetEnrichmentMaxContactsPerJob() {
			contactIDs = append(contactIDs, contact.ID)
		}
	}

	if err := s.companyContactRepo.LinkToJob(ctx, job.ID, contactIDs); err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("enrich job contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return len(contactIDs), nil
}

func (s *EnrichmentService) discoverPeople(
	ctx context.Context,
	job *jobModel.JobPosting,
	roles []string,
) ([]enrichmentTypes.Person, string, *config.CustomError) {
	query := enrichmentTypes.DiscoveryQuery{
		CompanyDomain: job.CompanyDomain,
		Titles:        config.GetSearchTitles(roles),
		Limit:         config.GetEnrichmentPeoplePerSearch(),
	}

	var lastErr error

	for _, provider := range s.providers {
		discoverer, capable := provider.(providers.Discoverer)
		if !capable {
			continue
		}

		var people []enrichmentTypes.Person

		err := s.callProvider(ctx, provider, 0, func(ctx context.Context) error {
			var err error

			people, err = discoverer.DiscoverPeople(ctx, query)

			return err
		})
		if err == nil {
			return people, provider.Code(), nil
		}

		lastErr = err

		utils.LogWarn(
			ctx,
			"people discovery failed",
			"provider",
			provider.Code(),
			"error",
			err.Error(),
		)
	}

	if lastErr == nil {
		utils.LogWarn(ctx, config.ErrNoDiscoveryProvider.Error(), "job_id", job.ID)

		return nil, "", nil
	}

	errCode := config.GetErrCodeServiceUnavailable()
	if !isProviderSkipped(lastErr) {
		errCode = utils.ClassifyUpstreamError(lastErr)
	}

	return nil, "", config.NewCustomError(fmt.Errorf("discover people: %w", lastErr), errCode)
}

func selectPeople(
	people []enrichmentTypes.Person,
	roles []string,
	sourceCode string,
) []enrichmentTypes.SelectedPerson {
	limit := config.GetEnrichmentMaxContactsPerJob()
	byRole := make(map[string][]enrichmentTypes.Person, len(roles))

	for _, person := range people {
		role, classified := config.ClassifyContactRole(person.Title)
		if classified && slices.Contains(roles, role) {
			byRole[role] = append(byRole[role], person)
		}
	}

	selected := make([]enrichmentTypes.SelectedPerson, 0, limit)

	for round := 0; len(selected) < limit; round++ {
		added := false

		for _, role := range roles {
			if round < len(byRole[role]) && len(selected) < limit {
				selected = append(selected, enrichmentTypes.SelectedPerson{
					Person:     byRole[role][round],
					Role:       role,
					SourceCode: sourceCode,
				})
				added = true
			}
		}

		if !added {
			break
		}
	}

	return selected
}

func (s *EnrichmentService) resolveContacts(
	ctx context.Context,
	job *jobModel.JobPosting,
	selected []enrichmentTypes.SelectedPerson,
) []*enrichmentModel.CompanyContact {
	contacts := make([]*enrichmentModel.CompanyContact, len(selected))

	tasks := make([]utils.ConcurrentTask, 0, len(selected))
	for index := range selected {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "resolve contact",
			Run: func(ctx context.Context) error {
				contacts[index] = s.resolveContact(ctx, job, selected[index])

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(ctx, tasks...); err != nil {
		utils.LogError(ctx, err, "job_id", job.ID)
	}

	return contacts
}

func (s *EnrichmentService) resolveContact(
	ctx context.Context,
	job *jobModel.JobPosting,
	selected enrichmentTypes.SelectedPerson,
) *enrichmentModel.CompanyContact {
	details, providerCode := s.findContactDetails(ctx, selected.Person)
	if providerCode == "" {
		providerCode = selected.SourceCode
	}

	emailStatus := modelHelpers.EmailStatusUnverified
	if details != nil && details.Email != "" {
		emailStatus = s.emailVerifier.Verify(
			ctx,
			modelHelpers.NormalizeEmail(details.Email),
			details.EmailVerified,
		)
	}

	contact := selected.Person.ToCompanyContactModel(
		job,
		selected.Role,
		providerCode,
		details,
		emailStatus,
	)
	if providerCode != selected.SourceCode {
		contact.ProviderPersonID = nil
	}

	if contact.WorkEmail == nil && contact.LinkedinURL == nil {
		return nil
	}

	return contact
}

func (s *EnrichmentService) findContactDetails(
	ctx context.Context,
	person enrichmentTypes.Person,
) (*enrichmentTypes.ContactDetails, string) {
	var (
		fallback     *enrichmentTypes.ContactDetails
		fallbackCode string
	)

	for _, provider := range s.providers {
		finder, capable := provider.(providers.ContactFinder)
		if !capable {
			continue
		}

		var details *enrichmentTypes.ContactDetails

		err := s.callProvider(
			ctx,
			provider,
			config.GetCreditCostPerLookup(),
			func(ctx context.Context) error {
				var err error

				details, err = finder.FindContact(ctx, person.Merge(fallback))

				return err
			},
		)
		if err != nil {
			utils.LogWarn(
				ctx,
				"contact lookup failed",
				"provider",
				provider.Code(),
				"error",
				err.Error(),
			)

			continue
		}

		if details == nil {
			continue
		}

		if details.Email != "" {
			return details, provider.Code()
		}

		if fallback == nil {
			fallback, fallbackCode = details, provider.Code()
		}
	}

	return fallback, fallbackCode
}

func (s *EnrichmentService) saveContact(
	ctx context.Context,
	job *jobModel.JobPosting,
	contact *enrichmentModel.CompanyContact,
) (*enrichmentModel.CompanyContact, *config.CustomError) {
	existing, err := s.companyContactRepo.FindExisting(ctx, contact)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("save contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if existing != nil {
		contact.AdoptIdentity(existing, utils.GetCurrTimeStamp())
	}

	contact.UpdateTimestamp()

	if err := contact.Validate(); err != nil {
		utils.LogWarn(ctx, "discovered contact rejected", "job_id", job.ID, "error", err.Error())

		return nil, nil
	}

	var saved *enrichmentModel.CompanyContact

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		saved, err = s.companyContactRepo.Upsert(ctx, contact)
		if err != nil {
			return err
		}

		return s.companyContactRepo.LinkToJob(ctx, job.ID, []string{saved.ID})
	})
	if errors.Is(err, config.ErrDuplicateRecord) {
		utils.LogWarn(ctx, "contact already stored by a concurrent run", "job_id", job.ID)

		return nil, nil
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("save contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return saved, nil
}

func (s *EnrichmentService) requestAsyncContact(
	ctx context.Context,
	job *jobModel.JobPosting,
	contact *enrichmentModel.CompanyContact,
	person enrichmentTypes.Person,
) bool {
	if contact.LinkedinURL == nil {
		return false
	}

	person.LinkedinURL = *contact.LinkedinURL

	for _, provider := range s.providers {
		finder, capable := provider.(providers.AsyncContactFinder)
		if !capable {
			continue
		}

		var externalRequestID string

		err := s.callProvider(
			ctx,
			provider,
			config.GetCreditCostPerLookup(),
			func(ctx context.Context) error {
				var err error

				externalRequestID, err = finder.RequestContact(ctx, person)

				return err
			},
		)
		if err != nil {
			utils.LogWarn(
				ctx,
				"async contact request failed",
				"provider",
				provider.Code(),
				"error",
				err.Error(),
			)

			continue
		}

		request := enrichmentModel.NewEnrichmentRequest()
		request.JobID = job.ID
		request.ContactID = contact.ID
		request.Provider = provider.Code()
		request.ExternalRequestID = externalRequestID
		request.CreditsUsed = config.GetCreditCostPerLookup()

		if err := request.Validate(); err != nil {
			utils.LogError(ctx, err, "provider", provider.Code())

			return false
		}

		if _, err := s.enrichmentRequestRepo.Upsert(ctx, request); err != nil {
			utils.LogError(ctx, err, "provider", provider.Code())

			return false
		}

		return true
	}

	return false
}
