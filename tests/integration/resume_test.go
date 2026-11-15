//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	matchingTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/matchingManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerServer "github.com/AdityaSinghRajawat/autoreach-backend/worker/server"
)

const (
	resumeMarker  = "RESUME_ALPHA"
	nearJobMarker = "JOB_NEAR"
	farJobMarker  = "JOB_FAR"
)

func resumesPath(candidateID string) string {
	return candidatePath(candidateID) + "/resumes"
}

func matchedJobsPath(candidateID, query string) string {
	return candidatePath(candidateID) + "/jobs" + query
}

func directionVector(components map[int]float64) []float32 {
	vector := make([]float32, config.GetOpenAIEmbeddingDimensions())

	norm := 0.0
	for _, value := range components {
		norm += value * value
	}

	for index, value := range components {
		vector[index] = float32(value / math.Sqrt(norm))
	}

	return vector
}

func pinMatchingVectors(embedder *mocks.Embedder) {
	embedder.Pin(resumeMarker, directionVector(map[int]float64{0: 1}))
	embedder.Pin(nearJobMarker, directionVector(map[int]float64{0: 0.9, 1: math.Sqrt(1 - 0.81)}))
	embedder.Pin(farJobMarker, directionVector(map[int]float64{2: 1}))
}

func (f jobFixture) uploadResume(
	t *testing.T,
	fileName string,
	content []byte,
) candidateTypes.UploadResumeResponse {
	t.Helper()

	exchange := helpers.DoMultipart(
		t, f.router, resumesPath(f.candidateID), config.GetResumeFormField(), fileName, content,
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusAccepted, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[candidateTypes.UploadResumeResponse](t, exchange)
}

func (f jobFixture) getResume(t *testing.T, resumeID string) candidateModel.Resume {
	t.Helper()

	exchange := helpers.Do(
		t,
		f.router,
		http.MethodGet,
		resumesPath(f.candidateID)+"/"+resumeID,
		nil,
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[candidateModel.Resume](t, exchange)
}

func (f jobFixture) processedResume(t *testing.T, lines ...string) candidateModel.Resume {
	t.Helper()

	uploaded := f.uploadResume(t, "resume.pdf", helpers.BuildPDF(t, lines...))

	result, customErr := f.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr)
	require.Equal(t, modelHelpers.ResumeStatusEmbedded, result.Status)

	return f.getResume(t, uploaded.Resume.ID)
}

func (f jobFixture) embeddedJob(t *testing.T, title, description string) string {
	t.Helper()

	ingested := f.ingest(t, withFields(validJob(), map[string]any{
		"jobTitle":       title,
		"jobDescription": description,
	}))
	require.Len(t, ingested.QueuedTasks, 1)

	jobID := ingested.QueuedTasks[0].JobID
	_, customErr := f.deps.Services.Job.EmbedJobs(context.Background(), []string{jobID})
	require.Nil(t, customErr)

	return jobID
}

func (f jobFixture) listMatchedJobs(
	t *testing.T,
	query string,
) ([]matchingTypes.MatchedJob, helpers.Envelope) {
	t.Helper()

	exchange := helpers.Do(
		t,
		f.router,
		http.MethodGet,
		matchedJobsPath(f.candidateID, query),
		nil,
		bearer(f.userToken),
	)
	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)

	return helpers.DecodeData[[]matchingTypes.MatchedJob](
			t,
			exchange,
		), helpers.DecodeEnvelope(
			t,
			exchange,
		)
}

