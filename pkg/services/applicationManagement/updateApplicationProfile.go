package applicationManagement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
)

func (s *ApplicationService) UpdateApplicationProfile(
	ctx context.Context,
	candidateID string,
	profile applicationTypes.ApplicationProfile,
) (*applicationTypes.ApplicationProfile, *config.CustomError) {
	candidate, customErr := s.findCandidate(ctx, candidateID)
	if customErr != nil {
		return nil, customErr
	}

	settings := map[string]json.RawMessage{}

	encoded, err := json.Marshal(profile)
	if err == nil {
		err = json.Unmarshal(candidate.Settings, &settings)
	}

	if err == nil {
		settings[config.GetApplicationProfileKey()] = encoded
		candidate.Settings, err = json.Marshal(settings)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("update application profile: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	candidate.UpdateTimestamp()

	if err := candidate.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("update application profile: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	if _, err := s.candidateRepo.Upsert(ctx, candidate); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("update application profile: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return &profile, nil
}
