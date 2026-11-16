package dashboardManagement

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
)

type ActivityEventRepo struct {
	postgres             *db.PostgresConn
	activityEventsTable  string
	jobPostingsTable     string
	companyContactsTable string
}

func NewActivityEventRepo(postgres *db.PostgresConn) (*ActivityEventRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new activity event repo: %w", config.ErrMissingDependency)
	}

	return &ActivityEventRepo{
		postgres:             postgres,
		activityEventsTable:  config.GetTableName(config.TableActivityEvents),
		jobPostingsTable:     config.GetTableName(config.TableJobPostings),
		companyContactsTable: config.GetTableName(config.TableCompanyContacts),
	}, nil
}

// ListBefore returns the newest events older than the cursor (all events when it is nil).
func (r *ActivityEventRepo) ListBefore(
	ctx context.Context,
	candidateID string,
	before *dashboardTypes.ActivityCursor,
	limit int,
) ([]dashboardTypes.ActivityItem, error) {
	if before == nil {
		return r.list(
			ctx,
			`WHERE a.candidate_id = $1 ORDER BY a.occurred_at DESC, a.event_id DESC LIMIT $2`,
			candidateID,
			limit,
		)
	}

	return r.list(ctx, `
		WHERE a.candidate_id = $1 AND (a.occurred_at, a.event_id) < ($2, $3)
		ORDER BY a.occurred_at DESC, a.event_id DESC LIMIT $4
	`, candidateID, before.OccurredAt, before.EventID, limit)
}

// ListAfter returns the oldest events newer than the cursor, for streaming.
func (r *ActivityEventRepo) ListAfter(
	ctx context.Context,
	candidateID string,
	after dashboardTypes.ActivityCursor,
	limit int,
) ([]dashboardTypes.ActivityItem, error) {
	return r.list(ctx, `
		WHERE a.candidate_id = $1 AND (a.occurred_at, a.event_id) > ($2, $3)
		ORDER BY a.occurred_at, a.event_id LIMIT $4
	`, candidateID, after.OccurredAt, after.EventID, limit)
}

func (r *ActivityEventRepo) list(
	ctx context.Context,
	filter string,
	arguments ...any,
) ([]dashboardTypes.ActivityItem, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT a.event_id, a.event_type, a.channel, a.occurred_at, a.job_id, j.job_title, j.company_name,
			a.contact_id, contact.full_name, a.application_id, a.message_id
		FROM `+r.activityEventsTable+` a
		JOIN `+r.jobPostingsTable+` j ON j.id = a.job_id
		LEFT JOIN `+r.companyContactsTable+` contact ON contact.id = a.contact_id
		`+filter, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list activity: %w", err)
	}

	items, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (dashboardTypes.ActivityItem, error) {
			item := dashboardTypes.ActivityItem{}
			err := row.Scan(
				&item.EventID,
				&item.EventType,
				&item.Channel,
				&item.OccurredAt,
				&item.JobID,
				&item.JobTitle,
				&item.CompanyName,
				&item.ContactID,
				&item.ContactName,
				&item.ApplicationID,
				&item.MessageID,
			)

			return item, err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list activity: %w", err)
	}

	return items, nil
}
