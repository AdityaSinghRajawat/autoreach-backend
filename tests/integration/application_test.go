//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const whyHereQuestion = "Why do you want to work here?"

type atsStub struct {
	mutex       sync.Mutex
	server      *httptest.Server
	submissions []map[string][]string
	resumeNames []string
	failSubmits bool
}

func newATSStub(t *testing.T) *atsStub {
	t.Helper()

	stub := &atsStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(stub.handle))
	t.Cleanup(stub.server.Close)

	return stub
}

func (s *atsStub) handle(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/greenhouse/boards/acme/jobs/10293":
		_, _ = io.WriteString(writer, `{"id":10293,"questions":[
			{"required":true,"label":"First Name","fields":[{"name":"first_name","type":"input_text"}]},
			{"required":true,"label":"Last Name","fields":[{"name":"last_name","type":"input_text"}]},
			{"required":true,"label":"Email","fields":[{"name":"email","type":"input_text"}]},
			{"required":false,"label":"Phone","fields":[{"name":"phone","type":"input_text"}]},
			{"required":true,"label":"Resume/CV","fields":[{"name":"resume","type":"input_file"},{"name":"resume_text","type":"textarea"}]},
			{"required":true,"label":"LinkedIn Profile","fields":[{"name":"question_1","type":"input_text"}]},
			{"required":true,"label":"Are you legally authorized to work in the US?","fields":[{"name":"question_2","type":"multi_value_single_select","values":[{"label":"Yes","value":1},{"label":"No","value":0}]}]},
			{"required":true,"label":"`+whyHereQuestion+`","fields":[{"name":"question_3","type":"textarea"}]}
		]}`)
	case request.Method == http.MethodPost && request.URL.Path == "/greenhouse/boards/acme/jobs/10293":
		s.record(request)

		if s.failing() {
			writer.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, _ = io.WriteString(writer, `{}`)
	case request.Method == http.MethodGet && request.URL.Path == "/lever/postings/acme/abc-123":
		_, _ = io.WriteString(writer, `{"id":"abc-123","text":"Senior Go Engineer"}`)
	case request.Method == http.MethodPost && request.URL.Path == "/lever/postings/acme/abc-123" &&
		request.URL.Query().Get("key") == "lever-key":
		s.record(request)

		_, _ = io.WriteString(writer, `{"ok":true,"applicationId":"lever-application-1"}`)
	default:
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (s *atsStub) record(request *http.Request) {
	if err := request.ParseMultipartForm(1 << 20); err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.submissions = append(s.submissions, request.MultipartForm.Value)
	for _, file := range request.MultipartForm.File["resume"] {
		s.resumeNames = append(s.resumeNames, file.Filename)
	}
}

func (s *atsStub) failing() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.failSubmits
}

func (s *atsStub) submissionCount() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.submissions)
}

type applicationFixture struct {
	jobFixture
	stub *atsStub
}

func newApplicationFixture(
	t *testing.T,
	greenhouseKeys, leverKeys map[string]string,
) applicationFixture {
	t.Helper()

	fixture := newJobFixture(t)
	stub := newATSStub(t)

	fixture.deps.Appliers = appliers.NewApplierRegistry(appliers.RegistryOptions{
		GreenhouseAPIURL: stub.server.URL + "/greenhouse",
		LeverAPIURL:      stub.server.URL + "/lever",
		LeverEUAPIURL:    stub.server.URL + "/lever-eu",
		GreenhouseKeys:   greenhouseKeys,
		LeverKeys:        leverKeys,
	})
	require.NoError(t, fixture.deps.Wire())
	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		fixture.deps.Appliers = nil
		require.NoError(t, fixture.deps.Wire())
	})

	fixture.processedResume(t, "Alex Candidate", "Go engineer")

	return applicationFixture{jobFixture: fixture, stub: stub}
}

func (f applicationFixture) saveProfile(t *testing.T, answers map[string]string) {
	t.Helper()

	saved := helpers.Do(
		t,
		f.router,
		http.MethodPut,
		candidatePath(f.candidateID)+"/application-profile",
		map[string]any{
			"phone":          "+1 555 0100",
			"linkedinUrl":    "https://www.linkedin.com/in/alex-candidate",
			"workAuthorized": true,
			"answers":        answers,
		},
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, saved.Recorder.Code, string(saved.Body))
	helpers.AssertMatchesSpec(t, saved)
}

func (f applicationFixture) apply(t *testing.T, job map[string]any) applicationModel.Application {
	t.Helper()

	jobID := f.ingest(t, job).QueuedTasks[0].JobID

	created := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		candidatePath(f.candidateID)+"/applications",
		map[string]any{"jobId": jobID},
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	application := helpers.DecodeData[applicationModel.Application](t, created)
	assert.Equal(t, modelHelpers.ApplicationStatusQueued, application.Status)

	require.Nil(
		t,
		f.deps.Services.Application.PrepareApplication(context.Background(), application.ID),
	)

	return *f.application(t, application.ID).Application
}

