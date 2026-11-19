//go:build integration

package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

type dashboardSeed struct {
	matchedJobID string
	emailOneID   string
	baseTime     time.Time
}

func (f jobFixture) seedDashboard(t *testing.T) dashboardSeed {
	t.Helper()

	ingested := f.ingest(
		t,
		validJob(),
		withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer"}),
	)
	matchedJobID, otherJobID := ingested.QueuedTasks[0].JobID, ingested.QueuedTasks[1].JobID
	base := time.Now().UTC().Add(-6 * time.Hour).Truncate(time.Second)
	ctx := context.Background()
	pool := f.deps.Postgres.GetPool()

	exec := func(query string, arguments ...any) {
		t.Helper()

		_, err := pool.Exec(ctx, query, arguments...)
		require.NoError(t, err)
	}

	resumeID, contactOne, contactTwo, contactThree := utils.GenerateUUID(), utils.GenerateUUID(), utils.GenerateUUID(), utils.GenerateUUID()
	emailOne, emailTwo, invite, applicationID := utils.GenerateUUID(), utils.GenerateUUID(), utils.GenerateUUID(), utils.GenerateUUID()

	exec(
		`INSERT INTO candidate_resumes (id, candidate_id, file_name, storage_key, content_type, size_bytes,
		content_sha256) VALUES ($1, $2, 'resume.pdf', 'resumes/x', 'application/pdf', 10, 'abc')`,
		resumeID,
		f.candidateID,
	)
	exec(
		`INSERT INTO job_matches (candidate_id, job_id, resume_id, raw_similarity, match_score, created_at)
		VALUES ($1, $2, $3, 0.9, 0.91, $5), ($1, $4, $3, 0.5, 0.50, $5)`,
		f.candidateID,
		matchedJobID,
		resumeID,
		otherJobID,
		base,
	)

	for _, contact := range []struct{ id, name, email, jobID string }{
		{contactOne, "Riley Recruiter", "riley@acme.com", matchedJobID},
		{contactTwo, "Morgan Manager", "morgan@acme.com", matchedJobID},
		{contactThree, "Unmatched Contact", "other@acme.com", otherJobID},
	} {
		exec(
			`INSERT INTO company_contacts (id, company_domain, company_name, full_name, role_type, title,
			work_email, enrichment_provider) VALUES ($1, 'acme.com', 'Acme Corp', $2, 'RECRUITER', 'Recruiter', $3, 'APOLLO')`,
			contact.id,
			contact.name,
			contact.email,
		)
		exec(
			`INSERT INTO job_contacts (job_id, contact_id) VALUES ($1, $2)`,
			contact.jobID,
			contact.id,
		)
	}

	exec(
		`INSERT INTO applications (id, candidate_id, job_id, match_score, status, ats_type, apply_url,
		submitted_at, updated_at) VALUES ($1, $2, $3, 0.91, 'INTERVIEWING', 'GREENHOUSE',
		'https://boards.greenhouse.io/acme/jobs/10293', $4, $5)`,
		applicationID,
		f.candidateID,
		matchedJobID,
		base.Add(time.Hour),
		base.Add(2*time.Hour),
	)

	message := `INSERT INTO outreach_messages (id, candidate_id, contact_id, job_id, channel, status, body_content,
		scheduled_at, sent_at, replied_at) VALUES ($1, $2, $3, $4, $5, $6, 'Hello', $7, $7, $8)`
	exec(
		message,
		emailOne,
		f.candidateID,
		contactOne,
		matchedJobID,
		"EMAIL",
		"SENT",
		base.Add(3*time.Hour),
		nil,
	)
	exec(
		message,
		emailTwo,
		f.candidateID,
		contactTwo,
		matchedJobID,
		"EMAIL",
		"REPLIED",
		base.Add(3*time.Hour),
		base.Add(5*time.Hour),
	)
	exec(message, invite, f.candidateID, contactOne, matchedJobID, "LINKEDIN_CONNECT", "ACCEPTED",
		base.Add(3*time.Hour), nil)

	event := `INSERT INTO outreach_events (message_id, event_type, occurred_at) VALUES ($1, $2, $3)`
	exec(event, emailOne, "SENT", base.Add(3*time.Hour))
	exec(event, emailOne, "OPENED", base.Add(4*time.Hour))
	exec(event, emailTwo, "REPLIED", base.Add(5*time.Hour))

	return dashboardSeed{matchedJobID: matchedJobID, emailOneID: emailOne, baseTime: base}
}

