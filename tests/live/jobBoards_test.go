//go:build live

package live

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const defaultLiveBoards = "GREENHOUSE:stripe,LEVER:palantir,ASHBY:openai,SMARTRECRUITERS:servicenow"

// Reads real public job boards (LIVE_JOB_BOARDS, as ATS:token pairs) and checks that every job maps
// to a valid posting. Run with: go test -tags=live ./tests/live/...
func TestLiveJobBoardsMapToValidPostings(t *testing.T) {
	t.Setenv("SMARTRECRUITERS_MAX_DETAILS_PER_POLL", "5")
	setup.Config(t)

	boards := os.Getenv("LIVE_JOB_BOARDS")
	if boards == "" {
		boards = defaultLiveBoards
	}

	registry := jobBoards.NewJobBoardRegistry(jobBoards.RegistryOptions{
		GreenhouseAPIURL:      config.GetGreenhouseBoardsAPIURL(),
		LeverAPIURL:           config.GetLeverPostingsAPIURL(),
		LeverEUAPIURL:         config.GetLeverEUPostingsAPIURL(),
		AshbyAPIURL:           config.GetAshbyJobBoardAPIURL(),
		SmartRecruitersAPIURL: config.GetSmartRecruitersAPIURL(),
	})

	for _, pair := range strings.Split(boards, ",") {
		atsType, token, found := strings.Cut(strings.TrimSpace(pair), ":")
		require.True(t, found, "LIVE_JOB_BOARDS entries look like ATS:token")

		t.Run(atsType+"/"+token, func(t *testing.T) {
			board := jobBoards.Find(registry, atsType)
			require.NotNil(t, board, "unknown ATS %s", atsType)

			source := jobModel.NewJobSource()
			source.ATSType = atsType
			source.BoardToken = strings.ToLower(token)
			source.CompanyName = token
			source.CompanyDomain = "example.com"

			jobs, err := board.ListJobs(context.Background(), source, nil)
			require.NoError(t, err)
			require.NotEmpty(t, jobs)

			invalid := 0

			for _, boardJob := range jobs {
				item := boardJob.ToIngestJobItem(source)
				job := item.ToJobPostingModel()
				job.JobSourceID = &source.ID

				if err := job.Validate(); err != nil {
					invalid++

					if invalid <= 3 {
						t.Logf("invalid job %s (%s): %v", boardJob.ExternalID, boardJob.URL, err)
					}

					continue
				}

				assert.NotContains(t, job.JobDescription, "&lt;", "description is still escaped")
				assert.NotContains(t, job.JobDescription, "<p>", "description still has markup")
			}

			t.Logf(
				"%d jobs, %d invalid; first: %q at %s",
				len(jobs),
				invalid,
				jobs[0].Title,
				jobs[0].URL,
			)
			assert.LessOrEqual(t, invalid*20, len(jobs), "more than 5%% of jobs failed validation")
		})
	}
}
