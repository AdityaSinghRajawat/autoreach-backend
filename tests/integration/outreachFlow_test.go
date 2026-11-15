//go:build integration

package integration

import (
	"context"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const (
	contactAddress  = "sarah@acme.test"
	contactPassword = "sarah-password"
)

type outreachFixture struct {
	jobFixture
	accountID string
	jobID     string
	contactID string
}

func newOutreachFixture(t *testing.T) outreachFixture {
	t.Helper()

	fixture := newJobFixture(t)
	account := fixture.connectMailbox(t, senderPassword)
	require.Equal(
		t,
		modelHelpers.ChannelAccountStatusActive,
		fixture.verifyMailbox(t, account.ID).Status,
	)

	jobID := fixture.ingest(t, withFields(validJob(), map[string]any{"companyDomain": "acme.test"})).
		QueuedTasks[0].JobID

	return outreachFixture{
		jobFixture: fixture,
		accountID:  account.ID,
		jobID:      jobID,
		contactID:  fixture.seedContact(t, jobID, "Sarah Lee", contactAddress),
	}
}

func (f jobFixture) seedContact(t *testing.T, jobID, fullName, address string) string {
	t.Helper()

	contact := enrichmentModel.NewCompanyContact()
	contact.CompanyDomain = "acme.test"
	contact.CompanyName = "Acme Corp"
	contact.FullName = fullName
	contact.RoleType = modelHelpers.ContactRoleHiringManager
	contact.Title = "Engineering Manager"
	contact.EnrichmentProvider = "APOLLO"
	contact.SetEmail(address, modelHelpers.EmailStatusVerified)

	ctx := context.Background()
	_, err := f.deps.Repositories.CompanyContact.Upsert(ctx, contact)
	require.NoError(t, err)
	require.NoError(
		t,
		f.deps.Repositories.CompanyContact.LinkToJob(ctx, jobID, []string{contact.ID}),
	)

	return contact.ID
}

func (f outreachFixture) startCampaign(t *testing.T) {
	t.Helper()

	first := f.createTemplate(
		t,
		"{{jobTitle}} at {{companyName}}",
		"Hi {{contactFirstName}}, I'm {{candidateName}}.",
	)
	followUp := f.createTemplate(t, "Re: {{jobTitle}}", "Following up, {{contactFirstName}}.")
	f.approveTemplate(t, first.ID)
	f.approveTemplate(t, followUp.ID)

	f.createCampaign(t, map[string]any{
		"name": "Default",
		"steps": []map[string]any{
			{"templateId": first.ID, "delayDays": 0},
			{"templateId": followUp.ID, "delayDays": 3},
		},
	}, http.StatusCreated)
}

func (f outreachFixture) dispatchAndSend(t *testing.T) []string {
	t.Helper()

	ctx := context.Background()
	service := f.deps.Services.Outreach

	_, customErr := service.DispatchDueMessages(ctx)
	require.Nil(t, customErr)

	var claimed []string

	rows, err := f.deps.Postgres.GetPool().Query(ctx,
		"SELECT id FROM outreach_messages WHERE status = 'SENDING' ORDER BY step_number")
	require.NoError(t, err)

	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		claimed = append(claimed, id)
	}

	rows.Close()

	for _, id := range claimed {
		require.Nil(t, service.SendColdEmail(ctx, id))
	}

	return claimed
}

func (f outreachFixture) messages(t *testing.T) []outreachModel.Message {
	t.Helper()

	listed := helpers.Do(t, f.router, http.MethodGet,
		candidatePath(f.candidateID)+"/outreach-messages?perPage=100", nil, bearer(f.userToken))
	require.Equal(t, http.StatusOK, listed.Recorder.Code, string(listed.Body))
	helpers.AssertMatchesSpec(t, listed)

	messages := helpers.DecodeData[[]outreachModel.Message](t, listed)
	sortByStep(messages)

	return messages
}

func TestCampaignEmailIsSentAndAReplyPausesTheFollowUps(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	planned, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	assert.Equal(t, 2, planned.Planned)

	again, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	assert.Zero(t, again.Planned, "planning is idempotent and never contacts the same person twice")

	sent := fixture.dispatchAndSend(t)
	require.Len(t, sent, 1, "the follow-up waits for its delay")

	messages := fixture.messages(t)
	require.Len(t, messages, 2)
	assert.Equal(t, modelHelpers.OutreachStatusSent, messages[0].Status)
	assert.Equal(t, "Senior Go Engineer at Acme Corp", *messages[0].Subject)
	assert.Equal(t, "Hi Sarah, I'm Alex Candidate.", messages[0].BodyContent)
	assert.Equal(t, modelHelpers.OutreachStatusPending, messages[1].Status)

	delivered := fetchInbox(t, contactAddress, contactPassword)
	original := findBySubject(t, delivered, "Senior Go Engineer at Acme Corp")

	replyTo(t, contactAddress, contactPassword, senderAddress, original.MessageID)
	require.Nil(t, fixture.deps.Services.Outreach.SweepMailbox(ctx, fixture.accountID))

	messages = fixture.messages(t)
	assert.Equal(t, modelHelpers.OutreachStatusReplied, messages[0].Status)
	assert.NotNil(t, messages[0].RepliedAt)
	assert.Equal(
		t,
		modelHelpers.OutreachStatusPaused,
		messages[1].Status,
		"a reply pauses every later step",
	)
}

