//go:build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const jobSourcesPath = "/api/v1/admin/job-sources"

const greenhouseBoardBothJobs = `{"jobs":[
	{"id":1,"title":"Senior Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/1",
	 "content":"&lt;p&gt;Build &lt;strong&gt;APIs&lt;/strong&gt; in Go.&lt;/p&gt;&lt;ul&gt;&lt;li&gt;Postgres&lt;/li&gt;&lt;/ul&gt;",
	 "updated_at":"2026-11-01T10:00:00-05:00","location":{"name":"Remote"},"departments":[{"name":"Engineering"}]},
	{"id":2,"title":"Staff Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/2",
	 "content":"&lt;p&gt;Lead the platform team.&lt;/p&gt;","location":{"name":"Berlin"},"departments":[]},
	{"id":3,"title":"","absolute_url":"https://boards.greenhouse.io/acme/jobs/3","content":"untitled",
	 "location":{"name":""},"departments":[]}
]}`

const greenhouseBoardOneJob = `{"jobs":[
	{"id":1,"title":"Senior Go Engineer","absolute_url":"https://boards.greenhouse.io/acme/jobs/1",
	 "content":"&lt;p&gt;Build &lt;strong&gt;APIs&lt;/strong&gt; in Go.&lt;/p&gt;","location":{"name":"Remote"},"departments":[]}
]}`

const leverPostings = `[{"id":"abc-123","text":"Backend Engineer","hostedUrl":"https://jobs.lever.co/globex/abc-123",
	"descriptionPlain":"Join Globex.","additionalPlain":"Equal opportunity employer.","createdAt":1762000000000,
	"lists":[{"text":"Requirements","content":"<li>Go</li><li>Kubernetes</li>"}],
	"categories":{"location":"London","team":"Platform"}}]`

type boardsStub struct {
	server       *httptest.Server
	jobRemoved   atomic.Bool
	boardFailing atomic.Bool
}

func newBoardsStub(t *testing.T) *boardsStub {
	t.Helper()

	stub := &boardsStub{}
	stub.server = httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")

			switch {
			case stub.boardFailing.Load():
				writer.WriteHeader(http.StatusNotFound)
			case request.URL.Path == "/greenhouse/boards/acme/jobs" && request.URL.Query().Get("content") == "true":
				if stub.jobRemoved.Load() {
					_, _ = io.WriteString(writer, greenhouseBoardOneJob)

					return
				}

				_, _ = io.WriteString(writer, greenhouseBoardBothJobs)
			case request.URL.Path == "/lever-eu/postings/globex" && request.URL.Query().Get("mode") == "json":
				_, _ = io.WriteString(writer, leverPostings)
			default:
				writer.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(stub.server.Close)

	return stub
}

type jobSourceFixture struct {
	jobFixture
	stub       *boardsStub
	adminToken string
}

func newJobSourceFixture(t *testing.T) jobSourceFixture {
	t.Helper()

	fixture := newJobFixture(t)
	stub := newBoardsStub(t)

	fixture.deps.JobBoards = jobBoards.NewJobBoardRegistry(jobBoards.RegistryOptions{
		GreenhouseAPIURL: stub.server.URL + "/greenhouse",
		LeverAPIURL:      stub.server.URL + "/lever",
		LeverEUAPIURL:    stub.server.URL + "/lever-eu",
	})
	require.NoError(t, fixture.deps.Wire())
	fixture.router = routes.NewRoutes(fixture.deps)

	t.Cleanup(func() {
		fixture.deps.JobBoards = nil
		require.NoError(t, fixture.deps.Wire())
	})

	return jobSourceFixture{
		jobFixture: fixture,
		stub:       stub,
		adminToken: loginAdmin(t, fixture.router),
	}
}

func (f jobSourceFixture) createSource(t *testing.T, body map[string]any) jobModel.JobSource {
	t.Helper()

	created := helpers.Do(t, f.router, http.MethodPost, jobSourcesPath, body, bearer(f.adminToken))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	return helpers.DecodeData[jobModel.JobSource](t, created)
}

func (f jobSourceFixture) poll(t *testing.T, sourceID string) helpers.Exchange {
	t.Helper()

	polled := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		jobSourcesPath+"/"+sourceID+"/poll",
		nil,
		bearer(f.adminToken),
	)
	helpers.AssertMatchesSpec(t, polled)

	return polled
}

func (f jobSourceFixture) sourceJobs(t *testing.T, sourceID string) map[string]jobModel.JobPosting {
	t.Helper()

	rows, err := f.deps.Postgres.GetPool().Query(context.Background(), `
		SELECT external_job_id, job_title, job_description, COALESCE(location, ''), COALESCE(department, ''),
			source::text, status::text
		FROM job_postings WHERE job_source_id = $1
	`, sourceID)
	require.NoError(t, err)

	defer rows.Close()

	jobs := map[string]jobModel.JobPosting{}

	for rows.Next() {
		var (
			job                              jobModel.JobPosting
			externalID, location, department string
		)
		require.NoError(t, rows.Scan(
			&externalID,
			&job.JobTitle,
			&job.JobDescription,
			&location,
			&department,
			&job.Source,
			&job.Status,
		))
		job.Location, job.Department = &location, &department
		jobs[externalID] = job
	}

	require.NoError(t, rows.Err())

	return jobs
}