func TestUploadedPDFResumeIsStoredParsedEmbeddedAndActivated(t *testing.T) {
	fixture := newJobFixture(t)

	uploaded := fixture.uploadResume(t, "../Alex Resume.pdf", helpers.BuildPDF(
		t,
		"Alex Candidate",
		"Senior engineer: Go, Kubernetes and PostgreSQL.",
		"Also C++ and Node.js",
	))

	assert.Equal(t, string(workerConfigs.TaskProcessResume), uploaded.TaskName)
	assert.Equal(t, workerConfigs.PublishStatusEnqueued, uploaded.Status)
	assert.Equal(t, "Alex Resume.pdf", uploaded.Resume.FileName)
	assert.Equal(t, modelHelpers.ResumeStatusUploaded, uploaded.Resume.Status)
	assert.False(t, uploaded.Resume.IsActive)

	result, customErr := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr)
	assert.Equal(t, modelHelpers.ResumeStatusEmbedded, result.Status)

	resume := fixture.getResume(t, uploaded.Resume.ID)
	assert.Equal(t, modelHelpers.ResumeStatusEmbedded, resume.Status)
	assert.True(t, resume.IsActive)
	assert.Nil(t, resume.FailureReason)

	metadata := candidateTypes.ResumeMetadata{}
	require.NoError(t, json.Unmarshal(resume.ParsedMetadata, &metadata))
	assert.Equal(t, []string{"c++", "go", "kubernetes", "node.js", "postgresql"}, metadata.Skills)
	assert.Positive(t, metadata.WordCount)

	var dimensions int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(),
		`SELECT vector_dims(embedding) FROM candidate_resumes WHERE id = $1`, resume.ID,
	).Scan(&dimensions))
	assert.Equal(t, config.GetOpenAIEmbeddingDimensions(), dimensions)
}

func TestUploadedDocxResumeIsProcessed(t *testing.T) {
	fixture := newJobFixture(t)

	uploaded := fixture.uploadResume(
		t,
		"resume.docx",
		helpers.BuildDocx(t, "Alex Candidate", "Python and Django"),
	)
	assert.Equal(t, modelHelpers.ContentTypeDOCX, uploaded.Resume.ContentType)

	_, customErr := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr)

	resume := fixture.getResume(t, uploaded.Resume.ID)
	assert.True(t, resume.IsActive)
	assert.JSONEq(
		t,
		`{"skills":["django","python"],"jobTitles":[],"yearsOfExperience":null,"education":[],`+
			`"wordCount":5,"parsedBy":"dictionary"}`,
		string(resume.ParsedMetadata),
		"the skill dictionary is the fallback when the parsing model fails",
	)
}

func TestUploadRejectsMissingUnsupportedAndOversizedFiles(t *testing.T) {
	fixture := newJobFixture(t)
	path := resumesPath(fixture.candidateID)
	headers := bearer(fixture.userToken)
	field := config.GetResumeFormField()

	missing := helpers.DoMultipart(t, fixture.router, path, "", "", nil, headers)
	unsupported := helpers.DoMultipart(
		t,
		fixture.router,
		path,
		field,
		"resume.pdf",
		[]byte("plain text"),
		headers,
	)
	oversized := helpers.DoMultipart(
		t, fixture.router, path, field, "resume.pdf",
		append([]byte("%PDF-1.4\n"), make([]byte, config.GetResumeMaxBytes())...), headers,
	)
	anonymous := helpers.DoMultipart(
		t,
		fixture.router,
		path,
		field,
		"resume.pdf",
		helpers.BuildPDF(t, "x"),
		nil,
	)

	require.Equal(t, http.StatusUnprocessableEntity, missing.Recorder.Code, string(missing.Body))
	require.Equal(
		t,
		http.StatusUnprocessableEntity,
		unsupported.Recorder.Code,
		string(unsupported.Body),
	)
	assert.Equal(t, []config.ErrorDetail{{Field: "file", Issue: "must be a PDF or DOCX document"}},
		helpers.DecodeEnvelope(t, unsupported).Error.Details)
	assert.Equal(t, http.StatusRequestEntityTooLarge, oversized.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
	helpers.AssertMatchesSpec(t, unsupported)
	helpers.AssertMatchesSpec(t, oversized)
}

func TestUnreadableResumeIsMarkedFailedWithoutRetrying(t *testing.T) {
	fixture := newJobFixture(t)

	uploaded := fixture.uploadResume(t, "broken.pdf", []byte("%PDF-1.4\nnot a real document"))

	result, customErr := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr, "a document that can never be read must not be retried")
	assert.Equal(t, modelHelpers.ResumeStatusFailed, result.Status)

	resume := fixture.getResume(t, uploaded.Resume.ID)
	assert.Equal(t, modelHelpers.ResumeStatusFailed, resume.Status)
	assert.False(t, resume.IsActive)
	require.NotNil(t, resume.FailureReason)
	assert.Equal(t, config.ErrUnreadableDocument.Error(), *resume.FailureReason)
}

