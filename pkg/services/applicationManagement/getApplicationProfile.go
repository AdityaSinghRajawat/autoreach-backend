package applicationManagement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *ApplicationService) GetApplicationProfile(
	ctx context.Context,
	candidateID string,
) (*applicationTypes.ApplicationProfile, *config.CustomError) {
	candidate, customErr := s.findCandidate(ctx, candidateID)
	if customErr != nil {
		return nil, customErr
	}

	profile, err := applicationProfileOf(candidate)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get application profile: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return &profile, nil
}

func (s *ApplicationService) findCandidate(
	ctx context.Context,
	candidateID string,
) (*candidateModel.Candidate, *config.CustomError) {
	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("find candidate: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if candidate == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("find candidate: %w", config.ErrCandidateNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return candidate, nil
}

func applicationProfileOf(
	candidate *candidateModel.Candidate,
) (applicationTypes.ApplicationProfile, error) {
	settings := map[string]json.RawMessage{}
	if err := json.Unmarshal(candidate.Settings, &settings); err != nil {
		return applicationTypes.ApplicationProfile{}, err
	}

	profile := applicationTypes.ApplicationProfile{}

	raw, found := settings[config.GetApplicationProfileKey()]
	if !found {
		return profile, nil
	}

	if err := json.Unmarshal(raw, &profile); err != nil {
		return applicationTypes.ApplicationProfile{}, err
	}

	return profile, nil
}
