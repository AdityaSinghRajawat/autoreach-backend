package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *EnrichmentService) ListJobContacts(
	ctx context.Context,
	candidateID, jobID string,
) ([]*enrichmentModel.CompanyContact, *config.CustomError) {
	if customErr := s.requireMatchedJob(ctx, candidateID, jobID); customErr != nil {
		return nil, customErr
	}

	contacts, err := s.companyContactRepo.ListByJob(ctx, jobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list job contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return contacts, nil
}

func (s *EnrichmentService) RequestJobEnrichment(
	ctx context.Context,
	candidateID, jobID string,
) (*enrichmentTypes.RequestEnrichmentResponse, *config.CustomError) {
	if customErr := s.requireMatchedJob(ctx, candidateID, jobID); customErr != nil {
		return nil, customErr
	}

	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskEnrichContacts,
		Payload: workerCustomTypes.EnrichContactsPayload{JobID: jobID},
	})
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("request job enrichment: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	return enrichmentTypes.NewRequestEnrichmentResponse(
		jobID,
		string(workerConfigs.TaskEnrichContacts),
		workerConfigs.PublishStatusEnqueued,
	), nil
}

func (s *EnrichmentService) ListCandidateContacts(
	ctx context.Context,
	candidateID string,
	params enrichmentTypes.ListContactsParams,
) ([]*enrichmentModel.CompanyContact, *sharedTypes.PaginationMeta, *config.CustomError) {
	page := params.Normalize()

	var (
		contacts []*enrichmentModel.CompanyContact
		total    int
	)

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "list candidate contacts",
			Run: func(ctx context.Context) error {
				var err error

				contacts, err = s.companyContactRepo.ListByCandidate(ctx, candidateID, page)

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "count candidate contacts",
			Run: func(ctx context.Context) error {
				var err error

				total, err = s.companyContactRepo.CountByCandidate(ctx, candidateID)

				return err
			},
		},
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list candidate contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return contacts, &meta, nil
}

func (s *EnrichmentService) requireMatchedJob(
	ctx context.Context,
	candidateID, jobID string,
) *config.CustomError {
	if _, err := uuid.Parse(jobID); err != nil {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", config.ErrJobNotMatched),
			config.GetErrCodeNotFound(),
		)
	}

	matched, err := s.jobMatchRepo.Exists(ctx, candidateID, jobID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !matched {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", config.ErrJobNotMatched),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
