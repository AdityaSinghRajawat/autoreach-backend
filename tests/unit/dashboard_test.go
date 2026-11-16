package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

func TestMetricsResponseReproducesTheSpecificationExample(t *testing.T) {
	setup.Config(t)

	counts := dashboardTypes.CandidateCounts{
		TotalJobsScraped:            45200,
		JobsMatched:                 128,
		ApplicationsSubmitted:       94,
		ContactsDiscovered:          156,
		ColdEmailsSent:              112,
		LinkedinRequestsSent:        84,
		LinkedinConnectionsAccepted: 24,
		RecruiterResponses:          18,
	}

	encoded, err := json.Marshal(counts.ToMetricsResponse("9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"))
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"candidateId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		"metrics": {
			"totalJobsScraped": 45200,
			"jobsMatched": 128,
			"applicationsSubmitted": 94,
			"contactsDiscovered": 156,
			"coldEmailsSent": 112,
			"linkedinRequestsSent": 84,
			"linkedinConnectionsAccepted": 24,
			"recruiterResponses": 18,
			"responseRatePercentage": 16.07
		}
	}`, string(encoded))

	empty := dashboardTypes.CandidateCounts{RecruiterResponses: 2}.ToMetricsResponse("x")
	assert.Zero(
		t,
		empty.Metrics.ResponseRatePercentage,
		"no emails sent means a zero rate, not a division by zero",
	)
}

func TestActivityCursorRoundTrips(t *testing.T) {
	setup.Config(t)

	cursor := dashboardTypes.ActivityCursor{
		OccurredAt: time.Date(2026, 11, 16, 9, 30, 0, 123456000, time.UTC),
		EventID:    "outreach:4d6f|odd",
	}

	parsed, valid := dashboardTypes.ParseActivityCursor(cursor.Encode())
	require.True(t, valid)
	assert.Equal(t, cursor, parsed)

	for _, invalid := range []string{"", "not base64!", "bm8tc2VwYXJhdG9y"} {
		_, valid := dashboardTypes.ParseActivityCursor(invalid)
		assert.False(t, valid, invalid)
	}
}
