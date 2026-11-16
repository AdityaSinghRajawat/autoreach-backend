package smartrecruiters

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
)

func (i *SmartRecruitersIntegration) GetPosting(
	ctx context.Context,
	companyID, postingID string,
) (jobTypes.BoardJob, error) {
	item := posting{}
	path := fmt.Sprintf(config.GetSmartRecruitersPostingsPath(), url.PathEscape(companyID)) +
		"/" + url.PathEscape(postingID)

	if err := i.httpClient.Get(ctx, path, nil, &item); err != nil {
		return jobTypes.BoardJob{}, fmt.Errorf("get smartrecruiters posting: %w", err)
	}

	return item.toBoardJob(), nil
}