func TestTheThirtyFirstEmailOfTheDayIsDeferred(t *testing.T) {
	fixture := newOutreachFixture(t)
	ctx := context.Background()

	_, err := fixture.deps.Postgres.GetPool().Exec(ctx, `
		INSERT INTO outreach_messages (candidate_id, contact_id, job_id, channel_account_id, step_number,
			channel, status, subject, body_content, scheduled_at, sent_at)
		SELECT $1, $2, $3, $4, step, 'EMAIL', 'SENT', 'Earlier', 'Earlier', NOW(), NOW() - INTERVAL '1 hour'
		FROM generate_series(1, 30) AS step
	`, fixture.candidateID, fixture.contactID, fixture.jobID, fixture.accountID)
	require.NoError(t, err)

	otherContact := fixture.seedContact(t, fixture.jobID, "Dana Ray", "dana@acme.test")
	created := helpers.Do(t, fixture.router, http.MethodPost,
		candidatePath(fixture.candidateID)+"/outreach-messages", map[string]any{
			"jobId": fixture.jobID, "contactId": otherContact, "channels": []string{"EMAIL"},
			"subject": "Hello {{contactFirstName}}", "emailBody": "About {{jobTitle}}",
		}, bearer(fixture.userToken))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)
	message := helpers.DecodeData[outreachTypes.CreateMessagesResponse](t, created).Messages[0]
	assert.Equal(t, "Hello Dana", *message.Subject)

	fixture.dispatchAndSend(t)

	deferred := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(
			fixture.candidateID,
		)+"/outreach-messages/"+message.ID,
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, deferred.Recorder.Code)
	helpers.AssertMatchesSpec(t, deferred)

	stored := helpers.DecodeData[outreachModel.Message](t, deferred)
	assert.Equal(t, modelHelpers.OutreachStatusPending, stored.Status)
	assert.Nil(t, stored.SentAt)
	assert.WithinDuration(t, time.Now().Add(23*time.Hour), stored.ScheduledAt, 10*time.Minute,
		"it waits until the oldest send of the last day falls out of the window")
}

func TestBounceEndsTheConversationAndSuppressesTheAddress(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	fixture.dispatchAndSend(t)

	var providerMessageID string

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		"SELECT provider_message_id FROM outreach_messages WHERE step_number = 1",
	).Scan(&providerMessageID))

	sendRaw(t, contactAddress, contactPassword, "mailer-daemon@example.com", senderAddress,
		deliveryStatusNotification(senderAddress, contactAddress, providerMessageID))
	require.Nil(t, fixture.deps.Services.Outreach.SweepMailbox(ctx, fixture.accountID))

	messages := fixture.messages(t)
	assert.Equal(t, modelHelpers.OutreachStatusBounced, messages[0].Status)
	assert.Equal(t, modelHelpers.OutreachStatusCancelled, messages[1].Status)

	retry := helpers.Do(t, fixture.router, http.MethodPost,
		candidatePath(fixture.candidateID)+"/outreach-messages", map[string]any{
			"jobId": fixture.jobID, "contactId": fixture.contactID, "channels": []string{"EMAIL"},
			"subject": "Hello", "emailBody": "Again",
		}, bearer(fixture.userToken))
	require.Equal(
		t,
		http.StatusUnprocessableEntity,
		retry.Recorder.Code,
		"a bounced address is suppressed",
	)
}

func TestManualOutreachRejectsUnsupportedChannelsAndCanBeCancelled(t *testing.T) {
	fixture := newOutreachFixture(t)
	path := candidatePath(fixture.candidateID) + "/outreach-messages"
	request := map[string]any{
		"jobId":     fixture.jobID,
		"contactId": fixture.contactID,
		"subject":   "Hi",
		"emailBody": "Hello",
	}

	request["channels"] = []string{"EMAIL", "LINKEDIN_CONNECT"}
	linkedin := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		path,
		request,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusUnprocessableEntity, linkedin.Recorder.Code)

	request["channels"] = []string{"EMAIL"}
	created := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		path,
		request,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	messageID := helpers.DecodeData[outreachTypes.CreateMessagesResponse](t, created).Messages[0].ID

	duplicate := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		path,
		request,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code)

	cancelled := helpers.Do(t, fixture.router, http.MethodPatch, path+"/"+messageID,
		map[string]any{"status": "CANCELLED"}, bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, cancelled.Recorder.Code)
	helpers.AssertMatchesSpec(t, cancelled)
	assert.Equal(t, modelHelpers.OutreachStatusCancelled,
		helpers.DecodeData[outreachModel.Message](t, cancelled).Status)

	again := helpers.Do(t, fixture.router, http.MethodPatch, path+"/"+messageID,
		map[string]any{"status": "CANCELLED"}, bearer(fixture.userToken))
	assert.Equal(t, http.StatusConflict, again.Recorder.Code)
}

