//go:build integration

package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

type linkedinFixture struct {
	jobFixture
	jobID     string
	contactID string
}

func newLinkedinFixture(t *testing.T) linkedinFixture {
	t.Helper()

	fixture := newJobFixture(t)
	fixture.enableAutomation(t)
	jobID := fixture.ingest(t, withFields(validJob(), map[string]any{"companyDomain": "acme.test"})).
		QueuedTasks[0].JobID

	fixture.seedMatch(t, jobID)

	return linkedinFixture{
		jobFixture: fixture,
		jobID:      jobID,
		contactID:  fixture.seedLinkedinContact(t, jobID, "Sarah Lee", "sarah-lee"),
	}
}

func (f jobFixture) seedLinkedinContact(t *testing.T, jobID, fullName, handle string) string {
	t.Helper()

	profile := "https://www.linkedin.com/in/" + handle

	contact := enrichmentModel.NewCompanyContact()
	contact.CompanyDomain = "acme.test"
	contact.CompanyName = "Acme Corp"
	contact.FullName = fullName
	contact.RoleType = modelHelpers.ContactRoleHiringManager
	contact.Title = "Engineering Manager"
	contact.EnrichmentProvider = "CONTACTOUT"
	contact.LinkedinURL = &profile

	ctx := context.Background()
	_, err := f.deps.Repositories.CompanyContact.Upsert(ctx, contact)
	require.NoError(t, err)
	require.NoError(
		t,
		f.deps.Repositories.CompanyContact.LinkToJob(ctx, jobID, []string{contact.ID}),
	)

	return contact.ID
}

func (f jobFixture) createChannelTemplate(
	t *testing.T,
	channel, body string,
) outreachModel.Template {
	t.Helper()

	created := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		templatesPath(f.candidateID),
		map[string]any{
			"name": channel, "channel": channel, "body": body,
		},
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	return helpers.DecodeData[outreachModel.Template](t, created)
}

func (f jobFixture) dispatchAll(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	service := f.deps.Services.Outreach

	_, customErr := service.DispatchDueMessages(ctx)
	require.Nil(t, customErr)

	rows, err := f.deps.Postgres.GetPool().Query(ctx,
		"SELECT id, channel FROM outreach_messages WHERE status = 'SENDING'")
	require.NoError(t, err)

	claimed := map[string]string{}

	for rows.Next() {
		var id, channel string
		require.NoError(t, rows.Scan(&id, &channel))
		claimed[id] = channel
	}

	rows.Close()

	for id, channel := range claimed {
		switch modelHelpers.OutreachChannel(channel) {
		case modelHelpers.OutreachChannelLinkedinConnect:
			require.Nil(t, service.SendLinkedinInvite(ctx, id))
		case modelHelpers.OutreachChannelLinkedinDM:
			require.Nil(t, service.SendLinkedinDM(ctx, id))
		default:
			require.Nil(t, service.SendColdEmail(ctx, id))
		}
	}
}

func (f jobFixture) confirm(
	t *testing.T,
	messageID, status string,
	expected int,
) outreachModel.Message {
	t.Helper()

	confirmed := helpers.Do(t, f.router, http.MethodPost,
		candidatePath(f.candidateID)+"/outreach-messages/"+messageID+"/confirm",
		map[string]any{"status": status}, bearer(f.userToken))
	require.Equal(t, expected, confirmed.Recorder.Code, string(confirmed.Body))
	helpers.AssertMatchesSpec(t, confirmed)

	if expected != http.StatusOK {
		return outreachModel.Message{}
	}

	return helpers.DecodeData[outreachModel.Message](t, confirmed)
}

func (f jobFixture) outreachMessages(t *testing.T) []outreachModel.Message {
	t.Helper()

	listed := helpers.Do(t, f.router, http.MethodGet,
		candidatePath(f.candidateID)+"/outreach-messages?perPage=100", nil, bearer(f.userToken))
	require.Equal(t, http.StatusOK, listed.Recorder.Code, string(listed.Body))
	helpers.AssertMatchesSpec(t, listed)

	messages := helpers.DecodeData[[]outreachModel.Message](t, listed)
	sortByStep(messages)

	return messages
}

