package candidateManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) UpdateCandidate(
	ctx context.Context,
	candidateID string,
	updateCandidateRequest candidateTypes.UpdateCandidateRequest,
) (*candidateModel.Candidate, *config.CustomError) {
	if updateCandidateRequest.IsEmpty() {
		return nil, config.NewCustomError(
			fmt.Errorf("update candidate: %w", config.ErrNoFieldsSupplied),
			config.GetErrCodeBadRequest(),
		)
	}

	existing, customErr := s.GetCandidate(ctx, candidateID)
	if customErr != nil {
		return nil, customErr
	}

	candidate := updateCandidateRequest.ToCandidateModel(existing)
	candidate.UpdateTimestamp()

	if err := candidate.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("update candidate: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	updated, err := s.candidateRepo.Upsert(ctx, candidate)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("update candidate: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("update candidate: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return updated, nil
}
