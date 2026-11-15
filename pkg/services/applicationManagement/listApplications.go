package applicationManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
)

func (s *ApplicationService) ListApplications(
	ctx context.Context,
	candidateID string,
	params applicationTypes.ListApplicationsParams,
) ([]*applicationTypes.ApplicationDetail, *sharedTypes.PaginationMeta, *config.CustomError) {
	page := params.Normalize()

	applications, total, err := s.applicationRepo.ListByCandidate(
		ctx,
		candidateID,
		params.Status,
		page,
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list applications: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	details, customErr := s.withJobs(ctx, applications)
	if customErr != nil {
		return nil, nil, customErr
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return details, &meta, nil
}

func (s *ApplicationService) withJobs(
	ctx context.Context,
	applications []*applicationModel.Application,
) ([]*applicationTypes.ApplicationDetail, *config.CustomError) {
	jobs, err := s.jobPostingRepo.ListByIDs(ctx, applicationTypes.JobIDs(applications))
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("load application jobs: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return applicationTypes.NewApplicationDetails(applications, jobs), nil
}
