//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

var emailChangeLinkPattern = regexp.MustCompile(`https://\S+`)

func TestChangingThePasswordNeedsTheCurrentOneAndEndsEverySession(t *testing.T) {
	router := newRouter(t)
	session := register(t, router, "alex@example.com")

	wrong := helpers.Do(
		t,
		router,
		http.MethodPut,
		mePath+"/password",
		userTypes.ChangePasswordRequest{
			CurrentPassword: "not-my-password", NewPassword: newPassword,
		},
		bearer(session.AccessToken),
	)
	assert.Equal(t, http.StatusForbidden, wrong.Recorder.Code)

	changed := helpers.Do(
		t,
		router,
		http.MethodPut,
		mePath+"/password",
		userTypes.ChangePasswordRequest{
			CurrentPassword: testPassword, NewPassword: newPassword,
		},
		bearer(session.AccessToken),
	)
	require.Equal(t, http.StatusNoContent, changed.Recorder.Code, string(changed.Body))
	helpers.AssertMatchesSpec(t, changed)

	oldSession := helpers.Do(t, router, http.MethodPost, refreshPath,
		authTypes.RefreshRequest{RefreshToken: session.RefreshToken}, nil)
	assert.Equal(t, http.StatusUnauthorized, oldSession.Recorder.Code)

	oldPassword := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusUnauthorized, oldPassword.Recorder.Code)

	login := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: newPassword}, nil)
	assert.Equal(t, http.StatusOK, login.Recorder.Code)
}

func TestTheExportHoldsTheCandidatesRecordsAndNoSecrets(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)
	fixture.processedResume(t, "Alex Candidate, Go")

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		context.Background(),
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)

	exported := helpers.Do(t, fixture.router, http.MethodGet, mePath+"/export", nil,
		bearer(fixture.userToken))
	require.Equal(t, http.StatusOK, exported.Recorder.Code, string(exported.Body))
	helpers.AssertMatchesSpec(t, exported)

	export := helpers.DecodeData[userTypes.UserDataExport](t, exported)
	require.NotNil(t, export.User)
	require.NotNil(t, export.Candidate)
	assert.Equal(t, fixture.candidateID, export.Candidate.ID)
	assert.NotEmpty(t, export.Resumes)
	assert.NotEmpty(t, export.JobMatches)
	assert.NotEmpty(t, export.ChannelAccounts)
	assert.NotEmpty(t, export.OutreachTemplates)
	assert.NotEmpty(t, export.OutreachCampaigns)
	assert.NotEmpty(t, export.OutreachMessages)

	body := string(exported.Body)
	for _, secret := range []string{
		"passwordHash", "password_hash", "encryptedCredentials",
		"storageKey", "trackingToken", "embedding\"",
	} {
		assert.NotContains(t, body, secret)
	}
}

func TestDeletingAnAccountRemovesTheCandidatesDataAndFiles(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)
	resume := fixture.processedResume(t, "Alex Candidate, Go")

	ctx := context.Background()

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)

	stored, err := fixture.deps.Repositories.Resume.FindByID(ctx, resume.ID)
	require.NoError(t, err)

	wrong := helpers.Do(t, fixture.router, http.MethodDelete, mePath,
		userTypes.DeleteAccountRequest{Password: "not-my-password"}, bearer(fixture.userToken))
	assert.Equal(t, http.StatusForbidden, wrong.Recorder.Code)

	deleted := helpers.Do(t, fixture.router, http.MethodDelete, mePath,
		userTypes.DeleteAccountRequest{Password: testPassword}, bearer(fixture.userToken))
	require.Equal(t, http.StatusNoContent, deleted.Recorder.Code, string(deleted.Body))
	helpers.AssertMatchesSpec(t, deleted)
	utils.DrainBackgroundGoroutines()

	for _, table := range []string{
		"candidates", "candidate_resumes", "job_matches",
		"candidate_channel_accounts", "outreach_templates", "outreach_campaigns", "outreach_messages",
	} {
		var remaining int

		require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
			`SELECT COUNT(*) FROM `+table).Scan(&remaining))
		assert.Zero(t, remaining, table)
	}

	_, err = fixture.deps.FileStore.Read(ctx, stored.StorageKey)
	require.ErrorIs(t, err, config.ErrStoredFileNotFound, "the resume file is deleted too")

	assert.NotNil(t, fixture.storedContact(t, fixture.contactID), "shared contacts stay")

	login := helpers.Do(t, fixture.router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusUnauthorized, login.Recorder.Code)
}

