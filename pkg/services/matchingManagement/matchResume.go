package matchingManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
)

func (s *MatchingService) MatchResume(
	ctx context.Context,
	resumeID string,
) (*matchingTypes.MatchResult, *config.CustomError) {
	embedding, err := s.resumeRepo.FindEmbedding(ctx, resumeID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if embedding == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match resume: %w", config.ErrResumeNotReady),
			config.GetErrCodeNotFound(),
		)
	}

	if !embedding.IsActive {
		return &matchingTypes.MatchResult{}, nil
	}

	candidate, err := s.candidateRepo.FindByID(ctx, embedding.CandidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if candidate == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match resume: %w", config.ErrCandidateNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	var similarities []matchingTypes.JobSimilarity

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		similarities, err = s.jobPostingRepo.ListNearestActive(
			ctx,
			embedding.Vector,
			config.GetMatchTopK(),
			config.GetMatchEFSearch(),
		)

		return err
	})
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("match resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	threshold := matchingTypes.StorageThreshold(candidate.MinMatchScore)
	matches := make([]*matchingModel.JobMatch, 0, len(similarities))

	for _, similarity := range similarities {
		match := similarity.ToJobMatchModel(embedding.CandidateID, embedding.ResumeID)
		if match.MatchScore >= threshold {
			matches = append(matches, match)
		}
	}

	if customErr := s.saveMatches(ctx, matches, func(ctx context.Context) error {
		return s.jobMatchRepo.DeleteStaleForCandidate(
			ctx,
			embedding.CandidateID,
			embedding.ResumeID,
		)
	}); customErr != nil {
		return nil, customErr
	}

	return &matchingTypes.MatchResult{Considered: len(similarities), Stored: len(matches)}, nil
}

func (s *MatchingService) saveMatches(
	ctx context.Context,
	matches []*matchingModel.JobMatch,
	afterUpsert func(ctx context.Context) error,
) *config.CustomError {
	for _, match := range matches {
		match.UpdateTimestamp()

		if err := match.Validate(); err != nil {
			return config.NewCustomError(
				fmt.Errorf("save matches: %w", err),
				config.GetErrCodeInternalServerError(),
			)
		}
	}

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.jobMatchRepo.UpsertMany(ctx, matches); err != nil {
			return err
		}

		if afterUpsert == nil {
			return nil
		}

		return afterUpsert(ctx)
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("save matches: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
