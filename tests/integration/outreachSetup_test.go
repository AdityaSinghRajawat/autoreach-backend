//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

func templatesPath(candidateID string) string {
	return candidatePath(candidateID) + "/outreach-templates"
}

func campaignsPath(candidateID string) string {
	return candidatePath(candidateID) + "/outreach-campaigns"
}

func (f jobFixture) createTemplate(t *testing.T, subject, body string) outreachModel.Template {
	t.Helper()

	created := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		templatesPath(f.candidateID),
		map[string]any{
			"name": "Intro", "channel": "EMAIL", "subject": subject, "body": body,
		},
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	return helpers.DecodeData[outreachModel.Template](t, created)
}

func (f jobFixture) approveTemplate(t *testing.T, templateID string) outreachModel.Template {
	t.Helper()

	approved := helpers.Do(
		t,
		f.router,
		http.MethodPatch,
		templatesPath(f.candidateID)+"/"+templateID,
		map[string]any{"status": "APPROVED"},
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, approved.Recorder.Code, string(approved.Body))
	helpers.AssertMatchesSpec(t, approved)

	return helpers.DecodeData[outreachModel.Template](t, approved)
}

func (f jobFixture) createCampaign(
	t *testing.T,
	body map[string]any,
	expectedStatus int,
) helpers.Exchange {
	t.Helper()

	created := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		campaignsPath(f.candidateID),
		body,
		bearer(f.userToken),
	)
	require.Equal(t, expectedStatus, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	return created
}

func TestTemplatesAreValidatedApprovedAndSentBackToDraftWhenReworded(t *testing.T) {
	fixture := newJobFixture(t)

	rejected := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		templatesPath(fixture.candidateID),
		map[string]any{
			"name":    "Bad",
			"channel": "EMAIL",
			"body":    "Hi {{contactFirstName}}, I am {{favouriteColour}}",
		},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusUnprocessableEntity, rejected.Recorder.Code)
	helpers.AssertMatchesSpec(t, rejected)
	assert.ElementsMatch(t, []string{"subject", "body"}, detailFields(t, rejected),
		"an email needs a subject and only known placeholders are allowed")

	template := fixture.createTemplate(t, "{{jobTitle}} at {{companyName}}",
		"Hi {{contactFirstName}}, I work with {{matchingSkills}}.")
	assert.Equal(t, modelHelpers.TemplateStatusDraft, template.Status)
	assert.Nil(t, template.ApprovedAt)

	approved := fixture.approveTemplate(t, template.ID)
	assert.Equal(t, modelHelpers.TemplateStatusApproved, approved.Status)
	assert.NotNil(t, approved.ApprovedAt)

	renamed := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		templatesPath(fixture.candidateID)+"/"+template.ID,
		map[string]any{"name": "Renamed"},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, renamed.Recorder.Code)
	assert.Equal(
		t,
		modelHelpers.TemplateStatusApproved,
		helpers.DecodeData[outreachModel.Template](
			t,
			renamed,
		).Status,
		"a rename keeps the approval",
	)

	reworded := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		templatesPath(fixture.candidateID)+"/"+template.ID,
		map[string]any{"body": "Hello {{contactName}}"},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, reworded.Recorder.Code)
	assert.Equal(
		t,
		modelHelpers.TemplateStatusDraft,
		helpers.DecodeData[outreachModel.Template](
			t,
			reworded,
		).Status,
		"new wording needs a new approval",
	)

	listed := helpers.Do(t, fixture.router, http.MethodGet, templatesPath(fixture.candidateID), nil,
		bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
	assert.Len(t, helpers.DecodeData[[]outreachModel.Template](t, listed), 1)
}

func TestCampaignsUseOwnEmailTemplatesAndOnlyOneIsActive(t *testing.T) {
	fixture := newJobFixture(t)
	first := fixture.createTemplate(t, "Hello", "Hi {{contactFirstName}}")
	followUp := fixture.createTemplate(t, "Re: Hello", "Following up, {{contactFirstName}}")

	created := fixture.createCampaign(t, map[string]any{
		"name": "Default",
		"steps": []map[string]any{
			{"templateId": first.ID, "delayDays": 0},
			{"templateId": followUp.ID, "delayDays": 3},
		},
	}, http.StatusCreated)
	campaign := helpers.DecodeData[outreachTypes.CampaignDetail](t, created)
	assert.Equal(t, modelHelpers.CampaignStatusActive, campaign.Status)
	require.Len(t, campaign.Steps, 2)
	assert.Equal(t, 2, campaign.Steps[1].StepNumber)
	assert.Equal(t, 3, campaign.Steps[1].DelayDays)

	fixture.createCampaign(t, map[string]any{
		"name": "Second", "steps": []map[string]any{{"templateId": first.ID, "delayDays": 0}},
	}, http.StatusConflict)

	other := register(t, fixture.router, "other@example.com")
	foreign := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		templatesPath(*other.User.CandidateID),
		map[string]any{
			"name": "Theirs", "channel": "EMAIL", "subject": "Hi", "body": "Hi",
		},
		bearer(other.AccessToken),
	)
	require.Equal(t, http.StatusCreated, foreign.Recorder.Code)

	fixture.createCampaign(t, map[string]any{
		"name":   "Borrowed",
		"status": "PAUSED",
		"steps": []map[string]any{
			{
				"templateId": helpers.DecodeData[outreachModel.Template](t, foreign).ID,
				"delayDays":  0,
			},
		},
	}, http.StatusUnprocessableEntity)

	inUse := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		templatesPath(fixture.candidateID)+"/"+first.ID,
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(
		t,
		http.StatusConflict,
		inUse.Recorder.Code,
		"a template used by a campaign cannot be deleted",
	)

	paused := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		campaignsPath(fixture.candidateID)+"/"+campaign.ID,
		map[string]any{
			"status": "PAUSED",
			"steps":  []map[string]any{{"templateId": followUp.ID, "delayDays": 0}},
		},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, paused.Recorder.Code, string(paused.Body))
	helpers.AssertMatchesSpec(t, paused)
	updated := helpers.DecodeData[outreachTypes.CampaignDetail](t, paused)
	assert.Equal(t, modelHelpers.CampaignStatusPaused, updated.Status)
	assert.Len(t, updated.Steps, 1, "supplying steps replaces them")

	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		templatesPath(fixture.candidateID)+"/"+first.ID,
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusNoContent, deleted.Recorder.Code)

	listed := helpers.Do(t, fixture.router, http.MethodGet, campaignsPath(fixture.candidateID), nil,
		bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
}

func detailFields(t *testing.T, exchange helpers.Exchange) []string {
	t.Helper()

	fields := []string{}
	for _, detail := range helpers.DecodeEnvelope(t, exchange).Error.Details {
		fields = append(fields, detail.Field)
	}

	return fields
}