func TestEmbeddingOutageLeavesTheResumeRetryableAndRejectionFailsIt(t *testing.T) {
	fixture := newJobFixture(t)
	embedder := setup.Embedder(t)
	uploaded := fixture.uploadResume(t, "resume.pdf", helpers.BuildPDF(t, "Alex Candidate, Go"))

	embedder.SetErr(&utils.APIError{StatusCode: http.StatusTooManyRequests})

	_, transient := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.NotNil(t, transient)
	assert.True(t, config.IsRetryableErrCode(transient.ErrCode))
	assert.Equal(
		t,
		modelHelpers.ResumeStatusParsed,
		fixture.getResume(t, uploaded.Resume.ID).Status,
	)

	embedder.SetErr(nil)

	_, customErr := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr, "a retry after the outage completes the resume")
	assert.Equal(t, 1, setup.ChatCompleter(t).CallCount(),
		"a retry reuses the parsed text instead of paying for parsing again")
	assert.True(t, fixture.getResume(t, uploaded.Resume.ID).IsActive)

	rejected := fixture.uploadResume(t, "second.pdf", helpers.BuildPDF(t, "Alex Candidate, Python"))

	embedder.SetErr(&utils.APIError{StatusCode: http.StatusUnauthorized})

	_, permanent := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		rejected.Resume.ID,
	)
	require.NotNil(t, permanent)
	assert.False(t, config.IsRetryableErrCode(permanent.ErrCode))
	assert.Equal(
		t,
		modelHelpers.ResumeStatusFailed,
		fixture.getResume(t, rejected.Resume.ID).Status,
	)
	assert.True(
		t,
		fixture.getResume(t, uploaded.Resume.ID).IsActive,
		"the earlier resume stays active",
	)
}

func TestOnlyOneResumeIsActiveAndActivationCanBeSwitched(t *testing.T) {
	fixture := newJobFixture(t)
	headers := bearer(fixture.userToken)
	first := fixture.processedResume(t, "First resume, Go")
	second := fixture.processedResume(t, "Second resume, Python")
	pending := fixture.uploadResume(t, "pending.pdf", helpers.BuildPDF(t, "Pending resume"))

	assert.False(
		t,
		fixture.getResume(t, first.ID).IsActive,
		"a newly embedded resume replaces the active one",
	)
	assert.True(t, fixture.getResume(t, second.ID).IsActive)

	activate := helpers.Do(
		t, fixture.router, http.MethodPatch, resumesPath(fixture.candidateID)+"/"+first.ID,
		map[string]any{"isActive": true}, headers,
	)
	require.Equal(t, http.StatusOK, activate.Recorder.Code, string(activate.Body))
	helpers.AssertMatchesSpec(t, activate)
	assert.True(t, helpers.DecodeData[candidateModel.Resume](t, activate).IsActive)
	assert.True(t, fixture.getResume(t, first.ID).IsActive)
	assert.False(t, fixture.getResume(t, second.ID).IsActive)

	notReady := helpers.Do(
		t, fixture.router, http.MethodPatch, resumesPath(fixture.candidateID)+"/"+pending.Resume.ID,
		map[string]any{"isActive": true}, headers,
	)
	deactivate := helpers.Do(
		t, fixture.router, http.MethodPatch, resumesPath(fixture.candidateID)+"/"+first.ID,
		map[string]any{"isActive": false}, headers,
	)
	assert.Equal(t, http.StatusConflict, notReady.Recorder.Code)
	assert.Equal(t, http.StatusUnprocessableEntity, deactivate.Recorder.Code)
	helpers.AssertMatchesSpec(t, notReady)

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		resumesPath(fixture.candidateID),
		nil,
		headers,
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)

	resumes := helpers.DecodeData[[]candidateModel.Resume](t, listed)
	require.Len(t, resumes, 3)
	assert.Equal(t, pending.Resume.ID, resumes[0].ID, "newest first")
}