func TestLinkedinInviteWaitsForTheCandidateAndTheMessageWaitsForAcceptance(t *testing.T) {
	fixture := newLinkedinFixture(t)
	ctx := context.Background()

	invite := fixture.createChannelTemplate(
		t,
		"LINKEDIN_CONNECT",
		"Hi {{contactFirstName}}, I'd love to connect about {{jobTitle}}.",
	)
	note := fixture.createChannelTemplate(
		t,
		"LINKEDIN_DM",
		"Thanks for connecting, {{contactFirstName}}!",
	)
	fixture.approveTemplate(t, invite.ID)
	fixture.approveTemplate(t, note.ID)
	fixture.createCampaign(t, map[string]any{
		"name": "LinkedIn",
		"steps": []map[string]any{
			{"templateId": invite.ID, "delayDays": 0},
			{"templateId": note.ID, "delayDays": 0},
		},
	}, http.StatusCreated)

	planned, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	assert.Equal(t, 2, planned.Planned, "LinkedIn-only campaigns need no mailbox")

	fixture.dispatchAll(t)

	messages := fixture.outreachMessages(t)
	require.Len(t, messages, 2)
	assert.Equal(t, modelHelpers.OutreachStatusActionRequired, messages[0].Status)
	require.NotNil(t, messages[0].LinkedinNote)
	assert.Equal(
		t,
		"Hi Sarah, I'd love to connect about Senior Go Engineer.",
		*messages[0].LinkedinNote,
	)
	assert.Equal(
		t,
		modelHelpers.OutreachStatusPending,
		messages[1].Status,
		"the message waits for acceptance",
	)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), messages[1].ScheduledAt, time.Hour)

	queue := helpers.Do(t, fixture.router, http.MethodGet,
		candidatePath(fixture.candidateID)+"/outreach-messages?status=ACTION_REQUIRED", nil,
		bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, queue.Recorder.Code)
	helpers.AssertMatchesSpec(t, queue)

	todo := helpers.DecodeData[[]outreachTypes.MessageDetail](t, queue)
	require.Len(t, todo, 1)
	require.NotNil(t, todo[0].Contact)
	assert.Equal(t, "Sarah Lee", todo[0].Contact.FullName)
	require.NotNil(t, todo[0].Contact.LinkedinURL)
	assert.Equal(t, "https://www.linkedin.com/in/sarah-lee", *todo[0].Contact.LinkedinURL,
		"the queue item carries the profile to open")

	single := helpers.Do(t, fixture.router, http.MethodGet,
		candidatePath(fixture.candidateID)+"/outreach-messages/"+todo[0].ID, nil,
		bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, single.Recorder.Code)
	helpers.AssertMatchesSpec(t, single)
	assert.Equal(
		t,
		"Sarah Lee",
		helpers.DecodeData[outreachTypes.MessageDetail](t, single).Contact.FullName,
	)

	fixture.confirm(t, messages[1].ID, "SENT", http.StatusConflict)
	assert.Equal(
		t,
		modelHelpers.OutreachStatusSent,
		fixture.confirm(t, messages[0].ID, "SENT", http.StatusOK).Status,
	)
	assert.Equal(t, modelHelpers.OutreachStatusAccepted,
		fixture.confirm(t, messages[0].ID, "ACCEPTED", http.StatusOK).Status)

	accepted := fixture.storedContact(t, fixture.contactID)
	assert.Nil(t, accepted.LinkedinConfirmedAt,
		"one candidate's confirmation does not change the contact everyone shares")
	assert.Len(t, fixture.contactSignals(t, fixture.contactID), 1)

	_, err := fixture.deps.Postgres.GetPool().Exec(ctx,
		"UPDATE outreach_messages SET scheduled_at = NOW() - INTERVAL '1 minute' WHERE id = $1", messages[1].ID)
	require.NoError(t, err)
	fixture.dispatchAll(t)

	messages = fixture.outreachMessages(t)
	assert.Equal(
		t,
		modelHelpers.OutreachStatusActionRequired,
		messages[1].Status,
		"accepted, so the message is ready",
	)

	replied := fixture.confirm(t, messages[1].ID, "REPLIED", http.StatusOK)
	assert.Equal(t, modelHelpers.OutreachStatusReplied, replied.Status)
	assert.NotNil(t, replied.RepliedAt)
}

