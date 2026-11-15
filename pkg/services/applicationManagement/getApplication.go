package applicationManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
)

func (s *ApplicationService) GetApplication(
	ctx context.Context,
	candidateID, applicationID string,
) (*applicationTypes.ApplicationDetail, *config.CustomError) {
	application, customErr := s.findApplication(ctx, candidateID, applicationID)
	if customErr != nil {
		return nil, customErr
	}

	details, customErr := s.withJobs(ctx, []*applicationModel.Application{application})
	if customErr != nil {
		return nil, customErr
	}

	return details[0], nil
}

func (s *ApplicationService) findApplication(
	ctx context.Context,
	candidateID, applicationID string,
) (*applicationModel.Application, *config.CustomError) {
	application, err := s.applicationRepo.FindByCandidate(ctx, candidateID, applicationID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if application == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get application: %w", config.ErrApplicationNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return application, nil
}
