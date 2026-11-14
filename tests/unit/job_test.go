package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
	workerAggregators "github.com/AdityaSinghRajawat/autoreach-backend/worker/aggregators"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

func TestNormalizeTitleRemovesNoiseButKeepsMeaning(t *testing.T) {
	cases := map[string]string{
		"Senior Go Engineer":                       "senior go engineer",
		"  Senior   Go   Engineer  ":               "senior go engineer",
		"Senior Go Engineer (Remote - US)":         "senior go engineer",
		"Senior Go Engineer [Hybrid]":              "senior go engineer",
		"Senior Go Engineer - Req #48213":          "senior go engineer",
		"Senior Go Engineer, REQ-10293":            "senior go engineer",
		"Sr. C++ / C# Developer":                   "sr c++ c# developer",
		"Engineer - Backend":                       "engineer backend",
		"Staff Engineer, Payments (R-0045921)":     "staff engineer payments",
		"Software Engineer II":                     "software engineer ii",
		"(Remote)":                                 "",
		"Product Manager — Growth & Monetisation!": "product manager growth monetisation",
	}
	for title, expected := range cases {
		assert.Equal(t, expected, modelHelpers.NormalizeJobTitle(title), "title %q", title)
	}
}

func TestJobHashIsStableAcrossFormattingAndDistinctAcrossCompaniesAndTitles(t *testing.T) {
	setup.Config(t)

	base := jobModel.ComputeJobHash(
		"acme.com",
		modelHelpers.NormalizeJobTitle("Senior Go Engineer"),
	)

	assert.Len(t, base, 64)
	assert.Equal(
		t,
		base,
		jobModel.ComputeJobHash(
			"WWW.Acme.com.",
			modelHelpers.NormalizeJobTitle("senior go engineer (Remote)"),
		),
	)
	assert.NotEqual(
		t,
		base,
		jobModel.ComputeJobHash("other.com", modelHelpers.NormalizeJobTitle("Senior Go Engineer")),
	)
	assert.NotEqual(
		t,
		base,
		jobModel.ComputeJobHash("acme.com", modelHelpers.NormalizeJobTitle("Staff Go Engineer")),
	)
}

func TestExternalURLMustBePublicHTTPSAndMatchTheATS(t *testing.T) {
	allowed := map[string]string{
		"https://boards.greenhouse.io/acme/jobs/10293":     "GREENHOUSE",
		"https://job-boards.greenhouse.io/acme/jobs/10293": "GREENHOUSE",
		"https://jobs.lever.co/acme/abc":                   "LEVER",
		"https://jobs.eu.lever.co/acme/abc":                "LEVER",
		"https://acme.wd5.myworkdayjobs.com/en-US/careers": "WORKDAY",
		"https://careers.acme.com/jobs/1":                  "OTHER",
	}
	for rawURL, atsType := range allowed {
		assert.True(
			t,
			validations.IsAllowedExternalURL(rawURL, atsType),
			"%s as %s",
			rawURL,
			atsType,
		)
	}

	rejected := map[string]string{
		"http://boards.greenhouse.io/acme/jobs/1":     "GREENHOUSE",
		"https://evil.example.com/greenhouse.io":      "GREENHOUSE",
		"https://greenhouse.io.evil.example.com/acme": "GREENHOUSE",
		"https://notgreenhouse.io/acme":               "GREENHOUSE",
		"https://jobs.lever.co.evil.com/acme":         "LEVER",
		"https://127.0.0.1/jobs":                      "OTHER",
		"https://169.254.169.254/latest/meta-data":    "OTHER",
		"https://localhost/jobs":                      "OTHER",
		"https://db.internal/jobs":                    "OTHER",
		"https://intranet/jobs":                       "OTHER",
		"https://user:pass@careers.acme.com/jobs":     "OTHER",
		"ftp://careers.acme.com/jobs":                 "OTHER",
		"not a url":                                   "OTHER",
	}
	for rawURL, atsType := range rejected {
		assert.False(
			t,
			validations.IsAllowedExternalURL(rawURL, atsType),
			"%s as %s",
			rawURL,
			atsType,
		)
	}
}

func TestEmbeddingTextIsStructuredAndTruncated(t *testing.T) {
	location := "Remote - US"
	source := jobTypes.JobEmbeddingSource{
		CompanyName:    "Acme Corp",
		JobTitle:       "Senior Go Engineer",
		Location:       &location,
		JobDescription: "We are seeking a Go specialist.",
	}

	full := source.EmbeddingText(0)
	assert.Equal(
		t,
		"Title: Senior Go Engineer\nCompany: Acme Corp\nLocation: Remote - US\n\nWe are seeking a Go specialist.",
		full,
	)
	assert.Equal(t, "Title: Sen", source.EmbeddingText(10))
}