func TestResumesAreInvisibleToOtherTenantsAndDeletionRemovesTheFile(t *testing.T) {
	fixture := newJobFixture(t)
	intruder := register(t, fixture.router, "intruder@example.com")
	resume := fixture.processedResume(t, "Alex Candidate, Go")
	ownerPath := resumesPath(fixture.candidateID) + "/" + resume.ID
	intruderPath := resumesPath(*intruder.User.CandidateID) + "/" + resume.ID

	viaOwnerPath := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		ownerPath,
		nil,
		bearer(intruder.AccessToken),
	)
	viaOwnPath := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		intruderPath,
		nil,
		bearer(intruder.AccessToken),
	)
	deleteViaOwnPath := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		intruderPath,
		nil,
		bearer(intruder.AccessToken),
	)
	malformed := helpers.Do(
		t, fixture.router, http.MethodGet, resumesPath(fixture.candidateID)+"/not-a-uuid", nil,
		bearer(fixture.userToken),
	)

	assert.Equal(t, http.StatusNotFound, viaOwnerPath.Recorder.Code)
	assert.Equal(
		t,
		http.StatusNotFound,
		viaOwnPath.Recorder.Code,
		"a resume id from another tenant must not resolve",
	)
	assert.Equal(t, http.StatusNotFound, deleteViaOwnPath.Recorder.Code)
	assert.Equal(t, http.StatusNotFound, malformed.Recorder.Code)

	stored, err := fixture.deps.Repositories.Resume.FindByID(context.Background(), resume.ID)
	require.NoError(t, err)

	_, err = fixture.deps.FileStore.Read(context.Background(), stored.StorageKey)
	require.NoError(t, err)

	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		ownerPath,
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusNoContent, deleted.Recorder.Code)
	helpers.AssertMatchesSpec(t, deleted)
	utils.DrainBackgroundGoroutines()

	_, err = fixture.deps.FileStore.Read(context.Background(), stored.StorageKey)
	require.ErrorIs(t, err, config.ErrStoredFileNotFound)

	again := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		ownerPath,
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusNotFound, again.Recorder.Code)
}

func TestResumeIsMatchedAgainstJobsAndListedByScore(t *testing.T) {
	fixture := newJobFixture(t)
	pinMatchingVectors(setup.Embedder(t))

	nearJobID := fixture.embeddedJob(t, "Senior Go Engineer", "We need Go. "+nearJobMarker)
	farJobID := fixture.embeddedJob(t, "Pastry Chef", "We need croissants. "+farJobMarker)
	resume := fixture.processedResume(t, "Alex Candidate "+resumeMarker)

	result, customErr := fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)
	assert.Equal(t, 2, result.Considered)
	assert.Equal(t, 2, result.Stored)

	aboveThreshold, envelope := fixture.listMatchedJobs(t, "")
	require.Len(t, aboveThreshold, 1, "the default minScore is the candidate's 0.80 threshold")
	assert.Equal(t, nearJobID, aboveThreshold[0].JobID)
	assert.InDelta(t, 0.9, aboveThreshold[0].MatchScore, 0.001)
	assert.Equal(t, "Acme Corp", aboveThreshold[0].CompanyName)
	assert.Equal(t, "Senior Go Engineer", aboveThreshold[0].JobTitle)
	assert.Equal(t, "GREENHOUSE", aboveThreshold[0].ATSType)
	assert.Zero(t, aboveThreshold[0].DiscoveredContactsCount)
	assert.Nil(t, aboveThreshold[0].ApplicationStatus)
	assert.JSONEq(t, `{"page":1,"perPage":20,"totalItems":1,"totalPages":1}`, string(envelope.Meta))

	all, envelope := fixture.listMatchedJobs(t, "?minScore=0&perPage=1&page=2")
	require.Len(t, all, 1)
	assert.Equal(t, farJobID, all[0].JobID, "page two of a best-first list holds the weaker match")
	assert.JSONEq(t, `{"page":2,"perPage":1,"totalItems":2,"totalPages":2}`, string(envelope.Meta))

	invalid := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		matchedJobsPath(fixture.candidateID, "?minScore=1.5"),
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Recorder.Code)
	helpers.AssertMatchesSpec(t, invalid)
}