func mailAccountFor(t *testing.T, address, password string) sharedTypes.EmailAccount {
	t.Helper()

	smtpPort, err := strconv.Atoi(os.Getenv("TEST_SMTP_PORT"))
	require.NoError(t, err)
	imapPort, err := strconv.Atoi(os.Getenv("TEST_IMAP_PORT"))
	require.NoError(t, err)

	host := os.Getenv("TEST_MAIL_HOST")

	return sharedTypes.EmailAccount{
		Address:  address,
		Username: address,
		Password: password,
		SMTP:     sharedTypes.EmailServer{Host: host, Port: smtpPort, Security: "NONE"},
		IMAP:     sharedTypes.EmailServer{Host: host, Port: imapPort, Security: "NONE"},
	}
}

func fetchInbox(t *testing.T, address, password string) []sharedTypes.InboundEmail {
	t.Helper()

	client := utils.NewEmailUtil()
	account := mailAccountFor(t, address, password)

	cursor, err := client.VerifyEmailAccount(context.Background(), account)
	require.NoError(t, err)

	inbound, _, err := client.FetchInboxSince(context.Background(), account,
		sharedTypes.MailboxCursor{UIDValidity: cursor.UIDValidity})
	require.NoError(t, err)

	return inbound
}

func findBySubject(
	t *testing.T,
	inbound []sharedTypes.InboundEmail,
	_ string,
) sharedTypes.InboundEmail {
	t.Helper()
	require.NotEmpty(t, inbound, "the contact's mailbox received the email")

	return inbound[len(inbound)-1]
}

func replyTo(t *testing.T, from, password, to, inReplyTo string) {
	t.Helper()

	require.NoError(
		t,
		utils.NewEmailUtil().SendEmail(context.Background(), mailAccountFor(t, from, password),
			sharedTypes.OutgoingEmail{
				FromAddress: from,
				ToAddress:   to,
				Subject:     "Re: Senior Go Engineer at Acme Corp",
				Body:        "Thanks, let's talk.",
				MessageID:   utils.NewEmailMessageID(from),
				InReplyTo:   inReplyTo,
				Date:        time.Now(),
			}),
	)
}

func sendRaw(t *testing.T, login, password, from, to, raw string) {
	t.Helper()

	host := os.Getenv("TEST_MAIL_HOST")
	address := net.JoinHostPort(host, os.Getenv("TEST_SMTP_PORT"))
	require.NoError(t, smtp.SendMail(address, smtp.PlainAuth("", login, password, host), from,
		[]string{to}, []byte(strings.ReplaceAll(raw, "\n", "\r\n"))))
}

func deliveryStatusNotification(to, failedRecipient, originalMessageID string) string {
	return `From: Mail Delivery System <mailer-daemon@example.com>
To: ` + to + `
Subject: Undelivered Mail Returned to Sender
Message-ID: ` + utils.NewEmailMessageID("mailer-daemon@example.com") + `
MIME-Version: 1.0
Content-Type: multipart/report; report-type=delivery-status; boundary="report"

--report
Content-Type: text/plain

The message could not be delivered.

--report
Content-Type: message/delivery-status

Reporting-MTA: dns; mail.acme.test
Final-Recipient: rfc822; ` + failedRecipient + `
Action: failed
Status: 5.1.1

--report
Content-Type: text/rfc822-headers

Message-ID: ` + originalMessageID + `
Subject: Senior Go Engineer at Acme Corp

--report--
`
}

func sortByStep(messages []outreachModel.Message) {
	for i := 1; i < len(messages); i++ {
		for j := i; j > 0 && messages[j].StepNumber < messages[j-1].StepNumber; j-- {
			messages[j], messages[j-1] = messages[j-1], messages[j]
		}
	}
}

