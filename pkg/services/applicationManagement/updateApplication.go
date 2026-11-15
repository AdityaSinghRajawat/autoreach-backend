package applicationManagement

import (
	"context"
	"fmt"
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var allowedTransitionsI = map[modelHelpers.ApplicationStatus][]modelHelpers.ApplicationStatus{
	modelHelpers.ApplicationStatusSubmitted: {
		modelHelpers.ApplicationStatusActionRequired,
		modelHelpers.ApplicationStatusFailed,
	},
	modelHelpers.ApplicationStatusInterviewing: {modelHelpers.ApplicationStatusSubmitted},
	modelHelpers.ApplicationStatusRejected: {
		modelHelpers.ApplicationStatusSubmitted,
		modelHelpers.ApplicationStatusInterviewing,
	},
}

func (s *ApplicationService) UpdateApplication(
	ctx context.Context,
	candidateID, applicationID string,
	request applicationTypes.UpdateApplicationRequest,
) (*applicationModel.Application, *config.CustomError) {
	application, customErr := s.findApplication(ctx, candidateID, applicationID)
	if customErr != nil {
		return nil, customErr
	}

	target := modelHelpers.ApplicationStatus(request.Status)
	if !slices.Contains(allowedTransitionsI[target], application.Status) {
		return nil, config.NewCustomError(
			fmt.Errorf("update application: %w", config.ErrApplicationTransition),
			config.GetErrCodeConflict(),
		)
	}

	if target == modelHelpers.ApplicationStatusSubmitted {
		application.MarkSubmitted(nil)
		appendLog(application, "SUBMITTED", "submitted by the candidate")
	} else {
		application.Status = target
		appendLog(application, string(target), "updated by the candidate")
	}

	if customErr := s.saveApplication(ctx, application, "update application"); customErr != nil {
		return nil, customErr
	}

	return application, nil
}
