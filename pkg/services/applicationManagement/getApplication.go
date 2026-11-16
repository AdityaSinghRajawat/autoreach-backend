package applicationManagement

import (
	"context"

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

	return config.RequireFound("get application", application, err, config.ErrApplicationNotFound)
}
