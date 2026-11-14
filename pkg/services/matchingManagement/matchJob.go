package matchingManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *MatchingService) MatchJob(
	ctx context.Context,
	jobID string,
) (*matchingTypes.MatchResult, *config.CustomError) {
	similarities, err := s.resumeRepo.ListActiveSimilaritiesToJob(ctx, jobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match job: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	matches := make([]*matchingModel.JobMatch, 0, len(similarities))
	matched := make([]matchingTypes.ResumeSimilarity, 0, len(similarities))

	for _, similarity := range similarities {
		match := similarity.ToJobMatchModel(jobID)
		if match.MatchScore >= matchingTypes.StorageThreshold(similarity.MinMatchScore) {
			matches = append(matches, match)
			matched = append(matched, similarity)
		}
	}

	if customErr := s.saveMatches(ctx, matches, nil); customErr != nil {
		return nil, customErr
	}

	tasks := make([]utils.ConcurrentTask, 0, len(matched))
	for _, similarity := range matched {
		tasks = append(tasks, utils.ConcurrentTask{
			Name: "fan out matches",
			Run: func(ctx context.Context) error {
				s.fanOutMatches(
					ctx,
					similarity.CandidateID,
					similarity.MinMatchScore,
					similarity.AutoOutreachEnabled,
				)

				return nil
			},
		})
	}

	if err := utils.RunConcurrently(ctx, tasks...); err != nil {
		utils.LogError(ctx, err, "job_id", jobID)
	}

	return &matchingTypes.MatchResult{Considered: len(similarities), Stored: len(matches)}, nil
}