func TestCandidateMetricsMatchTheSpecifiedShape(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	exchange := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(fixture.candidateID)+"/metrics",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	assert.JSONEq(t, `{
		"candidateId": "`+fixture.candidateID+`",
		"metrics": {
			"totalJobsScraped": 2,
			"jobsMatched": 1,
			"applicationsSubmitted": 1,
			"contactsDiscovered": 2,
			"coldEmailsSent": 2,
			"linkedinRequestsSent": 1,
			"linkedinConnectionsAccepted": 1,
			"recruiterResponses": 1,
			"responseRatePercentage": 50
		}
	}`, string(helpers.DecodeEnvelope(t, exchange).Data))

	other := register(t, fixture.router, "other@example.com")
	hidden := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(fixture.candidateID)+"/metrics",
		nil,
		bearer(other.AccessToken),
	)
	assert.Equal(
		t,
		http.StatusNotFound,
		hidden.Recorder.Code,
		"another candidate's metrics are not visible",
	)

	admin := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(fixture.candidateID)+"/metrics",
		nil,
		bearer(loginAdmin(t, fixture.router)),
	)
	assert.Equal(t, http.StatusOK, admin.Recorder.Code)
}

func TestDashboardFunnelUsesTheSignedInCandidate(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	exchange := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/api/v1/dashboard/funnel",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	funnel := helpers.DecodeData[dashboardTypes.FunnelResponse](t, exchange)
	assert.Equal(t, fixture.candidateID, funnel.CandidateID)
	assert.Equal(t, []dashboardTypes.FunnelStage{
		{Stage: "JOBS_SCRAPED", Count: 2},
		{Stage: "JOBS_MATCHED", Count: 1},
		{Stage: "APPLICATIONS_SUBMITTED", Count: 1},
		{Stage: "COLD_EMAILS_SENT", Count: 2},
		{Stage: "LINKEDIN_REQUESTS_SENT", Count: 1},
		{Stage: "RECRUITER_RESPONSES", Count: 1},
		{Stage: "INTERVIEWS", Count: 1},
	}, funnel.Stages)

	other := register(t, fixture.router, "other@example.com")
	foreign := helpers.Do(t, fixture.router, http.MethodGet,
		"/api/v1/dashboard/funnel?candidateId="+fixture.candidateID, nil, bearer(other.AccessToken))
	assert.Equal(
		t,
		http.StatusNotFound,
		foreign.Recorder.Code,
		"candidates cannot pick another candidate",
	)

	adminWithoutCandidate := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/api/v1/dashboard/funnel",
		nil,
		bearer(loginAdmin(t, fixture.router)),
	)
	assert.Equal(t, http.StatusNotFound, adminWithoutCandidate.Recorder.Code)
}

