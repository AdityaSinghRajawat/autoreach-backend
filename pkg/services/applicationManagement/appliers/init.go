package appliers

import (
	"context"

	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	greenhouseApplier "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers/greenhouse"
	leverApplier "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers/lever"
)

type Applier interface {
	Code() string
	Supports(job *jobModel.JobPosting) bool
	FetchForm(
		ctx context.Context,
		job *jobModel.JobPosting,
	) (applicationTypes.ApplicationForm, error)
	CanSubmit(job *jobModel.JobPosting) bool
	Submit(
		ctx context.Context,
		job *jobModel.JobPosting,
		submission applicationTypes.Submission,
	) (string, error)
}

type RegistryOptions struct {
	GreenhouseAPIURL string
	LeverAPIURL      string
	LeverEUAPIURL    string
	GreenhouseKeys   map[string]string
	LeverKeys        map[string]string
}

func NewApplierRegistry(options RegistryOptions) []Applier {
	return []Applier{
		greenhouseApplier.NewGreenhouseApplier(
			greenhouseIntegration.NewGreenhouseIntegration(options.GreenhouseAPIURL),
			options.GreenhouseKeys,
		),
		leverApplier.NewLeverApplier(
			leverIntegration.NewLeverIntegration(options.LeverAPIURL),
			leverIntegration.NewLeverIntegration(options.LeverEUAPIURL),
			options.LeverKeys,
		),
	}
}

func Find(registry []Applier, job *jobModel.JobPosting) Applier {
	for _, applier := range registry {
		if applier.Supports(job) {
			return applier
		}
	}

	return nil
}
