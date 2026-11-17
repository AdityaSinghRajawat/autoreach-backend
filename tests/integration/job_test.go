//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerServer "github.com/AdityaSinghRajawat/autoreach-backend/worker/server"
)

const (
	ingestPath     = "/api/v1/jobs/ingest"
	apiClientsPath = "/api/v1/admin/api-clients"
	adminEmail     = "admin@example.com"
	adminPassword  = "admin-password-123"
)

type jobFixture struct {
	deps        *dependencies.Dependencies
	router      http.Handler
	apiKey      string
	userToken   string
	candidateID string
}

func newJobFixture(t *testing.T) jobFixture {
	t.Helper()
	deps := setup.Dependencies(t)
	router := routes.NewRoutes(deps)
	created := createAPIClient(t, deps, router, []string{"job:ingest:any"})
	user := register(t, router, "alex@example.com")

	return jobFixture{
		deps:        deps,
		router:      router,
		apiKey:      created.APIKey,
		userToken:   user.AccessToken,
		candidateID: *user.User.CandidateID,
	}
}

func adminToken(t *testing.T, deps *dependencies.Dependencies, router http.Handler) string {
	t.Helper()

	_, customErr := deps.Services.Auth.CreateAdmin(context.Background(), adminEmail, adminPassword)
	require.Nil(t, customErr)

	return loginAdmin(t, router)
}

func loginAdmin(t *testing.T, router http.Handler) string {
	t.Helper()

	exchange := helpers.Do(
		t,
		router,
		http.MethodPost,
		loginPath,
		authTypes.LoginRequest{Email: adminEmail, Password: adminPassword},
		nil,
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))

	return helpers.DecodeData[authTypes.SessionResponse](t, exchange).AccessToken
}

func createAPIClient(
	t *testing.T,
	deps *dependencies.Dependencies,
	router http.Handler,
	scopes []string,
) authTypes.CreateAPIClientResponse {
	t.Helper()
	exchange := helpers.Do(
		t,
		router,
		http.MethodPost,
		apiClientsPath,
		authTypes.CreateAPIClientRequest{
			Name:   "linkedin-scraper",
			Scopes: scopes,
		},
		bearer(adminToken(t, deps, router)),
	)
	require.Equal(t, http.StatusCreated, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[authTypes.CreateAPIClientResponse](t, exchange)
}

func apiKeyHeader(apiKey string) map[string]string {
	return map[string]string{config.GetAPIKeyHeader(): apiKey}
}

func validJob() map[string]any {
	return map[string]any{
		"source":         "LINKEDIN",
		"companyName":    "Acme Corp",
		"companyDomain":  "acme.com",
		"jobTitle":       "Senior Go Engineer",
		"jobDescription": "We are seeking a Go specialist...",
		"location":       "Remote - US",
		"atsType":        "GREENHOUSE",
		"externalUrl":    "https://boards.greenhouse.io/acme/jobs/10293",
	}
}

func withFields(base, overrides map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(overrides))
	for key, value := range base {
		merged[key] = value
	}

	for key, value := range overrides {
		merged[key] = value
	}

	return merged
}

