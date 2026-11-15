package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type EventRepo struct {
	postgres    *db.PostgresConn
	eventsTable string
}

func NewEventRepo(postgres *db.PostgresConn) (*EventRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new event repo: %w", config.ErrMissingDependency)
	}

	return &EventRepo{
		postgres:    postgres,
		eventsTable: config.GetTableName(config.TableOutreachEvents),
	}, nil
}

func (r *EventRepo) Insert(ctx context.Context, event *outreachModel.Event) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.eventsTable+` (id, message_id, event_type, metadata, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.ID, event.MessageID, string(event.EventType), string(event.Metadata), event.OccurredAt,
		event.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert outreach event: %w", err)
	}

	return nil
}