func TestTheTwentyFirstLinkedinInviteOfTheDayIsDeferred(t *testing.T) {
	fixture := newLinkedinFixture(t)
	ctx := context.Background()

	_, err := fixture.deps.Postgres.GetPool().Exec(ctx, `
		INSERT INTO outreach_messages (candidate_id, contact_id, job_id, step_number, channel, status,
			body_content, linkedin_note, scheduled_at)
		SELECT $1, $2, $3, step, 'LINKEDIN_CONNECT', 'ACTION_REQUIRED', 'Hi', 'Hi', NOW()
		FROM generate_series(1, 20) AS step
	`, fixture.candidateID, fixture.contactID, fixture.jobID)
	require.NoError(t, err)

	other := fixture.seedLinkedinContact(t, fixture.jobID, "Dana Ray", "dana-ray")
	created := helpers.Do(t, fixture.router, http.MethodPost,
		candidatePath(fixture.candidateID)+"/outreach-messages", map[string]any{
			"jobId": fixture.jobID, "contactId": other, "channels": []string{"LINKEDIN_CONNECT"},
			"linkedinNote": "Hi {{contactFirstName}}, let's connect.",
		}, bearer(fixture.userToken))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	message := helpers.DecodeData[outreachTypes.CreateMessagesResponse](t, created).Messages[0]
	assert.Equal(t, "Hi Dana, let's connect.", *message.LinkedinNote)

	fixture.dispatchAll(t)

	var status string

	var scheduledAt time.Time

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		"SELECT status, scheduled_at FROM outreach_messages WHERE id = $1", message.ID,
	).Scan(&status, &scheduledAt))
	assert.Equal(t, "PENDING", status, "twenty invites already wait for the candidate")
	assert.True(t, scheduledAt.After(time.Now().Add(30*time.Minute)))
}

func TestManualLinkedinMessagesAndTemplatesAreValidated(t *testing.T) {
	fixture := newLinkedinFixture(t)
	path := candidatePath(fixture.candidateID) + "/outreach-messages"

	missingNote := helpers.Do(t, fixture.router, http.MethodPost, path, map[string]any{
		"jobId":     fixture.jobID,
		"contactId": fixture.contactID,
		"channels":  []string{"LINKEDIN_CONNECT", "EMAIL"},
	}, bearer(fixture.userToken))
	require.Equal(t, http.StatusUnprocessableEntity, missingNote.Recorder.Code)
	helpers.AssertMatchesSpec(t, missingNote)
	assert.ElementsMatch(
		t,
		[]string{"emailBody", "linkedinNote", "subject"},
		detailFields(t, missingNote),
	)

	longNote := strings.Repeat("{{jobTitle}} ", 40)
	tooLong := helpers.Do(t, fixture.router, http.MethodPost, path, map[string]any{
		"jobId":        fixture.jobID,
		"contactId":    fixture.contactID,
		"channels":     []string{"LINKEDIN_CONNECT"},
		"linkedinNote": longNote[:290],
	}, bearer(fixture.userToken))
	assert.Equal(
		t,
		http.StatusUnprocessableEntity,
		tooLong.Recorder.Code,
		"the limit applies after placeholders",
	)

	noEmail := helpers.Do(t, fixture.router, http.MethodPost, path, map[string]any{
		"jobId": fixture.jobID, "contactId": fixture.contactID, "channels": []string{"EMAIL"},
		"subject": "Hi", "emailBody": "Hello",
	}, bearer(fixture.userToken))
	assert.Equal(
		t,
		http.StatusUnprocessableEntity,
		noEmail.Recorder.Code,
		"this contact has only LinkedIn",
	)

	template := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		templatesPath(fixture.candidateID),
		map[string]any{
			"name": "Too long", "channel": "LINKEDIN_CONNECT", "body": strings.Repeat("a", 301),
		},
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusUnprocessableEntity, template.Recorder.Code)
}