func (f jobFixture) ingest(t *testing.T, jobs ...map[string]any) jobTypes.IngestJobsResponse {
	t.Helper()
	exchange := helpers.Do(
		t,
		f.router,
		http.MethodPost,
		ingestPath,
		map[string]any{"jobs": jobs},
		apiKeyHeader(f.apiKey),
	)
	require.Equal(t, http.StatusAccepted, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[jobTypes.IngestJobsResponse](t, exchange)
}

func (f jobFixture) getJob(t *testing.T, jobID string) jobModel.JobPosting {
	t.Helper()
	exchange := helpers.Do(
		t,
		f.router,
		http.MethodGet,
		"/api/v1/jobs/"+jobID,
		nil,
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[jobModel.JobPosting](t, exchange)
}

func TestIngestStoresTheJobAndQueuesAnEmbeddingTask(t *testing.T) {
	fixture := newJobFixture(t)

	result := fixture.ingest(t, withFields(validJob(), map[string]any{
		"department":     "Engineering",
		"salaryMin":      150000,
		"salaryMax":      190000,
		"salaryCurrency": "USD",
	}))

	assert.Equal(t, 1, result.IngestedCount)
	require.Len(t, result.QueuedTasks, 1)
	assert.Equal(t, "task:embed_job", result.QueuedTasks[0].TaskName)
	assert.Equal(t, workerConfigs.PublishStatusEnqueued, result.QueuedTasks[0].Status)

	job := fixture.getJob(t, result.QueuedTasks[0].JobID)
	assert.Equal(t, "LINKEDIN", job.Source)
	assert.Equal(t, "Acme Corp", job.CompanyName)
	assert.Equal(t, "acme.com", job.CompanyDomain)
	assert.Equal(t, "Senior Go Engineer", job.JobTitle)
	require.NotNil(t, job.Department)
	assert.Equal(t, "Engineering", *job.Department)
	require.NotNil(t, job.SalaryMax)
	assert.InDelta(t, 190000, *job.SalaryMax, 0.001)
	assert.Equal(t, "GREENHOUSE", job.ATSType)
	assert.EqualValues(t, "ACTIVE", job.Status)
	assert.False(t, job.IsEmbedded)
}

func TestIngestDeduplicatesAcrossSourcesByCompanyDomainAndNormalisedTitle(t *testing.T) {
	fixture := newJobFixture(t)

	first := fixture.ingest(t, validJob())
	second := fixture.ingest(t, withFields(validJob(), map[string]any{
		"source":        "INDEED",
		"companyDomain": "WWW.Acme.com",
		"jobTitle":      "senior go engineer (Remote)",
	}))
	sameBatch := fixture.ingest(t,
		withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer"}),
		withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer [Hybrid]"}),
	)

	require.Len(t, first.QueuedTasks, 1)
	require.Len(t, second.QueuedTasks, 1)
	assert.Equal(t, first.QueuedTasks[0].JobID, second.QueuedTasks[0].JobID)
	assert.Equal(t, 2, sameBatch.IngestedCount)
	require.Len(
		t,
		sameBatch.QueuedTasks,
		1,
		"two postings that hash to the same job produce one task",
	)
	assert.NotEqual(t, first.QueuedTasks[0].JobID, sameBatch.QueuedTasks[0].JobID)

	var stored int
	require.NoError(
		t,
		fixture.deps.Postgres.GetPool().
			QueryRow(context.Background(), `SELECT COUNT(*) FROM job_postings`).
			Scan(&stored),
	)
	assert.Equal(t, 2, stored)
}

func TestIngestRejectsTheWholeRequestAndNamesEveryInvalidField(t *testing.T) {
	fixture := newJobFixture(t)

	exchange := helpers.Do(t, fixture.router, http.MethodPost, ingestPath, map[string]any{
		"jobs": []any{
			validJob(),
			withFields(
				validJob(),
				map[string]any{
					"source":        "MONSTER",
					"companyDomain": "not a domain",
					"jobTitle":      "",
				},
			),
			withFields(
				validJob(),
				map[string]any{
					"jobTitle":    "Platform Engineer",
					"externalUrl": "https://evil.example.com/acme/jobs/1",
				},
			),
			withFields(
				validJob(),
				map[string]any{"jobTitle": "Data Engineer", "salaryMin": 200, "salaryMax": 100},
			),
			withFields(
				validJob(),
				map[string]any{
					"jobTitle":    "SRE",
					"atsType":     "OTHER",
					"externalUrl": "https://169.254.169.254/latest",
				},
			),
			withFields(validJob(), map[string]any{"jobTitle": "(Remote)"}),
		},
	}, apiKeyHeader(fixture.apiKey))

	require.Equal(t, http.StatusUnprocessableEntity, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)
	assert.ElementsMatch(t, []config.ErrorDetail{
		{
			Field: "jobs[1].source",
			Issue: "must be one of: LINKEDIN INDEED GLASSDOOR ZIPRECRUITER NAUKRI WELLFOUND GREENHOUSE LEVER WORKDAY ASHBY SMARTRECRUITERS ADZUNA DIRECT",
		},
		{Field: "jobs[1].companyDomain", Issue: "must be a valid domain name"},
		{Field: "jobs[1].jobTitle", Issue: "is required"},
		{
			Field: "jobs[2].externalUrl",
			Issue: "must be a public https URL on a host that matches atsType",
		},
		{Field: "jobs[3].salaryMax", Issue: "must be greater than or equal to salaryMin"},
		{
			Field: "jobs[4].externalUrl",
			Issue: "must be a public https URL on a host that matches atsType",
		},
		{Field: "jobs[5].jobTitle", Issue: "must contain letters or digits"},
	}, helpers.DecodeEnvelope(t, exchange).Error.Details)

	var stored int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM job_postings`,
	).Scan(&stored))
	assert.Zero(t, stored, "an invalid request must not store any of its postings")
}

func TestIngestRejectsAnEmptyOrOversizedBatch(t *testing.T) {
	fixture := newJobFixture(t)

	oversized := make([]map[string]any, config.GetMaxIngestBatchSize()+1)
	for index := range oversized {
		oversized[index] = validJob()
	}

	empty := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		ingestPath,
		map[string]any{"jobs": []any{}},
		apiKeyHeader(fixture.apiKey),
	)
	tooMany := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		ingestPath,
		map[string]any{"jobs": oversized},
		apiKeyHeader(fixture.apiKey),
	)

	assert.Equal(t, http.StatusUnprocessableEntity, empty.Recorder.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, tooMany.Recorder.Code)
	helpers.AssertMatchesSpec(t, empty)
}

func TestIngestRequiresAValidUnrevokedAPIKeyWithTheIngestScope(t *testing.T) {
	deps := setup.Dependencies(t)
	router := routes.NewRoutes(deps)
	created := createAPIClient(t, deps, router, []string{"job:ingest:any"})
	user := register(t, router, "alex@example.com")
	payload := map[string]any{"jobs": []any{validJob()}}
	admin := helpers.Do(
		t,
		router,
		http.MethodPost,
		loginPath,
		authTypes.LoginRequest{Email: adminEmail, Password: adminPassword},
		nil,
	)
	adminHeaders := bearer(helpers.DecodeData[authTypes.SessionResponse](t, admin).AccessToken)

	missing := helpers.Do(t, router, http.MethodPost, ingestPath, payload, nil)
	forged := helpers.Do(
		t,
		router,
		http.MethodPost,
		ingestPath,
		payload,
		apiKeyHeader(created.Client.KeyPrefix+"_forged-secret"),
	)
	garbage := helpers.Do(t, router, http.MethodPost, ingestPath, payload, apiKeyHeader("garbage"))
	userJWT := helpers.Do(t, router, http.MethodPost, ingestPath, payload, bearer(user.AccessToken))
	valid := helpers.Do(
		t,
		router,
		http.MethodPost,
		ingestPath,
		payload,
		apiKeyHeader(created.APIKey),
	)

	assert.Equal(t, http.StatusUnauthorized, missing.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, forged.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, garbage.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, userJWT.Recorder.Code)
	assert.Equal(t, http.StatusAccepted, valid.Recorder.Code)
	helpers.AssertMatchesSpec(t, missing)

	utils.DrainBackgroundGoroutines()

	listed := helpers.Do(t, router, http.MethodGet, apiClientsPath, nil, adminHeaders)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
	clients := helpers.DecodeData[[]authModel.APIClient](t, listed)
	require.Len(t, clients, 1)
	assert.NotNil(t, clients[0].LastUsedAt)
	assert.NotContains(t, string(listed.Body), created.APIKey)

	revoke := helpers.Do(
		t,
		router,
		http.MethodDelete,
		apiClientsPath+"/"+created.Client.ID,
		nil,
		adminHeaders,
	)
	require.Equal(t, http.StatusNoContent, revoke.Recorder.Code)
	helpers.AssertMatchesSpec(t, revoke)
	revokeAgain := helpers.Do(
		t,
		router,
		http.MethodDelete,
		apiClientsPath+"/"+created.Client.ID,
		nil,
		adminHeaders,
	)
	assert.Equal(t, http.StatusNotFound, revokeAgain.Recorder.Code)

	afterRevoke := helpers.Do(
		t,
		router,
		http.MethodPost,
		ingestPath,
		payload,
		apiKeyHeader(created.APIKey),
	)
	assert.Equal(t, http.StatusUnauthorized, afterRevoke.Recorder.Code)
}

func TestAdminRoutesRejectNonAdmins(t *testing.T) {
	router := newRouter(t)
	user := register(t, router, "alex@example.com")
	request := authTypes.CreateAPIClientRequest{
		Name:   "x",
		Scopes: []string{"job:ingest:any"},
	}

	forbidden := helpers.Do(
		t,
		router,
		http.MethodPost,
		apiClientsPath,
		request,
		bearer(user.AccessToken),
	)
	anonymous := helpers.Do(t, router, http.MethodGet, apiClientsPath, nil, nil)

	assert.Equal(t, http.StatusForbidden, forbidden.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
	helpers.AssertMatchesSpec(t, forbidden)
}

func TestEmbedJobsStoresVectorsAndIsIdempotent(t *testing.T) {
	fixture := newJobFixture(t)
	embedder := setup.Embedder(t)
	ingested := fixture.ingest(
		t,
		validJob(),
		withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer"}),
	)
	jobIDs := []string{ingested.QueuedTasks[0].JobID, ingested.QueuedTasks[1].JobID}

	first, customErr := fixture.deps.Services.Job.EmbedJobs(context.Background(), jobIDs)
	require.Nil(t, customErr)
	second, customErr := fixture.deps.Services.Job.EmbedJobs(context.Background(), jobIDs)
	require.Nil(t, customErr)

	assert.Equal(t, 2, first.Embedded)
	assert.Equal(t, 0, second.Embedded)
	assert.Equal(
		t,
		1,
		embedder.CallCount(),
		"already embedded jobs must not be sent to the embedder again",
	)
	assert.True(t, fixture.getJob(t, jobIDs[0]).IsEmbedded)

	var (
		dimensions int
		model      string
	)
	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(),
		`SELECT vector_dims(embedding), embedding_model FROM job_postings WHERE id = $1`, jobIDs[0],
	).Scan(&dimensions, &model))
	assert.Equal(t, config.GetAIEmbeddingDimensions(), dimensions)
	assert.Equal(t, mocks.EmbedderModel, model)
}

func TestReingestingAChangedDescriptionClearsTheEmbeddingAndRequeues(t *testing.T) {
	fixture := newJobFixture(t)
	first := fixture.ingest(t, validJob())
	jobID := first.QueuedTasks[0].JobID
	_, customErr := fixture.deps.Services.Job.EmbedJobs(context.Background(), []string{jobID})
	require.Nil(t, customErr)

	unchanged := fixture.ingest(t, validJob())
	changed := fixture.ingest(
		t,
		withFields(validJob(), map[string]any{"jobDescription": "A rewritten description."}),
	)

	assert.Equal(t, 1, unchanged.IngestedCount)
	assert.Empty(
		t,
		unchanged.QueuedTasks,
		"an embedded job with the same description needs no task",
	)
	require.Len(t, changed.QueuedTasks, 1)
	assert.Equal(t, jobID, changed.QueuedTasks[0].JobID)
	job := fixture.getJob(t, jobID)
	assert.False(t, job.IsEmbedded)
	assert.Equal(t, "A rewritten description.", job.JobDescription)
}

func TestEmbedFailuresAreClassifiedForRetry(t *testing.T) {
	fixture := newJobFixture(t)
	embedder := setup.Embedder(t)
	ingested := fixture.ingest(t, validJob())
	jobIDs := []string{ingested.QueuedTasks[0].JobID}

	embedder.SetErr(&utils.APIError{StatusCode: http.StatusTooManyRequests})

	_, transient := fixture.deps.Services.Job.EmbedJobs(context.Background(), jobIDs)

	embedder.SetErr(&utils.APIError{StatusCode: http.StatusUnauthorized})

	_, rejected := fixture.deps.Services.Job.EmbedJobs(context.Background(), jobIDs)

	embedder.SetErr(errors.New("dial tcp: connection refused"))

	_, network := fixture.deps.Services.Job.EmbedJobs(context.Background(), jobIDs)

	require.NotNil(t, transient)
	require.NotNil(t, rejected)
	require.NotNil(t, network)
	assert.True(t, config.IsRetryableErrCode(transient.ErrCode))
	assert.False(t, config.IsRetryableErrCode(rejected.ErrCode))
	assert.True(t, config.IsRetryableErrCode(network.ErrCode))
	assert.False(t, fixture.getJob(t, jobIDs[0]).IsEmbedded)
}

func TestIngestSurvivesAQueueOutageAndTheReconcilerRequeuesLater(t *testing.T) {
	deps := setup.Dependencies(t)
	workerClient := mocks.NewWorkerClient()
	service, err := jobService.NewJobService(jobService.JobServiceDependencies{
		JobPostingRepo:     deps.Repositories.JobPosting,
		TransactionManager: deps.TransactionManager,
		WorkerClient:       workerClient,
		Embedder:           setup.Embedder(t),
		JobSourceRepo:      deps.Repositories.JobSource,
		RateLimiter:        deps.RateLimiter,
		DomainResolver:     deps.Services.Enrichment,
	})
	require.NoError(t, err)

	request := jobTypes.IngestJobsRequest{Jobs: []jobTypes.IngestJobItem{{
		Source:         "DIRECT",
		CompanyName:    "Acme Corp",
		CompanyDomain:  "acme.com",
		JobTitle:       "Senior Go Engineer",
		JobDescription: "We are seeking a Go specialist...",
		ATSType:        "OTHER",
		ExternalURL:    "https://careers.acme.com/jobs/1",
	}}}

	workerClient.SetErr(errors.New("redis unavailable"))

	ingested, customErr := service.IngestJobs(context.Background(), request.ToJobPostingModels())
	require.Nil(t, customErr, "a queue outage must not fail an ingest that already committed")
	require.Len(t, ingested.QueuedTasks, 1)
	assert.Equal(t, workerConfigs.PublishStatusDeferred, ingested.QueuedTasks[0].Status)

	workerClient.SetErr(nil)
	time.Sleep(20 * time.Millisecond)

	requeued, customErr := service.RequeuePendingEmbeddings(context.Background())
	require.Nil(t, customErr)

	assert.Equal(t, 1, requeued)

	published := workerClient.Snapshot()
	require.Len(t, published, 1)
	assert.Equal(t, workerConfigs.TaskEmbedJob, published[0].TaskID)
}

func TestWorkerConsumesAggregatesAndEmbedsIngestedJobsEndToEnd(t *testing.T) {
	fixture := newJobFixture(t)
	embedder := setup.Embedder(t)
	worker, err := workerServer.NewWorkerServer(context.Background(), fixture.deps)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	stopped := make(chan error, 1)
	go func() { stopped <- worker.Start(ctx) }()

	defer func() {
		cancel()

		select {
		case err := <-stopped:
			assert.NoError(t, err)
		case <-time.After(30 * time.Second):
			t.Error("worker did not stop")
		}
	}()

	ingested := fixture.ingest(t,
		validJob(),
		withFields(validJob(), map[string]any{"jobTitle": "Staff Go Engineer"}),
		withFields(validJob(), map[string]any{"jobTitle": "Principal Go Engineer"}),
	)
	require.Len(t, ingested.QueuedTasks, 3)

	require.Eventually(t, func() bool {
		for _, task := range ingested.QueuedTasks {
			if !fixture.getJob(t, task.JobID).IsEmbedded {
				return false
			}
		}

		return true
	}, 30*time.Second, 200*time.Millisecond, "jobs were not embedded by the worker")

	assert.Equal(t, 3, embedder.InputCount())
	assert.Less(
		t,
		embedder.CallCount(),
		3,
		"embedding tasks should be batched into fewer provider calls",
	)

	metrics := string(helpers.Do(t, fixture.router, http.MethodGet, "/metrics", nil, nil).Body)
	assert.Contains(
		t,
		metrics,
		`autoreach_worker_tasks_total{outcome="success",stage="EMBEDDING",task="task:embed_job_batch"}`,
	)
	assert.Contains(t, metrics, `autoreach_queue_tasks{queue="embedding_queue",state="dead"}`)
}
