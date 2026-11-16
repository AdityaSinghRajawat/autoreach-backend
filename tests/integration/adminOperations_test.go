//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	adminTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/adminManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

const adminBasePath = "/api/v1/admin"

func testInspector(t *testing.T) *asynq.Inspector {
	t.Helper()

	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     config.GetRedisAddr(),
		Password: config.GetRedisPassword(),
		DB:       config.GetRedisDB(),
	})

	t.Cleanup(func() { _ = inspector.Close() })

	return inspector
}

func TestDeadTaskCanBeListedAndRetriedFromTheAPI(t *testing.T) {
	fixture := newJobFixture(t)
	token := loginAdmin(t, fixture.router)
	inspector := testInspector(t)
	sourceID := uuid.NewString()

	published, err := fixture.deps.WorkerClient.Publish(
		context.Background(),
		workerConfigs.PublishParams{
			TaskID:  workerConfigs.TaskPollJobSource,
			Payload: workerCustomTypes.PollJobSourcePayload{SourceID: sourceID},
		},
	)
	require.NoError(t, err)
	require.NoError(t, inspector.ArchiveTask(published.QueueName, published.MessageID))

	t.Cleanup(func() { _ = inspector.DeleteTask(published.QueueName, published.MessageID) })

	queues := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		adminBasePath+"/queues",
		nil,
		bearer(token),
	)
	require.Equal(t, http.StatusOK, queues.Recorder.Code, string(queues.Body))
	helpers.AssertMatchesSpec(t, queues)

	summaries := helpers.DecodeData[[]adminTypes.QueueSummary](t, queues)
	require.Len(t, summaries, len(workerConfigs.GetQueueDefinitions()))

	for _, summary := range summaries {
		if summary.Queue == published.QueueName {
			assert.Equal(t, 1, summary.Dead)
		}
	}

	deadPath := adminBasePath + "/queues/" + published.QueueName + "/dead"
	dead := helpers.Do(t, fixture.router, http.MethodGet, deadPath, nil, bearer(token))
	require.Equal(t, http.StatusOK, dead.Recorder.Code, string(dead.Body))
	helpers.AssertMatchesSpec(t, dead)

	tasks := helpers.DecodeData[[]adminTypes.DeadTask](t, dead)
	require.Len(t, tasks, 1)
	assert.Equal(t, published.MessageID, tasks[0].TaskID)
	assert.Equal(t, string(workerConfigs.TaskPollJobSource), tasks[0].TaskType)
	assert.JSONEq(t, `{"sourceId":"`+sourceID+`"}`, string(tasks[0].Payload))

	retried := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		deadPath+"/"+published.MessageID+"/retry",
		nil,
		bearer(token),
	)
	require.Equal(t, http.StatusNoContent, retried.Recorder.Code, string(retried.Body))

	info, err := inspector.GetTaskInfo(published.QueueName, published.MessageID)
	require.NoError(t, err)
	assert.Equal(t, asynq.TaskStatePending, info.State, "a retried task waits for a worker again")

	again := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		deadPath+"/"+published.MessageID+"/retry",
		nil,
		bearer(token),
	)
	assert.Equal(t, http.StatusNotFound, again.Recorder.Code, "only dead tasks can be retried")
	helpers.AssertMatchesSpec(t, again)

	unknownQueue := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		adminBasePath+"/queues/nope/dead",
		nil,
		bearer(token),
	)
	assert.Equal(t, http.StatusNotFound, unknownQueue.Recorder.Code)
}

func TestProviderStatusReportsCreditsAndDeliverability(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	ctx := context.Background()

	_, err := fixture.deps.RedisUtil.IncrementDaily(
		ctx,
		config.GetEnrichmentCreditNamespace(),
		config.ProviderApollo,
		7,
	)
	require.NoError(t, err)
	_, err = fixture.deps.Postgres.GetPool().Exec(ctx,
		`UPDATE outreach_messages SET status = 'BOUNCED' WHERE channel = 'EMAIL' AND status = 'SENT'`)
	require.NoError(t, err)

	exchange := helpers.Do(t, fixture.router, http.MethodGet, adminBasePath+"/providers", nil,
		bearer(loginAdmin(t, fixture.router)))
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	providers := helpers.DecodeData[adminTypes.ProvidersResponse](t, exchange)
	require.Len(t, providers.Enrichment, len(config.GetEnrichmentProviderPolicies()))

	for _, provider := range providers.Enrichment {
		if provider.Code == config.ProviderApollo {
			assert.Equal(t, int64(7), provider.CreditsUsedToday)
		}
	}

	assert.Equal(t, 2, providers.Email.Sent)
	assert.Equal(t, 1, providers.Email.Bounced)
	require.NotNil(t, providers.Email.DeliverabilityPercentage)
	assert.InDelta(t, 50.0, *providers.Email.DeliverabilityPercentage, 0.001)
}

