package dashboardManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
)

func (s *DashboardService) ListActivity(
	ctx context.Context,
	candidateID string,
	params dashboardTypes.ListActivityParams,
) (*dashboardTypes.ActivityPage, *config.CustomError) {
	if candidateID == "" {
		return nil, candidateNotFound()
	}

	limit := params.Limit
	if limit == 0 {
		limit = config.GetActivityDefaultLimit()
	}

	var before *dashboardTypes.ActivityCursor

	if params.Before != "" {
		cursor, valid := dashboardTypes.ParseActivityCursor(params.Before)
		if !valid {
			return nil, invalidCursor()
		}

		before = &cursor
	}

	items, err := s.activityEventRepo.ListBefore(ctx, candidateID, before, limit)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list activity: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	page := &dashboardTypes.ActivityPage{Items: items}
	if len(items) == limit {
		nextCursor := items[len(items)-1].Cursor().Encode()
		page.NextCursor = &nextCursor
	}

	return page, nil
}

func invalidCursor() *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("activity: %w", config.ErrInvalidCursor),
		config.GetErrCodeBadRequest(),
	)
}
