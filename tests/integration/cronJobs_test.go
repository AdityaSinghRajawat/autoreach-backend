//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/cronJobs"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (f jobFixture) cronManager(t *testing.T) *cronJobs.CronManager {
	t.Helper()

	manager, err := cronJobs.NewCronManager(cronJobs.CronManagerDependencies{
		RedisUtil:          f.deps.RedisUtil,
		AuthService:        f.deps.Services.Auth,
		JobService:         f.deps.Services.Job,
		CandidateService:   f.deps.Services.Candidate,
		OutreachService:    f.deps.Services.Outreach,
		ApplicationService: f.deps.Services.Application,
	})
	require.NoError(t, err)

	return manager
}

func TestACronClaimHoldsForItsWholeSlotOnEveryReplica(t *testing.T) {
	deps := newJobFixture(t).deps
	ctx := context.Background()
	slotKey := utils.BuildKey(
		config.GetKeyDelimiter(),
		config.GetKeyPrefix(),
		config.GetLockNamespace(),
		"cron:test",
	)
	claim := func() (bool, error) {
		return deps.RedisUtil.SetIfNotExists(ctx, slotKey, "claimed", 300*time.Millisecond)
	}

	first, err := claim()
	require.NoError(t, err)
	assert.True(t, first)

	second, err := claim()
	require.NoError(t, err)
	assert.False(t, second, "another replica's tick inside the slot skips the job")

	time.Sleep(350 * time.Millisecond)

	next, err := claim()
	require.NoError(t, err)
	assert.True(t, next, "the next slot can be claimed once this one expires")
}

func TestCronJobsStartAndStopCleanly(t *testing.T) {
	manager := newJobFixture(t).cronManager(t)

	ctx, cancel := context.WithCancel(context.Background())
	manager.StartJobs(ctx)
	cancel()

	stopped := make(chan struct{})

	go func() {
		manager.StopJobs()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("cron jobs did not stop")
	}
}

func TestTheDispatcherCronSendsDueMessages(t *testing.T) {
	fixture := newOutreachFixture(t)
	fixture.startCampaign(t)

	ctx := context.Background()

	_, customErr := fixture.deps.Services.Outreach.PlanOutreach(
		ctx,
		fixture.candidateID,
		fixture.jobID,
	)
	require.Nil(t, customErr)

	fixture.cronManager(t).RunOutreachDispatcher(ctx)

	messages := fixture.messages(t)
	require.NotEmpty(t, messages)
	assert.Equal(t, modelHelpers.OutreachStatusSending, messages[0].Status,
		"the first step is claimed and handed to the queue")
}

func TestTheRetentionAndReconcilerCronsRunAgainstLiveData(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	manager := fixture.cronManager(t)
	ctx := context.Background()

	_, err := fixture.deps.Postgres.GetPool().Exec(ctx, `
		UPDATE refresh_tokens SET expires_at = NOW() - INTERVAL '30 days'
		WHERE id = (SELECT id FROM refresh_tokens ORDER BY created_at LIMIT 1)
	`)
	require.NoError(t, err)

	manager.RunDataRetention(ctx)
	manager.RunStuckWorkReconciler(ctx)
	manager.RunReplySweeper(ctx)
	manager.RunJobSourcePoller(ctx)

	var expired int

	require.NoError(t, fixture.deps.Postgres.GetPool().QueryRow(ctx,
		`SELECT COUNT(*) FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '7 days'`,
	).Scan(&expired))
	assert.Zero(t, expired, "the retention run removed the long-expired session")
}
