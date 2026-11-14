package matchingManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
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

	for _, similarity := range similarities {
		match := similarity.ToJobMatchModel(jobID)
		if match.MatchScore >= matchingTypes.StorageThreshold(similarity.MinMatchScore) {
			matches = append(matches, match)
		}
	}

	if customErr := s.saveMatches(ctx, matches, nil); customErr != nil {
		return nil, customErr
	}

	return &matchingTypes.MatchResult{Considered: len(similarities), Stored: len(matches)}, nil
}
