package lever

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	leverIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ats/lever"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
)

type LeverApplier struct {
	global   *leverIntegration.LeverIntegration
	eu       *leverIntegration.LeverIntegration
	siteKeys map[string]string
}

func NewLeverApplier(
	global, eu *leverIntegration.LeverIntegration,
	siteKeys map[string]string,
) *LeverApplier {
	return &LeverApplier{global: global, eu: eu, siteKeys: siteKeys}
}

func (a *LeverApplier) Code() string {
	return string(modelHelpers.ATSTypeLever)
}

func (a *LeverApplier) Supports(job *jobModel.JobPosting) bool {
	_, _, _, found := sitePosting(job.ExternalURL)

	return found
}

// Lever's public API hides custom questions, so the form is the standard set, flagged incomplete.
func (a *LeverApplier) FetchForm(
	ctx context.Context,
	job *jobModel.JobPosting,
) (applicationTypes.ApplicationForm, error) {
	site, postingID, eu, _ := sitePosting(job.ExternalURL)

	if _, err := a.integrationFor(eu).GetPosting(ctx, site, postingID); err != nil {
		return applicationTypes.ApplicationForm{}, err
	}

	return applicationTypes.ApplicationForm{
		HasHiddenQuestions: true,
		Questions: []applicationTypes.FormQuestion{
			textQuestion("Full name", "name", true),
			textQuestion("Email", "email", true),
			{
				Label:    "Resume",
				Required: true,
				Fields:   []applicationTypes.FormField{{Name: "resume", Type: "input_file"}},
			},
			textQuestion("Phone", "phone", false),
			textQuestion("Current company", "org", false),
			textQuestion("LinkedIn URL", config.GetLeverLinkedinURLField(), false),
			textQuestion("GitHub URL", config.GetLeverGithubURLField(), false),
			textQuestion("Portfolio URL", config.GetLeverPortfolioURLField(), false),
		},
	}, nil
}

func (a *LeverApplier) CanSubmit(job *jobModel.JobPosting) bool {
	site, _, _, _ := sitePosting(job.ExternalURL)

	return a.siteKeys[site] != ""
}

func (a *LeverApplier) Submit(
	ctx context.Context,
	job *jobModel.JobPosting,
	submission applicationTypes.Submission,
) (string, error) {
	site, postingID, eu, _ := sitePosting(job.ExternalURL)

	apiKey := a.siteKeys[site]
	if apiKey == "" {
		return "", fmt.Errorf("apply to lever posting: %w", config.ErrMissingCredential)
	}

	return a.integrationFor(eu).ApplyToPosting(ctx, site, postingID, apiKey, submission)
}

func (a *LeverApplier) integrationFor(eu bool) *leverIntegration.LeverIntegration {
	if eu {
		return a.eu
	}

	return a.global
}

func textQuestion(label, name string, required bool) applicationTypes.FormQuestion {
	return applicationTypes.FormQuestion{
		Label:    label,
		Required: required,
		Fields:   []applicationTypes.FormField{{Name: name, Type: "input_text"}},
	}
}

// Public job pages look like https://jobs.lever.co/{site}/{postingId} (jobs.eu.lever.co in the EU).
func sitePosting(rawURL string) (string, string, bool, bool) {
	parsed, err := url.Parse(rawURL)
	host := ""

	if err == nil {
		host = strings.ToLower(parsed.Hostname())
	}

	if err != nil || !strings.HasSuffix(host, config.GetLeverHostSuffix()) {
		return "", "", false, false
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 2 || segments[0] == "" || segments[1] == "" ||
		segments[1] == config.GetLeverApplySegment() {
		return "", "", false, false
	}

	return strings.ToLower(
			segments[0],
		), segments[1], strings.Contains(
			host,
			config.GetLeverEUHostMarker(),
		), true
}
