package greenhouse

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	greenhouseIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/greenhouse"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type GreenhouseApplier struct {
	integration *greenhouseIntegration.GreenhouseIntegration
	boardKeys   map[string]string
}

func NewGreenhouseApplier(
	integration *greenhouseIntegration.GreenhouseIntegration,
	boardKeys map[string]string,
) *GreenhouseApplier {
	return &GreenhouseApplier{integration: integration, boardKeys: boardKeys}
}

func (a *GreenhouseApplier) Code() string {
	return string(modelHelpers.ATSTypeGreenhouse)
}

func (a *GreenhouseApplier) Supports(job *jobModel.JobPosting) bool {
	_, _, found := boardAndJob(job.ExternalURL)

	return found
}

func (a *GreenhouseApplier) FetchForm(
	ctx context.Context,
	job *jobModel.JobPosting,
) (applicationTypes.ApplicationForm, error) {
	boardToken, jobID, _ := boardAndJob(job.ExternalURL)

	return a.integration.GetJobForm(ctx, boardToken, jobID)
}

func (a *GreenhouseApplier) CanSubmit(job *jobModel.JobPosting) bool {
	boardToken, _, _ := boardAndJob(job.ExternalURL)

	return a.boardKeys[boardToken] != ""
}

func (a *GreenhouseApplier) Submit(
	ctx context.Context,
	job *jobModel.JobPosting,
	submission applicationTypes.Submission,
) (string, error) {
	boardToken, jobID, _ := boardAndJob(job.ExternalURL)

	apiKey := a.boardKeys[boardToken]
	if apiKey == "" {
		return "", fmt.Errorf("submit greenhouse application: %w", config.ErrMissingCredential)
	}

	return "", a.integration.SubmitApplication(ctx, boardToken, jobID, apiKey, submission)
}

// Public job pages look like https://boards.greenhouse.io/{board}/jobs/{id}.
func boardAndJob(rawURL string) (string, string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil ||
		!strings.HasSuffix(strings.ToLower(parsed.Hostname()), config.GetGreenhouseHostSuffix()) {
		return "", "", false
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 3 || segments[1] != config.GetGreenhouseJobsSegment() || segments[0] == "" ||
		segments[2] == "" {
		return "", "", false
	}

	return strings.ToLower(segments[0]), segments[2], true
}
