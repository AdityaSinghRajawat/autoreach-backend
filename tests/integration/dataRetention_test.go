//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpiredSessionsAreDeletedAfterTheGracePeriod(t *testing.T) {
	fixture := newJobFixture(t)
	ctx := context.Background()
	pool := fixture.deps.Postgres.GetPool()

	_, err := pool.Exec(ctx, `
		UPDATE refresh_tokens SET expires_at = NOW() - INTERVAL '30 days'
		WHERE id = (SELECT id FROM refresh_tokens ORDER BY created_at LIMIT 1)
	`)
	require.NoError(t, err)

	var before int

	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM refresh_tokens`).Scan(&before))

	deleted, customErr := fixture.deps.Services.Auth.PurgeExpiredRefreshTokens(ctx, 7*24*time.Hour)
	require.Nil(t, customErr)
	assert.Equal(t, 1, deleted, "only the session that expired weeks ago is removed")

	var after int

	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM refresh_tokens`).Scan(&after))
	assert.Equal(t, before-1, after)
}

func TestOutreachEventsAreKeptUnlessARetentionPeriodIsSet(t *testing.T) {
	fixture := newJobFixture(t)
	fixture.seedDashboard(t)

	ctx := context.Background()
	pool := fixture.deps.Postgres.GetPool()

	_, err := pool.Exec(ctx, `UPDATE outreach_events SET occurred_at = NOW() - INTERVAL '400 days'`)
	require.NoError(t, err)

	kept, customErr := fixture.deps.Services.Outreach.PurgeOutreachEvents(ctx, 0)
	require.Nil(t, customErr)
	assert.Zero(t, kept)

	deleted, customErr := fixture.deps.Services.Outreach.PurgeOutreachEvents(ctx, 365*24*time.Hour)
	require.Nil(t, customErr)
	assert.Positive(t, deleted)

	var remaining int

	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM outreach_events`).Scan(&remaining))
	assert.Zero(t, remaining)
}