func (f applicationFixture) application(
	t *testing.T,
	applicationID string,
) applicationTypes.ApplicationDetail {
	t.Helper()

	fetched := helpers.Do(t, f.router, http.MethodGet,
		candidatePath(f.candidateID)+"/applications/"+applicationID, nil, bearer(f.userToken))
	require.Equal(t, http.StatusOK, fetched.Recorder.Code, string(fetched.Body))
	helpers.AssertMatchesSpec(t, fetched)

	return helpers.DecodeData[applicationTypes.ApplicationDetail](t, fetched)
}

func decodeList[T any](t *testing.T, raw json.RawMessage) []T {
	t.Helper()

	values := []T{}
	require.NoError(t, json.Unmarshal(raw, &values))

	return values
}

func TestApplicationIsPreparedAndLeftForTheCandidateWhenAnAnswerIsMissing(t *testing.T) {
	fixture := newApplicationFixture(t, nil, nil)
	fixture.saveProfile(t, nil)

	application := fixture.apply(t, validJob())
	assert.Equal(t, modelHelpers.ApplicationStatusActionRequired, application.Status)
	assert.Equal(t, "https://boards.greenhouse.io/acme/jobs/10293", application.ApplyURL)
	assert.NotNil(t, application.ResumeID)

	answers := map[string]string{}
	for _, answer := range decodeList[applicationTypes.PreparedAnswer](t, application.PreparedAnswers) {
		answers[answer.FieldName] = answer.Value
	}

	assert.Equal(t, "Alex", answers["first_name"])
	assert.Equal(t, "Candidate", answers["last_name"])
	assert.Equal(t, "alex@example.com", answers["email"])
	assert.Equal(t, "+1 555 0100", answers["phone"])
	assert.Equal(t, "https://www.linkedin.com/in/alex-candidate", answers["question_1"])
	assert.Equal(t, "1", answers["question_2"], "yes/no questions use the option's value")
	assert.Equal(
		t,
		[]applicationTypes.MissingAnswer{{Label: whyHereQuestion, FieldName: "question_3"}},
		decodeList[applicationTypes.MissingAnswer](t, application.MissingAnswers),
	)
	assert.Zero(
		t,
		fixture.stub.submissionCount(),
		"nothing is submitted without the employer's key",
	)

	path := candidatePath(fixture.candidateID) + "/applications/" + application.ID
	early := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		path,
		map[string]any{"status": "INTERVIEWING"},
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusConflict, early.Recorder.Code)

	submitted := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		path,
		map[string]any{"status": "SUBMITTED"},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, submitted.Recorder.Code, string(submitted.Body))
	helpers.AssertMatchesSpec(t, submitted)
	assert.NotNil(t, helpers.DecodeData[applicationModel.Application](t, submitted).SubmittedAt)

	interviewing := helpers.Do(t, fixture.router, http.MethodPatch, path,
		map[string]any{"status": "INTERVIEWING"}, bearer(fixture.userToken))
	assert.Equal(t, http.StatusOK, interviewing.Recorder.Code)

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(
			fixture.candidateID,
		)+"/applications?status=INTERVIEWING",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)

	details := helpers.DecodeData[[]applicationTypes.ApplicationDetail](t, listed)
	require.Len(t, details, 1)
	require.NotNil(t, details[0].Job)
	assert.Equal(t, "Senior Go Engineer", details[0].Job.JobTitle)
}

func TestApplicationIsSubmittedThroughTheEmployerAPIWhenEveryAnswerIsKnown(t *testing.T) {
	fixture := newApplicationFixture(t, map[string]string{"acme": "greenhouse-key"}, nil)
	fixture.saveProfile(
		t,
		map[string]string{whyHereQuestion: "Your mission matches my experience."},
	)

	application := fixture.apply(t, validJob())
	assert.Equal(t, modelHelpers.ApplicationStatusSubmitted, application.Status)
	assert.NotNil(t, application.SubmittedAt)

	require.Equal(t, 1, fixture.stub.submissionCount())
	submission := fixture.stub.submissions[0]
	assert.Equal(t, []string{"Alex"}, submission["first_name"])
	assert.Equal(t, []string{"alex@example.com"}, submission["email"])
	assert.Equal(t, []string{"1"}, submission["question_2"])
	assert.Equal(t, []string{"Your mission matches my experience."}, submission["question_3"])
	assert.Equal(t, []string{"resume.pdf"}, fixture.stub.resumeNames)

	duplicate := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		candidatePath(fixture.candidateID)+"/applications",
		map[string]any{"jobId": application.JobID},
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code, "a job is applied to once")
}

