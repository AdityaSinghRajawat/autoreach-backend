package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const testKeyHash = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

func pointerTo[T any](value T) *T {
	return &value
}

func validIngestItem() jobTypes.IngestJobItem {
	return jobTypes.IngestJobItem{
		Source:         "LINKEDIN",
		CompanyName:    "  Acme Corp ",
		CompanyDomain:  "WWW.Acme.com",
		JobTitle:       " Senior Go Engineer (Remote) ",
		JobDescription: " We are seeking a Go specialist. ",
		ATSType:        "GREENHOUSE",
		ExternalURL:    "https://boards.greenhouse.io/acme/jobs/10293",
	}
}

func existingCandidate() *candidateModel.Candidate {
	candidate := candidateModel.NewCandidate()
	candidate.UserID = utils.GenerateUUID()
	candidate.Email = "alex@example.com"
	candidate.FullName = "Alex Candidate"
	candidate.LinkedinProfileURL = pointerTo("https://www.linkedin.com/in/alex")

	return candidate
}

func TestConstructorsStampIdentityTimestampsAndDefaults(t *testing.T) {
	setup.Config(t)

	user := userModel.NewUser()
	candidate := candidateModel.NewCandidate()
	job := jobModel.NewJobPosting()
	client := authModel.NewAPIClient()
	token := authModel.NewRefreshToken(time.Hour)

	for _, id := range []string{user.ID, candidate.ID, job.ID, client.ID, token.ID, token.FamilyID} {
		_, err := uuid.Parse(id)
		require.NoError(t, err)
	}

	assert.NotEqual(t, userModel.NewUser().ID, user.ID)
	assert.True(t, user.IsActive)
	assert.Equal(t, []string{config.RoleCandidate}, user.Roles)
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
	assert.Equal(t, user.CreatedAt, user.CreatedAt.Truncate(time.Microsecond),
		"timestamps must fit the database's microsecond precision so a stored row equals the model")

	assert.InDelta(t, config.GetDefaultMinMatchScore(), candidate.MinMatchScore, 0.0001)
	assert.False(t, candidate.AutoApplyEnabled, "automatic applying is opt-in")
	assert.False(t, candidate.AutoOutreachEnabled, "automatic outreach is opt-in")
	assert.Equal(t, "UTC", candidate.Timezone)
	assert.JSONEq(t, `{}`, string(candidate.Settings))

	assert.Equal(t, modelHelpers.JobStatusActive, job.Status)
	assert.Equal(t, time.Hour, token.ExpiresAt.Sub(token.CreatedAt))
}

func TestRegisterRequestConvertsToALinkedUserAndCandidate(t *testing.T) {
	setup.Config(t)

	request := authTypes.RegisterRequest{
		Email:    "Alex@Example.com",
		Password: "correct-horse-battery",
		FullName: "  Alex Candidate ",
	}

	user := request.ToUserModel()
	candidate := request.ToCandidateModel(user)

	assert.Equal(t, "alex@example.com", user.Email)
	assert.Empty(
		t,
		user.PasswordHash,
		"the service sets the hash; conversion never sees the password",
	)
	assert.Equal(t, user.ID, candidate.UserID)
	assert.Equal(t, "alex@example.com", candidate.Email)
	assert.Equal(t, "Alex Candidate", candidate.FullName)

	require.Error(t, user.Validate(), "a user without a password hash is not yet valid")

	user.PasswordHash = "argon-hash"
	require.NoError(t, user.Validate())
	require.NoError(t, candidate.Validate())
}

func TestUpdateCandidateRequestAppliesOnlySuppliedFieldsOntoTheExistingEntity(t *testing.T) {
	setup.Config(t)

	existing := existingCandidate()
	request := candidateTypes.UpdateCandidateRequest{
		FullName:         pointerTo("  Alex Updated  "),
		MinMatchScore:    pointerTo(0.9),
		AutoApplyEnabled: pointerTo(true),
		Settings:         json.RawMessage(`{"theme":"dark"}`),
	}

	candidate := request.ToCandidateModel(existing)

	assert.Equal(t, existing.ID, candidate.ID)
	assert.Equal(t, existing.UserID, candidate.UserID)
	assert.Equal(t, existing.CreatedAt, candidate.CreatedAt)
	assert.Equal(t, "Alex Updated", candidate.FullName)
	assert.InDelta(t, 0.9, candidate.MinMatchScore, 0.0001)
	assert.True(t, candidate.AutoApplyEnabled)
	assert.False(t, candidate.AutoOutreachEnabled, "a field that was not sent keeps its value")
	assert.Equal(t, existing.LinkedinProfileURL, candidate.LinkedinProfileURL)
	assert.Equal(t, existing.Timezone, candidate.Timezone)
	assert.JSONEq(t, `{"theme":"dark"}`, string(candidate.Settings))
	require.NoError(t, candidate.Validate())

	assert.Equal(
		t,
		"Alex Candidate",
		existing.FullName,
		"conversion must not mutate the entity it read",
	)
	assert.False(t, existing.AutoApplyEnabled)
}

func TestUpdateCandidateRequestClearsTheLinkedinURLAndReportsEmptiness(t *testing.T) {
	setup.Config(t)

	cleared := candidateTypes.UpdateCandidateRequest{LinkedinProfileURL: pointerTo("")}
	assert.False(t, cleared.IsEmpty())
	assert.Nil(t, cleared.ToCandidateModel(existingCandidate()).LinkedinProfileURL)

	empty := candidateTypes.UpdateCandidateRequest{}
	assert.True(t, empty.IsEmpty())
}