func TestAnAdministratorCannotDeleteItself(t *testing.T) {
	fixture := newJobFixture(t)
	token := loginAdmin(t, fixture.router)

	refused := helpers.Do(t, fixture.router, http.MethodDelete, mePath,
		userTypes.DeleteAccountRequest{Password: adminPassword}, bearer(token))
	assert.Equal(t, http.StatusConflict, refused.Recorder.Code)
}

func TestChangingTheEmailIsConfirmedFromTheNewInbox(t *testing.T) {
	router, mailer, _ := mailingRouter(t, false)
	session := register(t, router, "alex@example.com")
	other := register(t, router, "taken@example.com")

	requested := helpers.Do(
		t,
		router,
		http.MethodPost,
		mePath+"/email",
		userTypes.ChangeEmailRequest{
			NewEmail: "Alex.New@Example.com", Password: testPassword,
		},
		bearer(session.AccessToken),
	)
	require.Equal(t, http.StatusAccepted, requested.Recorder.Code, string(requested.Body))
	helpers.AssertMatchesSpec(t, requested)
	utils.DrainBackgroundGoroutines()

	taken := helpers.Do(t, router, http.MethodPost, mePath+"/email", userTypes.ChangeEmailRequest{
		NewEmail: "taken@example.com", Password: testPassword,
	}, bearer(session.AccessToken))
	assert.JSONEq(t, string(requested.Body), string(taken.Body), "a taken address answers the same")
	utils.DrainBackgroundGoroutines()
	assert.Empty(t, mailer.SentTo("taken@example.com"))

	sent := mailer.SentTo("alex.new@example.com")
	require.Len(t, sent, 1)
	assert.Equal(t, emailTemplates.EmailChangeEmailSubject, sent[0].Subject)

	link, err := url.Parse(emailChangeLinkPattern.FindString(sent[0].Body))
	require.NoError(t, err)

	confirm := userTypes.ConfirmEmailChangeRequest{
		Token: link.Query().Get("token"),
		Email: link.Query().Get("email"),
	}
	confirmPath := mePath + "/email/confirm"

	forged := confirm
	forged.Email = "mallory@example.com"
	assert.Equal(
		t,
		http.StatusBadRequest,
		helpers.Do(t, router, http.MethodPost, confirmPath, forged,
			bearer(session.AccessToken)).Recorder.Code,
		"the link names one address only",
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		helpers.Do(t, router, http.MethodPost, confirmPath, confirm,
			bearer(other.AccessToken)).Recorder.Code,
		"only the account it was sent for can use it",
	)

	confirmed := helpers.Do(
		t,
		router,
		http.MethodPost,
		confirmPath,
		confirm,
		bearer(session.AccessToken),
	)
	require.Equal(t, http.StatusNoContent, confirmed.Recorder.Code, string(confirmed.Body))
	helpers.AssertMatchesSpec(t, confirmed)
	utils.DrainBackgroundGoroutines()

	again := helpers.Do(
		t,
		router,
		http.MethodPost,
		confirmPath,
		confirm,
		bearer(session.AccessToken),
	)
	assert.Equal(t, http.StatusBadRequest, again.Recorder.Code, "the link works once")

	notices := mailer.SentTo("alex@example.com")
	require.NotEmpty(t, notices)
	assert.Equal(t, emailTemplates.EmailChangedEmailSubject, notices[len(notices)-1].Subject)

	newLogin := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex.new@example.com", Password: testPassword}, nil)
	require.Equal(t, http.StatusOK, newLogin.Recorder.Code, string(newLogin.Body))
	assert.Equal(t, "alex.new@example.com",
		helpers.DecodeData[authTypes.SessionResponse](t, newLogin).User.Email)

	oldLogin := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "alex@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusUnauthorized, oldLogin.Recorder.Code)
}
