//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func candidatePath(candidateID string) string {
	return "/api/v1/candidates/" + candidateID
}

func TestGetCandidateReturnsTheProfileWithDocumentedDefaults(t *testing.T) {
	router := newRouter(t)
	session := register(t, router, "alex@example.com")

	exchange := helpers.Do(
		t,
		router,
		http.MethodGet,
		candidatePath(*session.User.CandidateID),
		nil,
		bearer(session.AccessToken),
	)

	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)
	candidate := helpers.DecodeData[candidateModel.Candidate](t, exchange)
	assert.Equal(t, *session.User.CandidateID, candidate.ID)
	assert.Equal(t, session.User.UserID, candidate.UserID)
	assert.Equal(t, "alex@example.com", candidate.Email)
	assert.Equal(t, "Alex Candidate", candidate.FullName)
	assert.Nil(t, candidate.LinkedinProfileURL)
	assert.InDelta(t, 0.80, candidate.MinMatchScore, 0.0001)
	assert.False(t, candidate.AutoApplyEnabled, "automatic applying is opt-in")
	assert.False(t, candidate.AutoOutreachEnabled, "automatic outreach is opt-in")
	assert.Equal(t, "UTC", candidate.Timezone)
	assert.JSONEq(t, `{}`, string(candidate.Settings))
}

func TestCandidateRoutesAreInvisibleToOtherTenants(t *testing.T) {
	router := newRouter(t)
	owner := register(t, router, "owner@example.com")
	intruder := register(t, router, "intruder@example.com")
	ownerPath := candidatePath(*owner.User.CandidateID)

	read := helpers.Do(t, router, http.MethodGet, ownerPath, nil, bearer(intruder.AccessToken))
	write := helpers.Do(
		t,
		router,
		http.MethodPatch,
		ownerPath,
		map[string]any{"fullName": "Hijacked"},
		bearer(intruder.AccessToken),
	)
	anonymous := helpers.Do(t, router, http.MethodGet, ownerPath, nil, nil)
	malformed := helpers.Do(
		t,
		router,
		http.MethodGet,
		candidatePath("not-a-uuid"),
		nil,
		bearer(owner.AccessToken),
	)
	unknown := helpers.Do(
		t,
		router,
		http.MethodGet,
		candidatePath("00000000-0000-4000-8000-000000000000"),
		nil,
		bearer(owner.AccessToken),
	)

	assert.Equal(t, http.StatusNotFound, read.Recorder.Code)
	assert.Equal(t, http.StatusNotFound, write.Recorder.Code)
	assert.Equal(t, http.StatusUnauthorized, anonymous.Recorder.Code)
	assert.Equal(t, http.StatusNotFound, malformed.Recorder.Code)
	assert.Equal(t, http.StatusNotFound, unknown.Recorder.Code)
	helpers.AssertMatchesSpec(t, read)
	helpers.AssertMatchesSpec(t, anonymous)

	untouched := helpers.Do(t, router, http.MethodGet, ownerPath, nil, bearer(owner.AccessToken))
	assert.Equal(
		t,
		"Alex Candidate",
		helpers.DecodeData[candidateModel.Candidate](t, untouched).FullName,
	)
}

func TestAdminCanReadAnyCandidate(t *testing.T) {
	router := newRouter(t)
	owner := register(t, router, "owner@example.com")
	adminToken, _, err := utils.IssueAccessToken(adminPrincipal(), utils.GetCurrTimeStamp())
	require.NoError(t, err)

	exchange := helpers.Do(
		t,
		router,
		http.MethodGet,
		candidatePath(*owner.User.CandidateID),
		nil,
		bearer(adminToken),
	)

	assert.Equal(t, http.StatusOK, exchange.Recorder.Code)
}

func TestUpdateCandidateAppliesOnlyTheSuppliedFields(t *testing.T) {
	router := newRouter(t)
	session := register(t, router, "alex@example.com")
	path := candidatePath(*session.User.CandidateID)
	headers := bearer(session.AccessToken)

	exchange := helpers.Do(t, router, http.MethodPatch, path, map[string]any{
		"fullName":           "  Alex Updated  ",
		"linkedinProfileUrl": "https://www.linkedin.com/in/alex",
		"minMatchScore":      0.9,
		"autoApplyEnabled":   true,
		"timezone":           "Asia/Kolkata",
		"settings": map[string]any{
			"emailTracking":      true,
			"applicationProfile": map[string]any{"phone": "+10000000000"},
		},
	}, headers)

	require.Equal(t, http.StatusOK, exchange.Recorder.Code, string(exchange.Body))
	helpers.AssertMatchesSpec(t, exchange)
	updated := helpers.DecodeData[candidateModel.Candidate](t, exchange)
	assert.Equal(t, "Alex Updated", updated.FullName)
	require.NotNil(t, updated.LinkedinProfileURL)
	assert.Equal(t, "https://www.linkedin.com/in/alex", *updated.LinkedinProfileURL)
	assert.InDelta(t, 0.9, updated.MinMatchScore, 0.0001)
	assert.True(t, updated.AutoApplyEnabled)
	assert.False(t, updated.AutoOutreachEnabled, "a field that was not sent must keep its value")
	assert.Equal(t, "Asia/Kolkata", updated.Timezone)
	assert.JSONEq(t, `{"emailTracking":true}`, string(updated.Settings),
		"settings merge, and the application profile only changes through its own endpoint")
	assert.True(
		t,
		updated.UpdatedAt.After(updated.CreatedAt) || updated.UpdatedAt.Equal(updated.CreatedAt),
	)

	cleared := helpers.Do(
		t,
		router,
		http.MethodPatch,
		path,
		map[string]any{"linkedinProfileUrl": ""},
		headers,
	)
	require.Equal(t, http.StatusOK, cleared.Recorder.Code, string(cleared.Body))
	afterClear := helpers.DecodeData[candidateModel.Candidate](t, cleared)
	assert.Nil(t, afterClear.LinkedinProfileURL)
	assert.Equal(t, "Alex Updated", afterClear.FullName)
}

