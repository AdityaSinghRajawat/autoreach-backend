package smartrecruiters

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
)

// ListPostings returns every public posting without its description, which only GetPosting has.
func (i *SmartRecruitersIntegration) ListPostings(
	ctx context.Context,
	companyID string,
) ([]jobTypes.BoardJob, error) {
	jobs := []jobTypes.BoardJob{}
	path := fmt.Sprintf(config.GetSmartRecruitersPostingsPath(), url.PathEscape(companyID))

	for offset := 0; ; {
		page := postingsPage{}
		query := "?limit=" + strconv.Itoa(
			config.GetSmartRecruitersPageSize(),
		) + "&offset=" + strconv.Itoa(
			offset,
		)

		if err := i.httpClient.Get(ctx, path+query, nil, &page); err != nil {
			return nil, fmt.Errorf("list smartrecruiters postings: %w", err)
		}

		for _, item := range page.Content {
			jobs = append(jobs, item.toBoardJob())
		}

		offset += len(page.Content)
		if len(page.Content) == 0 || offset >= page.TotalFound {
			return jobs, nil
		}
	}
}