func TestFormatVectorProducesAPgvectorLiteral(t *testing.T) {
	assert.Equal(t, "[]", db.FormatVector(nil))
	assert.Equal(t, "[1,-0.5,0.25]", db.FormatVector([]float32{1, -0.5, 0.25}))
}

func TestEmbedAggregatorFoldsDeduplicatesAndSkipsMalformedMessages(t *testing.T) {
	first := "c39a04a1-872f-4c5c-9c04-749e7cf93c01"
	second := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	message := func(payload string) workerConfigs.Message {
		return workerConfigs.Message{
			TaskID:   workerConfigs.TaskEmbedJob,
			Payload:  []byte(payload),
			Metadata: map[string]string{workerConfigs.MetadataCorrelationID: "corr-1"},
		}
	}

	aggregated, err := workerAggregators.AggregateEmbedJobs(
		workerConfigs.GroupEmbedJob,
		[]workerConfigs.Message{
			message(`{"jobId":"` + first + `"}`),
			message(`{"jobId":"` + second + `"}`),
			message(`{"jobId":"` + first + `"}`),
			message(`{"jobId":"not-a-uuid"}`),
			message(`not json`),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, workerConfigs.TaskEmbedJobBatch, aggregated.TaskID)
	assert.Equal(t, "corr-1", aggregated.Metadata[workerConfigs.MetadataCorrelationID])

	var payload struct {
		JobIDs []string `json:"jobIds"`
	}
	require.NoError(t, json.Unmarshal(aggregated.Payload, &payload))
	assert.Equal(t, []string{first, second}, payload.JobIDs)

	_, err = workerAggregators.AggregateEmbedJobs(
		workerConfigs.GroupEmbedJob,
		[]workerConfigs.Message{message(`not json`)},
	)
	assert.Error(t, err)
}

func TestWorkerTopologyRejectsDriftBetweenConfigsHandlersAndAggregators(t *testing.T) {
	handler := workerConfigs.WorkerMessageHandler(nil)

	complete := workerConfigs.WorkerHandlerRegistry{}
	for _, taskID := range workerConfigs.GetTaskIDs() {
		complete[taskID] = handler
	}

	aggregators := workerAggregators.NewWorkerAggregatorRegistry()

	assert.NoError(t, workerConfigs.ValidateTopology(complete, aggregators))

	missingHandler := workerConfigs.WorkerHandlerRegistry{workerConfigs.TaskEmbedJob: handler}
	assert.ErrorContains(
		t,
		workerConfigs.ValidateTopology(missingHandler, aggregators),
		"no handler",
	)

	extraHandler := workerConfigs.WorkerHandlerRegistry{"task:unknown": handler}
	for taskID := range complete {
		extraHandler[taskID] = handler
	}

	assert.ErrorContains(t, workerConfigs.ValidateTopology(extraHandler, aggregators), "no config")

	assert.ErrorContains(
		t,
		workerConfigs.ValidateTopology(complete, workerConfigs.WorkerAggregatorRegistry{}),
		"no aggregator",
	)
}

func TestRetryDelayFollowsTheConfiguredBackoffAndClampsAtTheLastStep(t *testing.T) {
	assert.Equal(t, 5*time.Second, workerConfigs.RetryDelay(workerConfigs.TaskEmbedJob, 0))
	assert.Equal(t, 15*time.Second, workerConfigs.RetryDelay(workerConfigs.TaskEmbedJob, 1))
	assert.Equal(t, 5*time.Minute, workerConfigs.RetryDelay(workerConfigs.TaskEmbedJob, 99))
	assert.Equal(t, 10*time.Second, workerConfigs.RetryDelay("task:unknown", 0))
}

func TestOnlyTransientErrorCodesAreRetryable(t *testing.T) {
	assert.True(t, config.IsRetryableErrCode(config.GetErrCodeServiceUnavailable()))
	assert.True(t, config.IsRetryableErrCode(config.GetErrCodeInternalServerError()))
	assert.False(t, config.IsRetryableErrCode(config.GetErrCodeValidation()))
	assert.False(t, config.IsRetryableErrCode(config.GetErrCodeUpstreamRejected()))
	assert.False(t, config.IsRetryableErrCode(config.GetErrCodeBadRequest()))
}