func TestGreenhouseBoardPollIngestsJobsAndClosesRemovedOnes(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"atsType":       "GREENHOUSE",
		"boardToken":    "Acme",
		"companyName":   "Acme Corp",
		"companyDomain": "acme.com",
	})
	assert.Equal(t, "acme", source.BoardToken)

	duplicate := helpers.Do(t, fixture.router, http.MethodPost, jobSourcesPath, map[string]any{
		"atsType":       "GREENHOUSE",
		"boardToken":    "acme",
		"companyName":   "Acme",
		"companyDomain": "acme.com",
	}, bearer(fixture.adminToken))
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code)

	first := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, first.Recorder.Code, string(first.Body))
	firstResult := helpers.DecodeData[jobTypes.PollJobSourceResult](t, first)
	assert.Equal(
		t,
		jobTypes.PollJobSourceResult{SourceID: source.ID, Fetched: 3, Ingested: 2, Skipped: 1},
		firstResult,
	)

	jobs := fixture.sourceJobs(t, source.ID)
	require.Len(t, jobs, 2)
	assert.Equal(t, "Build APIs in Go.\nPostgres", jobs["1"].JobDescription)
	assert.Equal(t, "Engineering", *jobs["1"].Department)
	assert.Equal(t, "Berlin", *jobs["2"].Location)
	assert.Equal(t, "GREENHOUSE", jobs["1"].Source)

	fixture.stub.jobRemoved.Store(true)
	second := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, second.Recorder.Code, string(second.Body))
	assert.Equal(t, 1, helpers.DecodeData[jobTypes.PollJobSourceResult](t, second).Closed)

	jobs = fixture.sourceJobs(t, source.ID)
	assert.Equal(t, "ACTIVE", string(jobs["1"].Status))
	assert.Equal(t, "CLOSED", string(jobs["2"].Status))

	fixture.stub.boardFailing.Store(true)
	failed := fixture.poll(t, source.ID)
	assert.NotEqual(t, http.StatusOK, failed.Recorder.Code)
	assert.Equal(t, "ACTIVE", string(fixture.sourceJobs(t, source.ID)["1"].Status),
		"a failed fetch must not close any job")

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		jobSourcesPath,
		nil,
		bearer(fixture.adminToken),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
	sources := helpers.DecodeData[[]jobModel.JobSource](t, listed)
	require.Len(t, sources, 1)
	require.NotNil(t, sources[0].LastJobCount)
	assert.Equal(t, 1, *sources[0].LastJobCount)
	require.NotNil(t, sources[0].LastError)
}

func TestLeverEUBoardPollMapsPostings(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"atsType":       "LEVER",
		"boardToken":    "globex",
		"region":        "EU",
		"companyName":   "Globex",
		"companyDomain": "globex.com",
	})

	polled := fixture.poll(t, source.ID)
	require.Equal(t, http.StatusOK, polled.Recorder.Code, string(polled.Body))

	job := fixture.sourceJobs(t, source.ID)["abc-123"]
	assert.Equal(t, "Backend Engineer", job.JobTitle)
	assert.Equal(
		t,
		"Join Globex.\n\nRequirements\nGo\nKubernetes\n\nEqual opportunity employer.",
		job.JobDescription,
	)
	assert.Equal(t, "London", *job.Location)
	assert.Equal(t, "Platform", *job.Department)
}

func TestJobSourcesCanBePausedAndDeleted(t *testing.T) {
	fixture := newJobSourceFixture(t)
	source := fixture.createSource(t, map[string]any{
		"atsType":       "GREENHOUSE",
		"boardToken":    "acme",
		"companyName":   "Acme",
		"companyDomain": "acme.com",
	})
	sourcePath := jobSourcesPath + "/" + source.ID

	paused := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		sourcePath,
		map[string]any{"isActive": false},
		bearer(fixture.adminToken),
	)
	require.Equal(t, http.StatusOK, paused.Recorder.Code, string(paused.Body))
	helpers.AssertMatchesSpec(t, paused)
	assert.False(t, helpers.DecodeData[jobModel.JobSource](t, paused).IsActive)

	scheduled, customErr := fixture.deps.Services.Job.ScheduleJobSourcePolls(context.Background())
	require.Nil(t, customErr)
	assert.Zero(t, scheduled, "paused sources are not polled")

	empty := helpers.Do(
		t,
		fixture.router,
		http.MethodPatch,
		sourcePath,
		map[string]any{},
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusBadRequest, empty.Recorder.Code)

	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		sourcePath,
		nil,
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusNoContent, deleted.Recorder.Code)

	missing := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		sourcePath,
		nil,
		bearer(fixture.adminToken),
	)
	assert.Equal(t, http.StatusNotFound, missing.Recorder.Code)
	helpers.AssertMatchesSpec(t, missing)
}

func TestJobSourceRoutesRejectNonAdmins(t *testing.T) {
	fixture := newJobFixture(t)

	forbidden := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		jobSourcesPath,
		nil,
		bearer(fixture.userToken),
	)
	anonymous := helpers.Do(t, fixture.router, http.MethodGet, jobSourcesPath, nil, nil)

	assert.Equal(t, http.StatusForbidden, forbidden.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
}

func TestNaukriAndWellfoundJobsCanBeIngested(t *testing.T) {
	fixture := newJobFixture(t)

	ingested := fixture.ingest(t,
		withFields(validJob(), map[string]any{
			"source":      "NAUKRI",
			"atsType":     "OTHER",
			"externalUrl": "https://www.naukri.com/job-listings-go-1",
		}),
		withFields(validJob(), map[string]any{
			"source":      "WELLFOUND",
			"atsType":     "OTHER",
			"externalUrl": "https://wellfound.com/jobs/2-go",
		}),
	)

	assert.Equal(t, 2, ingested.IngestedCount)
}