func TestNewlyEmbeddedJobIsMatchedAgainstActiveResumes(t *testing.T) {
	fixture := newJobFixture(t)
	pinMatchingVectors(setup.Embedder(t))

	resume := fixture.processedResume(t, "Alex Candidate "+resumeMarker)
	_, customErr := fixture.deps.Services.Matching.MatchResume(context.Background(), resume.ID)
	require.Nil(t, customErr)

	before, _ := fixture.listMatchedJobs(t, "")
	require.Empty(t, before)

	jobID := fixture.embeddedJob(t, "Senior Go Engineer", "We need Go. "+nearJobMarker)

	result, customErr := fixture.deps.Services.Matching.MatchJob(context.Background(), jobID)
	require.Nil(t, customErr)
	assert.Equal(t, 1, result.Stored)

	again, customErr := fixture.deps.Services.Matching.MatchJob(context.Background(), jobID)
	require.Nil(t, customErr)
	assert.Equal(t, 1, again.Stored, "matching the same job twice updates the one row")

	after, envelope := fixture.listMatchedJobs(t, "")
	require.Len(t, after, 1)
	assert.Equal(t, jobID, after[0].JobID)
	assert.JSONEq(t, `{"page":1,"perPage":20,"totalItems":1,"totalPages":1}`, string(envelope.Meta))
}

