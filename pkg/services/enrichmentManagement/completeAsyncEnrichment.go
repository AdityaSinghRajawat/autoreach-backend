package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
)

func (s *EnrichmentService) CompleteAsyncEnrichment(
	ctx context.Context,
	providerCode, externalRequestID string,
	payload []byte,
) *config.CustomError {
	finder := s.asyncFinder(providerCode)
	if finder == nil || externalRequestID == "" {
		return nil
	}

	request, err := s.enrichmentRequestRepo.FindPendingByExternalID(
		ctx,
		providerCode,
		externalRequestID,
	)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("complete async enrichment: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if request == nil {
		return nil
	}

	results, err := finder.ParseCallback(payload)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("complete async enrichment: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	contact, err := s.companyContactRepo.FindByID(ctx, request.ContactID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("complete async enrichment: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	details := firstUsableResult(results)
	if contact == nil || details == nil {
		request.MarkFailed(config.GetReasonNoContactReturned())

		return s.saveAsyncOutcome(ctx, request, nil)
	}

	email := modelHelpers.NormalizeEmail(details.Email)
	contact.SetEmail(email, s.emailVerifier.Verify(ctx, email, details.EmailVerified))
	contact.EnrichmentProvider = providerCode
	contact.ProviderPersonID = nil

	if details.Phone != "" {
		contact.PhoneNumber = &details.Phone
	}

	contact.UpdateTimestamp()

	if err := contact.Validate(); err != nil {
		request.MarkFailed(err.Error())

		return s.saveAsyncOutcome(ctx, request, nil)
	}

	request.MarkCompleted()

	return s.saveAsyncOutcome(ctx, request, contact)
}

func (s *EnrichmentService) saveAsyncOutcome(
	ctx context.Context,
	request *enrichmentModel.EnrichmentRequest,
	contact *enrichmentModel.CompanyContact,
) *config.CustomError {
	request.UpdateTimestamp()

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if contact != nil {
			if _, err := s.companyContactRepo.Upsert(ctx, contact); err != nil {
				return err
			}
		}

		_, err := s.enrichmentRequestRepo.Upsert(ctx, request)

		return err
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("complete async enrichment: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func (s *EnrichmentService) asyncFinder(providerCode string) providers.AsyncContactFinder {
	for _, provider := range s.providers {
		if finder, capable := provider.(providers.AsyncContactFinder); capable &&
			provider.Code() == providerCode {
			return finder
		}
	}

	return nil
}

func firstUsableResult(results []enrichmentTypes.CallbackResult) *enrichmentTypes.ContactDetails {
	for index := range results {
		if results[index].Found && results[index].Details.Email != "" {
			return &results[index].Details
		}
	}

	return nil
}