func TestTrackingRecordsOpensAndOnlyFollowsSignedLinks(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	enabled := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		candidatePath(fixture.candidateID),
		map[string]any{
			"settings": map[string]any{"emailTracking": true},
		},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, enabled.Recorder.Code, string(enabled.Body))

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	fixture.dispatchAndSend(t)

	var token, messageID string

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		"SELECT id, tracking_token FROM outreach_messages WHERE step_number = 1",
	).Scan(&messageID, &token))

	opened := helpers.Do(t, fixture.router, http.MethodGet, "/t/o/"+token, nil, nil)
	require.Equal(t, http.StatusOK, opened.Recorder.Code)
	assert.Equal(t, "image/gif", opened.Recorder.Header().Get("Content-Type"))

	link := "https://acme.test/careers"
	signature := utils.Sign(
		utils.DeriveKey(config.GetJWTSecret(), config.GetTrackingKeyLabel()),
		token,
		link,
	)
	query := url.Values{"u": {link}, "s": {signature}}

	clicked := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/t/c/"+token+"?"+query.Encode(),
		nil,
		nil,
	)
	require.Equal(t, http.StatusFound, clicked.Recorder.Code)
	assert.Equal(t, link, clicked.Recorder.Header().Get("Location"))

	query.Set("u", "https://evil.test")
	forged := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/t/c/"+token+"?"+query.Encode(),
		nil,
		nil,
	)
	assert.Equal(t, http.StatusNotFound, forged.Recorder.Code, "unsigned links are not redirected")

	var events []string

	rows, err := fixture.deps.Postgres.GetPool().Query(ctx,
		"SELECT event_type FROM outreach_events WHERE message_id = $1 ORDER BY occurred_at", messageID)
	require.NoError(t, err)

	for rows.Next() {
		var eventType string
		require.NoError(t, rows.Scan(&eventType))
		events = append(events, eventType)
	}

	rows.Close()
	assert.ElementsMatch(t, []string{"SENT", "OPENED", "CLICKED"}, events)
}

func TestUnsubscribeLinkStopsTheSenderAndOnlyThePostChangesAnything(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	fixture.dispatchAndSend(t)

	messageID := fixture.messages(t)[0].ID
	signature := utils.Sign(
		utils.DeriveKey(config.GetJWTSecret(), config.GetUnsubscribeKeyLabel()),
		messageID,
	)
	path := "/u/" + messageID + "?s=" + signature

	forged := helpers.Do(t, fixture.router, http.MethodGet, "/u/"+messageID+"?s=forged", nil, nil)
	assert.Equal(t, http.StatusNotFound, forged.Recorder.Code)

	page := helpers.Do(t, fixture.router, http.MethodGet, path, nil, nil)
	require.Equal(t, http.StatusOK, page.Recorder.Code)
	assert.Contains(t, string(page.Body), "<form method=\"post\"")
	assert.Equal(
		t,
		modelHelpers.OutreachStatusPending,
		fixture.messages(t)[1].Status,
		"opening the link changes nothing",
	)

	unsubscribed := helpers.Do(t, fixture.router, http.MethodPost, path, nil, nil)
	require.Equal(t, http.StatusOK, unsubscribed.Recorder.Code)
	assert.Contains(t, string(unsubscribed.Body), "You have been unsubscribed")
	assert.Equal(t, modelHelpers.OutreachStatusCancelled, fixture.messages(t)[1].Status)

	var scopedToCandidate bool

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		"SELECT candidate_id IS NOT NULL FROM suppression_entries WHERE value = $1 AND reason = 'UNSUBSCRIBED'",
		contactAddress,
	).
		Scan(&scopedToCandidate))
	assert.True(t, scopedToCandidate, "the opt-out applies to this sender only")

	retry := helpers.Do(t, fixture.router, http.MethodPost,
		candidatePath(fixture.candidateID)+"/outreach-messages", map[string]any{
			"jobId": fixture.jobID, "contactId": fixture.contactID, "channels": []string{"EMAIL"},
			"subject": "Hello", "emailBody": "Again",
		}, bearer(fixture.userToken))
	assert.Equal(t, http.StatusUnprocessableEntity, retry.Recorder.Code)
}

func TestReplyingUnsubscribeIsTreatedAsAnOptOut(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)
	fixture.dispatchAndSend(t)

	original := findBySubject(t, fetchInbox(t, contactAddress, contactPassword), "")

	require.NoError(
		t,
		utils.NewEmailUtil().SendEmail(ctx, mailAccountFor(t, contactAddress, contactPassword),
			sharedTypes.OutgoingEmail{
				FromAddress: contactAddress,
				ToAddress:   senderAddress,
				Subject:     "Re: unsubscribe",
				Body:        "Please stop.",
				MessageID:   utils.NewEmailMessageID(contactAddress),
				InReplyTo:   original.MessageID,
				Date:        time.Now(),
			}),
	)
	require.Nil(t, fixture.deps.Services.Outreach.SweepMailbox(ctx, fixture.accountID))

	messages := fixture.messages(t)
	assert.Equal(
		t,
		modelHelpers.OutreachStatusSent,
		messages[0].Status,
		"an opt-out is not counted as a reply",
	)
	assert.Equal(t, modelHelpers.OutreachStatusCancelled, messages[1].Status)
}