func TestTheModelRejectsWhatConversionDoesNotCheck(t *testing.T) {
	setup.Config(t)

	blankName := candidateTypes.UpdateCandidateRequest{FullName: pointerTo("   ")}
	require.Error(t, blankName.ToCandidateModel(existingCandidate()).Validate())

	badTimezone := candidateTypes.UpdateCandidateRequest{Timezone: pointerTo("Mars/Olympus")}
	require.Error(t, badTimezone.ToCandidateModel(existingCandidate()).Validate())

	notAnObject := candidateTypes.UpdateCandidateRequest{Settings: json.RawMessage(`[1,2]`)}
	require.Error(t, notAnObject.ToCandidateModel(existingCandidate()).Validate())

	unknownRole := userModel.NewUser()
	unknownRole.Email = "alex@example.com"
	unknownRole.PasswordHash = "hash"
	unknownRole.Roles = []string{"SUPERUSER"}
	require.Error(t, unknownRole.Validate())

	token := authModel.NewRefreshToken(-time.Minute)
	token.UserID = utils.GenerateUUID()
	token.TokenHash = testKeyHash
	require.Error(t, token.Validate(), "a refresh token must expire after it is created")
}

func TestIngestJobItemConvertsToANormalisedHashedJobPosting(t *testing.T) {
	setup.Config(t)

	item := validIngestItem()

	job := item.ToJobPostingModel()

	assert.Equal(t, "Acme Corp", job.CompanyName)
	assert.Equal(t, "acme.com", job.CompanyDomain)
	assert.Equal(t, "Senior Go Engineer (Remote)", job.JobTitle)
	assert.Equal(t, "senior go engineer", job.NormalizedTitle)
	assert.Equal(t, "We are seeking a Go specialist.", job.JobDescription)
	assert.Equal(t, jobModel.ComputeJobHash("acme.com", "senior go engineer"), job.JobHash)
	assert.Equal(t, utils.Fingerprint("We are seeking a Go specialist."), job.DescriptionHash)
	assert.Contains(t, string(job.RawPayload), `"companyName"`)
	require.NoError(t, job.Validate())

	request := jobTypes.IngestJobsRequest{Jobs: []jobTypes.IngestJobItem{item, item}}
	jobs := request.ToJobPostingModels()
	require.Len(t, jobs, 2)
	assert.Equal(t, jobs[0].JobHash, jobs[1].JobHash)
	assert.NotEqual(t, jobs[0].ID, jobs[1].ID)
}

func TestTheJobModelEnforcesTheSameRulesAsTheRequest(t *testing.T) {
	setup.Config(t)

	wrongHost := validIngestItem()
	wrongHost.ExternalURL = "https://evil.example.com/acme/jobs/1"
	require.Error(t, wrongHost.ToJobPostingModel().Validate())

	invertedSalary := validIngestItem()
	invertedSalary.SalaryMin = pointerTo(200.0)
	invertedSalary.SalaryMax = pointerTo(100.0)
	require.Error(t, invertedSalary.ToJobPostingModel().Validate())

	onlyMaximum := validIngestItem()
	onlyMaximum.SalaryMax = pointerTo(100.0)
	require.NoError(t, onlyMaximum.ToJobPostingModel().Validate())

	unusableTitle := validIngestItem()
	unusableTitle.JobTitle = "(Remote)"
	require.Error(t, unusableTitle.ToJobPostingModel().Validate())
}

func TestCreateAPIClientRequestConvertsToAnAPIClientAwaitingItsKey(t *testing.T) {
	setup.Config(t)

	request := authTypes.CreateAPIClientRequest{
		Name:   "  linkedin-scraper ",
		Scopes: []string{"job:ingest:any"},
	}

	client := request.ToAPIClientModel()

	assert.Equal(t, "linkedin-scraper", client.Name)
	assert.Equal(t, []string{"job:ingest:any"}, client.Scopes)
	require.Error(t, client.Validate(), "a client without a key is not yet valid")

	client.KeyPrefix = "ark_abc123"
	client.KeyHash = testKeyHash
	require.NoError(t, client.Validate())

	client.KeyHash = "not-a-sha256-digest"
	require.Error(t, client.Validate())
}

func TestUpdateTimestampAdvancesOnlyUpdatedAt(t *testing.T) {
	setup.Config(t)

	user := userModel.NewUser()
	candidate := candidateModel.NewCandidate()
	job := jobModel.NewJobPosting()
	client := authModel.NewAPIClient()
	createdAt := user.CreatedAt
	past := createdAt.Add(-time.Hour)
	user.UpdatedAt, candidate.UpdatedAt, job.UpdatedAt, client.UpdatedAt = past, past, past, past
	job.LastSeenAt = past

	user.UpdateTimestamp()
	candidate.UpdateTimestamp()
	job.UpdateTimestamp()
	client.UpdateTimestamp()

	assert.Equal(t, createdAt, user.CreatedAt)

	for _, updatedAt := range []time.Time{
		user.UpdatedAt, candidate.UpdatedAt, job.UpdatedAt, job.LastSeenAt, client.UpdatedAt,
	} {
		assert.True(t, updatedAt.After(past))
		assert.WithinDuration(t, time.Now(), updatedAt, 5*time.Second)
	}
}