func TestLeverApplicationsSubmitWithTheSitesAPIKey(t *testing.T) {
	leverJob := withFields(validJob(), map[string]any{
		"atsType":     "LEVER",
		"externalUrl": "https://jobs.lever.co/acme/abc-123",
		"jobTitle":    "Backend Engineer",
	})

	withKey := newApplicationFixture(t, nil, map[string]string{"acme": "lever-key"})
	withKey.saveProfile(t, nil)

	submitted := withKey.apply(t, leverJob)
	assert.Equal(t, modelHelpers.ApplicationStatusSubmitted, submitted.Status)
	require.NotNil(t, submitted.ExternalConfirmation)
	assert.Equal(t, "lever-application-1", *submitted.ExternalConfirmation)
	require.Equal(t, 1, withKey.stub.submissionCount())
	assert.Equal(t, []string{"Alex Candidate"}, withKey.stub.submissions[0]["name"])
	assert.Equal(
		t,
		[]string{"https://www.linkedin.com/in/alex-candidate"},
		withKey.stub.submissions[0]["urls[LinkedIn]"],
	)
}

func TestUnsupportedFormsAndQueuedMatchesNeedTheCandidate(t *testing.T) {
	fixture := newApplicationFixture(t, nil, nil)

	other := fixture.apply(t, withFields(validJob(), map[string]any{
		"atsType":     "OTHER",
		"externalUrl": "https://careers.acme.com/jobs/1",
		"jobTitle":    "Platform Engineer",
	}))
	assert.Equal(t, modelHelpers.ApplicationStatusActionRequired, other.Status)
	assert.Contains(t, string(other.ExecutionLogs), "UNSUPPORTED_FORM")

	jobID := fixture.ingest(t, withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer"})).QueuedTasks[0].JobID

	var resumeID string

	ctx := context.Background()
	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		"SELECT id FROM candidate_resumes WHERE candidate_id = $1 AND is_active", fixture.candidateID,
	).
		Scan(&resumeID))
	_, err := fixture.deps.Postgres.GetPool().Exec(ctx, `
		INSERT INTO job_matches (candidate_id, job_id, resume_id, raw_similarity, match_score)
		VALUES ($1, $2, $3, 0.9, 0.9)
	`, fixture.candidateID, jobID, resumeID)
	require.NoError(t, err)

	queued, customErr := fixture.deps.Services.Application.QueueApplications(
		ctx,
		fixture.candidateID,
	)
	require.Nil(t, customErr)
	assert.Equal(t, 1, queued)

	again, customErr := fixture.deps.Services.Application.QueueApplications(
		ctx,
		fixture.candidateID,
	)
	require.Nil(t, customErr)
	assert.Zero(t, again, "queueing matches twice creates nothing new")

	profile := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		candidatePath(fixture.candidateID)+"/application-profile",
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, profile.Recorder.Code)
	helpers.AssertMatchesSpec(t, profile)
	assert.True(t, strings.Contains(string(profile.Body), `"data"`))
}

func TestAnAmbiguousSubmissionFailureIsNeverRetried(t *testing.T) {
	fixture := newApplicationFixture(t, map[string]string{"acme": "greenhouse-key"}, nil)
	fixture.saveProfile(
		t,
		map[string]string{whyHereQuestion: "Your mission matches my experience."},
	)
	fixture.stub.mutex.Lock()
	fixture.stub.failSubmits = true
	fixture.stub.mutex.Unlock()

	application := fixture.apply(t, validJob())

	assert.Equal(t, modelHelpers.ApplicationStatusActionRequired, application.Status,
		"the employer may already have it, so the candidate checks instead of a retry")
	require.NotNil(t, application.FailureReason)
	assert.Contains(t, *application.FailureReason, config.ErrSubmissionOutcomeUnknown.Error())

	require.Nil(
		t,
		fixture.deps.Services.Application.PrepareApplication(context.Background(), application.ID),
	)
	assert.Equal(t, 1, fixture.stub.submissionCount(), "the application reached the employer once")
}

func TestAnApplicationHeldByAnotherRunIsNotPreparedAgain(t *testing.T) {
	fixture := newApplicationFixture(t, map[string]string{"acme": "greenhouse-key"}, nil)
	fixture.saveProfile(
		t,
		map[string]string{whyHereQuestion: "Your mission matches my experience."},
	)

	ctx := context.Background()

	jobID := fixture.ingest(t, validJob()).QueuedTasks[0].JobID
	created := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		candidatePath(fixture.candidateID)+"/applications",
		map[string]any{"jobId": jobID},
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	applicationID := helpers.DecodeData[applicationModel.Application](t, created).ID

	now := time.Now().UTC()
	claimed, err := fixture.deps.Repositories.Application.Claim(
		ctx,
		applicationID,
		now.Add(-time.Minute),
		now,
	)
	require.NoError(t, err)
	require.True(t, claimed)

	require.Nil(t, fixture.deps.Services.Application.PrepareApplication(ctx, applicationID))

	assert.Zero(
		t,
		fixture.stub.submissionCount(),
		"a second run does nothing while the first holds the claim",
	)
	assert.Equal(
		t,
		modelHelpers.ApplicationStatusInProgress,
		fixture.application(t, applicationID).Status,
	)
}