func TestSuppressionsCanBeManagedByAdmins(t *testing.T) {
	fixture := newJobFixture(t)
	token := loginAdmin(t, fixture.router)
	suppressionsPath := adminBasePath + "/suppressions"

	created := helpers.Do(t, fixture.router, http.MethodPost, suppressionsPath,
		map[string]any{"entryType": "EMAIL", "value": " Hiring@Acme.com "}, bearer(token))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)

	entry := helpers.DecodeData[outreachModel.SuppressionEntry](t, created)
	assert.Equal(t, "hiring@acme.com", entry.Value)
	assert.Equal(t, "MANUAL", string(entry.Reason))

	duplicate := helpers.Do(t, fixture.router, http.MethodPost, suppressionsPath,
		map[string]any{"entryType": "EMAIL", "value": "hiring@acme.com"}, bearer(token))
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code)

	invalid := helpers.Do(t, fixture.router, http.MethodPost, suppressionsPath,
		map[string]any{"entryType": "DOMAIN", "value": "hiring@acme.com"}, bearer(token))
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Recorder.Code)
	helpers.AssertMatchesSpec(t, invalid)

	domain := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		suppressionsPath,
		map[string]any{
			"entryType":   "DOMAIN",
			"value":       "globex.com",
			"candidateId": fixture.candidateID,
		},
		bearer(token),
	)
	require.Equal(t, http.StatusCreated, domain.Recorder.Code, string(domain.Body))

	suppressed, err := fixture.deps.Repositories.Suppression.IsSuppressed(
		context.Background(), fixture.candidateID, "ceo@globex.com")
	require.NoError(t, err)
	assert.True(t, suppressed, "a domain entry suppresses every address at that domain")

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		suppressionsPath+"?q=acme",
		nil,
		bearer(token),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code, string(listed.Body))
	helpers.AssertMatchesSpec(t, listed)

	entries := helpers.DecodeData[[]outreachModel.SuppressionEntry](t, listed)
	require.Len(t, entries, 1)
	assert.Equal(t, entry.ID, entries[0].ID)

	meta := struct {
		TotalItems int `json:"totalItems"`
	}{}
	require.NoError(t, json.Unmarshal(helpers.DecodeEnvelope(t, listed).Meta, &meta))
	assert.Equal(t, 1, meta.TotalItems)

	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		suppressionsPath+"/"+entry.ID,
		nil,
		bearer(token),
	)
	assert.Equal(t, http.StatusNoContent, deleted.Recorder.Code)

	missing := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		suppressionsPath+"/"+entry.ID,
		nil,
		bearer(token),
	)
	assert.Equal(t, http.StatusNotFound, missing.Recorder.Code)
}

func TestOperationsRoutesRejectNonAdmins(t *testing.T) {
	fixture := newJobFixture(t)

	for _, path := range []string{"/queues", "/queues/embedding_queue/dead", "/providers", "/suppressions"} {
		exchange := helpers.Do(
			t,
			fixture.router,
			http.MethodGet,
			adminBasePath+path,
			nil,
			bearer(fixture.userToken),
		)
		assert.Equal(t, http.StatusForbidden, exchange.Recorder.Code, path)
	}
}

func TestMetricsEndpointExposesRequestsByRoute(t *testing.T) {
	fixture := newJobFixture(t)

	helpers.Do(t, fixture.router, http.MethodGet, "/health", nil, nil)
	exchange := helpers.Do(t, fixture.router, http.MethodGet, "/metrics", nil, nil)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code)

	body := string(exchange.Body)
	assert.True(
		t,
		strings.Contains(
			body,
			`autoreach_http_requests_total{method="GET",route="/health",status="200"}`,
		),
		"request counter by route template",
	)
	assert.Contains(t, body, "autoreach_http_request_duration_seconds_bucket")
}