func TestActivityFeedPagesNewestFirst(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	list := func(query string) dashboardTypes.ActivityPage {
		t.Helper()

		exchange := helpers.Do(
			t,
			fixture.router,
			http.MethodGet,
			"/api/v1/dashboard/activity"+query,
			nil,
			bearer(fixture.userToken),
		)
		require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
		helpers.AssertMatchesSpec(t, exchange)

		return helpers.DecodeData[dashboardTypes.ActivityPage](t, exchange)
	}

	first := list("?limit=4")
	require.NotNil(t, first.NextCursor)
	second := list("?limit=4&before=" + *first.NextCursor)

	eventTypes := []string{}
	for _, item := range append(first.Items, second.Items...) {
		eventTypes = append(eventTypes, item.EventType)
	}

	assert.Equal(t, []string{
		"OUTREACH_REPLIED", "OUTREACH_OPENED", "OUTREACH_SENT", "APPLICATION_INTERVIEWING",
		"APPLICATION_SUBMITTED", "JOB_MATCHED",
	}, eventTypes, "the match below the threshold is not shown")
	assert.Nil(t, second.NextCursor)

	reply := first.Items[0]
	assert.Equal(t, "Morgan Manager", *reply.ContactName)
	assert.Equal(t, "EMAIL", *reply.Channel)
	assert.Equal(t, "Senior Go Engineer", reply.JobTitle)

	invalid := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/api/v1/dashboard/activity?before=not-a-cursor",
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusBadRequest, invalid.Recorder.Code)
}

func TestActivityStreamResumesAndDeliversNewEvents(t *testing.T) {
	fixture := newJobFixture(t)
	seed := fixture.seedDashboard(t)
	server := httptest.NewServer(fixture.router)
	t.Cleanup(server.Close)

	oldest := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		"/api/v1/dashboard/activity?limit=100",
		nil,
		bearer(fixture.userToken),
	)
	items := helpers.DecodeData[dashboardTypes.ActivityPage](t, oldest).Items
	matched := items[len(items)-1]
	require.Equal(t, "JOB_MATCHED", matched.EventType)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		server.URL+"/api/v1/dashboard/activity/stream",
		nil,
	)
	require.NoError(t, err)
	request.Header.Set(config.GetAuthorizationHeader(), "Bearer "+fixture.userToken)
	request.Header.Set(config.GetLastEventIDHeader(), matched.Cursor().Encode())

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })

	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, config.GetSSEContentType(), response.Header.Get(config.GetContentTypeHeader()))

	lines := bufio.NewScanner(response.Body)
	nextEventType := func() string {
		t.Helper()

		for lines.Scan() {
			if data, found := strings.CutPrefix(lines.Text(), "data: "); found {
				event := struct {
					EventType string `json:"eventType"`
				}{}
				require.NoError(t, json.Unmarshal([]byte(data), &event))

				return event.EventType
			}
		}

		t.Fatalf("stream ended: %v", lines.Err())

		return ""
	}

	resumed := []string{}
	for range 5 {
		resumed = append(resumed, nextEventType())
	}

	assert.Equal(t, []string{
		"APPLICATION_SUBMITTED",
		"APPLICATION_INTERVIEWING",
		"OUTREACH_SENT",
		"OUTREACH_OPENED",
		"OUTREACH_REPLIED",
	}, resumed, "events after Last-Event-ID arrive oldest first")

	_, err = fixture.deps.Postgres.GetPool().Exec(context.Background(),
		`INSERT INTO outreach_events (message_id, event_type, occurred_at) VALUES ($1, 'CLICKED', $2)`,
		seed.emailOneID, time.Now().UTC())
	require.NoError(t, err)

	assert.Equal(t, "OUTREACH_CLICKED", nextEventType(), "new events are pushed while connected")
}

func TestACandidateCannotHoldMoreThanTheAllowedNumberOfStreams(t *testing.T) {
	fixture := newJobFixture(t)

	for range config.GetActivityStreamsPerCandidate() {
		seated, _, err := fixture.deps.RedisUtil.TakeSeat(
			context.Background(),
			utils.BuildKey(
				config.GetKeyDelimiter(),
				config.GetKeyPrefix(),
				config.GetSeatNamespace(),
				fixture.candidateID,
			),
			config.GetActivityStreamsPerCandidate(),
			time.Minute,
		)
		require.NoError(t, err)
		require.True(t, seated)
	}

	refused := helpers.Do(t, fixture.router, http.MethodGet, "/api/v1/dashboard/activity/stream",
		nil, bearer(fixture.userToken))
	assert.Equal(t, http.StatusTooManyRequests, refused.Recorder.Code, string(refused.Body))
	helpers.AssertMatchesSpec(t, refused)
}