func TestReplacingTheResumeDropsMatchesFromTheOldOne(t *testing.T) {
	fixture := newJobFixture(t)
	embedder := setup.Embedder(t)
	pinMatchingVectors(embedder)
	embedder.Pin("RESUME_BETA", directionVector(map[int]float64{2: 1}))

	nearJobID := fixture.embeddedJob(t, "Senior Go Engineer", "We need Go. "+nearJobMarker)
	farJobID := fixture.embeddedJob(t, "Pastry Chef", "We need croissants. "+farJobMarker)

	first := fixture.processedResume(t, "Alex Candidate "+resumeMarker)
	_, customErr := fixture.deps.Services.Matching.MatchResume(context.Background(), first.ID)
	require.Nil(t, customErr)

	matched, _ := fixture.listMatchedJobs(t, "")
	require.Len(t, matched, 1)
	assert.Equal(t, nearJobID, matched[0].JobID)

	second := fixture.processedResume(t, "Alex Candidate RESUME_BETA")
	_, customErr = fixture.deps.Services.Matching.MatchResume(context.Background(), second.ID)
	require.Nil(t, customErr)

	inactive, customErr := fixture.deps.Services.Matching.MatchResume(
		context.Background(),
		first.ID,
	)
	require.Nil(t, customErr)
	assert.Zero(t, inactive.Stored, "an inactive resume is not matched")

	rematched, _ := fixture.listMatchedJobs(t, "")
	require.Len(t, rematched, 1)
	assert.Equal(t, farJobID, rematched[0].JobID)
	assert.InDelta(t, 1.0, rematched[0].MatchScore, 0.001)

	var staleRows int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM job_matches WHERE resume_id = $1`, first.ID,
	).Scan(&staleRows))
	assert.Zero(t, staleRows)
}

func TestReconcilerRequeuesResumesWhoseTaskWasLost(t *testing.T) {
	fixture := newJobFixture(t)
	workerClient := mocks.NewWorkerClient()

	service, err := candidateService.NewCandidateService(
		candidateService.CandidateServiceDependencies{
			CandidateRepo:      fixture.deps.Repositories.Candidate,
			ResumeRepo:         fixture.deps.Repositories.Resume,
			TransactionManager: fixture.deps.TransactionManager,
			WorkerClient:       workerClient,
			FileStore:          fixture.deps.FileStore,
			Embedder:           setup.Embedder(t),
			ChatCompleter:      setup.ChatCompleter(t),
		},
	)
	require.NoError(t, err)

	request := candidateTypes.UploadResumeRequest{}
	request.SetFile("resume.pdf", helpers.BuildPDF(t, "Alex Candidate, Go"))

	workerClient.SetErr(errors.New("redis unavailable"))

	uploaded, customErr := service.UploadResume(
		context.Background(), request.ToResumeModel(fixture.candidateID), request.Content,
	)
	require.Nil(t, customErr, "a queue outage must not fail an upload that is already stored")
	assert.Equal(t, workerConfigs.PublishStatusDeferred, uploaded.Status)

	workerClient.SetErr(nil)
	time.Sleep(20 * time.Millisecond)

	requeued, customErr := service.RequeuePendingResumes(context.Background())
	require.Nil(t, customErr)
	assert.Equal(t, 1, requeued)

	published := workerClient.Snapshot()
	require.Len(t, published, 1)
	assert.Equal(t, workerConfigs.TaskProcessResume, published[0].TaskID)
}

func TestWorkerProcessesAnUploadedResumeAndMatchesItEndToEnd(t *testing.T) {
	fixture := newJobFixture(t)
	pinMatchingVectors(setup.Embedder(t))

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

	ingested := fixture.ingest(t, withFields(validJob(), map[string]any{
		"jobDescription": "We need Go. " + nearJobMarker,
	}))
	require.Len(t, ingested.QueuedTasks, 1)

	fixture.uploadResume(t, "resume.pdf", helpers.BuildPDF(t, "Alex Candidate "+resumeMarker))

	require.Eventually(t, func() bool {
		matched, _ := fixture.listMatchedJobs(t, "")

		return len(matched) == 1 && matched[0].JobID == ingested.QueuedTasks[0].JobID
	}, 40*time.Second, 250*time.Millisecond, "the worker did not produce the match")
}

func TestResumeProfileFromTheParsingModelIsNormalisedAndStored(t *testing.T) {
	fixture := newJobFixture(t)
	years := 7.5
	graduated := 2015
	setup.ChatCompleter(t).ReplyWith(candidateTypes.ResumeProfile{
		Skills:            []string{" Go ", "Kubernetes", "go", "", "PostgreSQL"},
		JobTitles:         []string{"Staff Engineer", "Staff Engineer", "Backend Engineer"},
		YearsOfExperience: &years,
		Education: []candidateTypes.ResumeEducation{
			{Degree: " BSc ", FieldOfStudy: "Computer Science", GraduationYear: &graduated},
			{},
		},
	})

	uploaded := fixture.uploadResume(t, "resume.pdf", helpers.BuildPDF(t, "Alex Candidate, Go"))

	_, customErr := fixture.deps.Services.Candidate.ProcessResume(
		context.Background(),
		uploaded.Resume.ID,
	)
	require.Nil(t, customErr)

	metadata := candidateTypes.ResumeMetadata{}
	require.NoError(
		t,
		json.Unmarshal(fixture.getResume(t, uploaded.Resume.ID).ParsedMetadata, &metadata),
	)
	assert.Equal(t, mocks.ChatModel, metadata.ParsedBy)
	assert.Equal(t, []string{"go", "kubernetes", "postgresql"}, metadata.Skills)
	assert.Equal(t, []string{"Staff Engineer", "Backend Engineer"}, metadata.JobTitles)
	require.NotNil(t, metadata.YearsOfExperience)
	assert.InDelta(t, 7.5, *metadata.YearsOfExperience, 0.001)
	assert.Equal(t, []candidateTypes.ResumeEducation{
		{Degree: "BSc", FieldOfStudy: "Computer Science", GraduationYear: &graduated},
	}, metadata.Education)
}