func TestUpdateCandidateRejectsInvalidValues(t *testing.T) {
	router := newRouter(t)
	session := register(t, router, "alex@example.com")
	path := candidatePath(*session.User.CandidateID)
	headers := bearer(session.AccessToken)

	invalid := helpers.Do(t, router, http.MethodPatch, path, map[string]any{
		"minMatchScore":      1.5,
		"timezone":           "Mars/Olympus",
		"linkedinProfileUrl": "https://evil.example.com/in/alex",
	}, headers)
	require.Equal(t, http.StatusUnprocessableEntity, invalid.Recorder.Code, string(invalid.Body))
	helpers.AssertMatchesSpec(t, invalid)
	assert.ElementsMatch(t, []config.ErrorDetail{
		{Field: "minMatchScore", Issue: "must be less than or equal to 1"},
		{Field: "timezone", Issue: "must be a valid IANA time zone"},
		{Field: "linkedinProfileUrl", Issue: "must be an https linkedin.com URL"},
	}, helpers.DecodeEnvelope(t, invalid).Error.Details)

	empty := helpers.Do(t, router, http.MethodPatch, path, map[string]any{}, headers)
	assert.Equal(t, http.StatusBadRequest, empty.Recorder.Code)

	blankName := helpers.Do(
		t,
		router,
		http.MethodPatch,
		path,
		map[string]any{"fullName": "   "},
		headers,
	)
	assert.Equal(t, http.StatusBadRequest, blankName.Recorder.Code)
}

func TestTransactionRollsBackEveryWriteWhenTheWorkFails(t *testing.T) {
	deps := setup.Dependencies(t)
	ctx := context.Background()
	failure := errors.New("abort after first write")

	user := userModel.NewUser()
	user.Email = "rollback@example.com"
	user.PasswordHash = "hash"

	candidate := candidateModel.NewCandidate()
	candidate.UserID = user.ID
	candidate.Email = user.Email
	candidate.FullName = "Rolled Back"

	err := deps.TransactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := deps.Repositories.User.Upsert(ctx, user); err != nil {
			return err
		}

		if _, err := deps.Repositories.Candidate.Upsert(ctx, candidate); err != nil {
			return err
		}

		return failure
	})

	require.ErrorIs(t, err, failure)
	stored, err := deps.Repositories.User.FindByEmail(ctx, "rollback@example.com")
	require.NoError(t, err)
	assert.Nil(t, stored)
}

func TestUpsertReturnsTheSameModelThatASubsequentReadReturns(t *testing.T) {
	deps := setup.Dependencies(t)
	ctx := context.Background()

	user := userModel.NewUser()
	user.Email = "roundtrip@example.com"
	user.PasswordHash = "hash"

	candidate := candidateModel.NewCandidate()
	candidate.UserID = user.ID
	candidate.Email = user.Email
	candidate.FullName = "Round Trip"

	upsertedUser, err := deps.Repositories.User.Upsert(ctx, user)
	require.NoError(t, err)
	assert.Same(t, user, upsertedUser, "upsert returns the model it was given, without a read")

	upsertedCandidate, err := deps.Repositories.Candidate.Upsert(ctx, candidate)
	require.NoError(t, err)
	assert.Same(t, candidate, upsertedCandidate)

	storedUser, err := deps.Repositories.User.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user, storedUser)

	storedCandidate, err := deps.Repositories.Candidate.FindByID(ctx, candidate.ID)
	require.NoError(t, err)
	assert.JSONEq(t, string(candidate.Settings), string(storedCandidate.Settings))

	storedCandidate.Settings = candidate.Settings
	assert.Equal(t, candidate, storedCandidate)

	candidate.FullName = "Round Trip Updated"
	candidate.UpdateTimestamp()

	_, err = deps.Repositories.Candidate.Upsert(ctx, candidate)
	require.NoError(t, err)

	afterUpdate, err := deps.Repositories.Candidate.FindByID(ctx, candidate.ID)
	require.NoError(t, err)
	assert.Equal(t, "Round Trip Updated", afterUpdate.FullName)
	assert.Equal(t, candidate.CreatedAt, afterUpdate.CreatedAt)
	assert.Equal(t, candidate.UpdatedAt, afterUpdate.UpdatedAt)

	var rows int
	require.NoError(
		t,
		deps.Postgres.GetPool().QueryRow(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&rows),
	)
	assert.Equal(t, 1, rows, "an upsert of an existing id must update, not insert")
}
