//go:build integration

package integration

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
)

const (
	senderAddress  = "alex@example.com"
	senderPassword = "alex-password"
)

func channelAccountsPath(candidateID string) string {
	return candidatePath(candidateID) + "/channel-accounts"
}

func mailAccountRequest(t *testing.T, address, password string) map[string]any {
	t.Helper()

	smtpPort, err := strconv.Atoi(os.Getenv("TEST_SMTP_PORT"))
	require.NoError(t, err)
	imapPort, err := strconv.Atoi(os.Getenv("TEST_IMAP_PORT"))
	require.NoError(t, err)

	return map[string]any{
		"emailAddress": address,
		"displayName":  "Alex Candidate",
		"username":     address,
		"password":     password,
		"smtpHost":     os.Getenv("TEST_MAIL_HOST"),
		"smtpPort":     smtpPort,
		"smtpSecurity": "NONE",
		"imapHost":     os.Getenv("TEST_MAIL_HOST"),
		"imapPort":     imapPort,
		"imapSecurity": "NONE",
	}
}

func (f jobFixture) connectMailbox(t *testing.T, password string) outreachModel.ChannelAccount {
	t.Helper()

	created := helpers.Do(t, f.router, http.MethodPost, channelAccountsPath(f.candidateID),
		mailAccountRequest(t, senderAddress, password), bearer(f.userToken))
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))
	helpers.AssertMatchesSpec(t, created)
	assert.NotContains(t, string(created.Body), password, "credentials are never returned")

	return helpers.DecodeData[outreachModel.ChannelAccount](t, created)
}

func (f jobFixture) verifyMailbox(t *testing.T, accountID string) outreachModel.ChannelAccount {
	t.Helper()

	verified := helpers.Do(t, f.router, http.MethodPost,
		channelAccountsPath(f.candidateID)+"/"+accountID+"/verify", nil, bearer(f.userToken))
	require.Equal(t, http.StatusOK, verified.Recorder.Code, string(verified.Body))
	helpers.AssertMatchesSpec(t, verified)

	return helpers.DecodeData[outreachModel.ChannelAccount](t, verified)
}

func TestMailboxIsStoredEncryptedAndVerifiedAgainstTheMailServer(t *testing.T) {
	fixture := newJobFixture(t)

	account := fixture.connectMailbox(t, senderPassword)
	assert.Equal(t, modelHelpers.ChannelAccountStatusUnverified, account.Status)
	assert.Equal(t, 30, account.DailyLimit)

	var stored []byte

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(context.Background(),
		"SELECT encrypted_credentials FROM candidate_channel_accounts WHERE id = $1", account.ID,
	).Scan(&stored))
	assert.NotContains(t, string(stored), senderPassword)

	verified := fixture.verifyMailbox(t, account.ID)
	assert.Equal(t, modelHelpers.ChannelAccountStatusActive, verified.Status)
	assert.NotNil(t, verified.LastVerifiedAt)
	assert.Nil(t, verified.LastError)

	var uidValidity *int64

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(context.Background(),
		"SELECT imap_uid_validity FROM candidate_channel_accounts WHERE id = $1", account.ID,
	).Scan(&uidValidity))
	assert.NotNil(t, uidValidity, "verification records the inbox cursor so old mail is not swept")

	listed := helpers.Do(
		t,
		fixture.router,
		http.MethodGet,
		channelAccountsPath(fixture.candidateID),
		nil,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusOK, listed.Recorder.Code)
	helpers.AssertMatchesSpec(t, listed)
	assert.Len(t, helpers.DecodeData[[]outreachModel.ChannelAccount](t, listed), 1)

	duplicate := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		channelAccountsPath(fixture.candidateID),
		mailAccountRequest(t, "ALEX@example.com", senderPassword),
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusConflict, duplicate.Recorder.Code)

	path := channelAccountsPath(fixture.candidateID) + "/" + account.ID
	deleted := helpers.Do(
		t,
		fixture.router,
		http.MethodDelete,
		path,
		nil,
		bearer(fixture.userToken),
	)
	assert.Equal(t, http.StatusNoContent, deleted.Recorder.Code)
	helpers.AssertMatchesSpec(t, deleted)

	again := helpers.Do(t, fixture.router, http.MethodDelete, path, nil, bearer(fixture.userToken))
	assert.Equal(t, http.StatusNotFound, again.Recorder.Code)
}

func TestMailboxWithWrongPasswordFailsVerificationWithTheReason(t *testing.T) {
	fixture := newJobFixture(t)

	account := fixture.connectMailbox(t, "wrong-password")
	verified := fixture.verifyMailbox(t, account.ID)

	assert.Equal(t, modelHelpers.ChannelAccountStatusFailed, verified.Status)
	require.NotNil(t, verified.LastError)
	assert.Equal(t, config.ErrMailLoginRejected.Error(), *verified.LastError,
		"users see a category, never the mail server's own reply")
}

func TestMailboxOnAnUnreachablePortReportsOnlyACategory(t *testing.T) {
	fixture := newJobFixture(t)

	request := mailAccountRequest(t, senderAddress, senderPassword)
	request["smtpPort"] = 1
	created := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		channelAccountsPath(fixture.candidateID),
		request,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusCreated, created.Recorder.Code, string(created.Body))

	verified := fixture.verifyMailbox(
		t,
		helpers.DecodeData[outreachModel.ChannelAccount](t, created).ID,
	)

	require.NotNil(t, verified.LastError)
	assert.Equal(t, config.ErrMailServerUnreachable.Error(), *verified.LastError,
		"the dial error would reveal whether a port is open")
}

func TestMailboxRejectsInvalidSettings(t *testing.T) {
	fixture := newJobFixture(t)

	request := mailAccountRequest(t, "not-an-email", senderPassword)
	request["smtpSecurity"] = "SSL3"

	rejected := helpers.Do(
		t,
		fixture.router,
		http.MethodPost,
		channelAccountsPath(fixture.candidateID),
		request,
		bearer(fixture.userToken),
	)
	require.Equal(t, http.StatusUnprocessableEntity, rejected.Recorder.Code)
	helpers.AssertMatchesSpec(t, rejected)
}
