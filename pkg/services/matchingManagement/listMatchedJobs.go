package matchingManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *MatchingService) ListMatchedJobs(
	ctx context.Context,
	candidateID string,
	params matchingTypes.ListMatchedJobsParams,
) ([]matchingTypes.MatchedJob, *sharedTypes.PaginationMeta, *config.CustomError) {
	minScore, customErr := s.resolveMinScore(ctx, candidateID, params.MinScore)
	if customErr != nil {
		return nil, nil, customErr
	}

	page := params.Normalize()

	var (
		jobs  []matchingTypes.MatchedJob
		total int
	)

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "list matched jobs",
			Run: func(ctx context.Context) error {
				var err error

				jobs, err = s.jobMatchRepo.ListForCandidate(ctx, candidateID, minScore, page)

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "count matched jobs",
			Run: func(ctx context.Context) error {
				var err error

				total, err = s.jobMatchRepo.CountForCandidate(ctx, candidateID, minScore)

				return err
			},
		},
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list matched jobs: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return jobs, &meta, nil
}

func (s *MatchingService) resolveMinScore(
	ctx context.Context,
	candidateID string,
	requested *float64,
) (float64, *config.CustomError) {
	if requested != nil {
		return *requested, nil
	}

	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)

	candidate, customErr := config.RequireFound(
		"list matched jobs", candidate, err, config.ErrCandidateNotFound,
	)
	if customErr != nil {
		return 0, customErr
	}

	return candidate.MinMatchScore, nil
}
