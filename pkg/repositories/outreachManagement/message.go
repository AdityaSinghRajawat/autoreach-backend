package outreachManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

const messageColumns = `id, candidate_id, contact_id, job_id, campaign_id, channel_account_id,
	template_id, step_number, channel, status, subject, body_content, linkedin_note, scheduled_at,
	provider_message_id, in_reply_to, tracking_token, attempt_count, sent_at, replied_at,
	failure_reason, created_at, updated_at`

type MessageRepo struct {
	postgres      *db.PostgresConn
	messagesTable string
	contactsTable string
}

func NewMessageRepo(postgres *db.PostgresConn) (*MessageRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new message repo: %w", config.ErrMissingDependency)
	}

	return &MessageRepo{
		postgres:      postgres,
		messagesTable: config.GetTableName(config.TableOutreachMessages),
		contactsTable: config.GetTableName(config.TableCompanyContacts),
	}, nil
}

func (r *MessageRepo) Upsert(
	ctx context.Context,
	message *outreachModel.Message,
) (*outreachModel.Message, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.messagesTable+` (`+messageColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23)
		ON CONFLICT (id) DO UPDATE SET
			channel_account_id = EXCLUDED.channel_account_id,
			status = EXCLUDED.status,
			subject = EXCLUDED.subject,
			body_content = EXCLUDED.body_content,
			scheduled_at = EXCLUDED.scheduled_at,
			provider_message_id = EXCLUDED.provider_message_id,
			in_reply_to = EXCLUDED.in_reply_to,
			tracking_token = EXCLUDED.tracking_token,
			attempt_count = EXCLUDED.attempt_count,
			sent_at = EXCLUDED.sent_at,
			replied_at = EXCLUDED.replied_at,
			failure_reason = EXCLUDED.failure_reason,
			updated_at = EXCLUDED.updated_at
	`, messageArguments(message)...)
	if db.IsUniqueViolation(err) {
		return nil, fmt.Errorf("upsert outreach message: %w", config.ErrDuplicateRecord)
	}

	if err != nil {
		return nil, fmt.Errorf("upsert outreach message: %w", err)
	}

	return message, nil
}

func (r *MessageRepo) InsertPlanned(
	ctx context.Context,
	messages []*outreachModel.Message,
) (int, error) {
	batch := &pgx.Batch{}
	for _, message := range messages {
		batch.Queue(`
			INSERT INTO `+r.messagesTable+` (`+messageColumns+`)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
				$19, $20, $21, $22, $23)
			ON CONFLICT ON CONSTRAINT unique_outreach_message_step DO NOTHING
		`, messageArguments(message)...)
	}

	results := r.postgres.Executor(ctx).SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()

	inserted := 0

	for range messages {
		tag, err := results.Exec()
		if err != nil {
			return inserted, fmt.Errorf("insert planned messages: %w", err)
		}

		inserted += int(tag.RowsAffected())
	}

	return inserted, nil
}

func (r *MessageRepo) FindByID(
	ctx context.Context,
	messageID string,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `id = $1`, messageID)
}

func (r *MessageRepo) FindByTrackingToken(
	ctx context.Context,
	token string,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `tracking_token = $1`, token)
}

func (r *MessageRepo) FindByCandidate(
	ctx context.Context,
	candidateID, messageID string,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `candidate_id = $1 AND id = $2`, candidateID, messageID)
}

func (r *MessageRepo) FindByProviderMessageIDs(
	ctx context.Context,
	providerMessageIDs []string,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `provider_message_id = ANY($1) ORDER BY step_number DESC LIMIT 1`,
		providerMessageIDs)
}

func (r *MessageRepo) FindLatestSentToAddress(
	ctx context.Context,
	channelAccountID, address string,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `channel_account_id = $1 AND sent_at IS NOT NULL AND contact_id IN (
			SELECT id FROM `+r.contactsTable+` WHERE LOWER(work_email) = LOWER($2)
		) ORDER BY sent_at DESC LIMIT 1`, channelAccountID, address)
}

func (r *MessageRepo) FindThreadRoot(
	ctx context.Context,
	message *outreachModel.Message,
) (*outreachModel.Message, error) {
	return r.findOne(ctx, `candidate_id = $1 AND contact_id = $2 AND job_id = $3 AND channel = $4
		AND step_number < $5 AND provider_message_id IS NOT NULL ORDER BY step_number LIMIT 1`,
		message.CandidateID, message.ContactID, message.JobID, string(message.Channel),
		message.StepNumber)
}

func (r *MessageRepo) HasContacted(
	ctx context.Context,
	candidateID, contactID string,
) (bool, error) {
	var exists bool

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM `+r.messagesTable+` WHERE candidate_id = $1 AND contact_id = $2)
	`, candidateID, contactID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check contacted: %w", err)
	}

	return exists, nil
}

func (r *MessageRepo) HasConversationEnded(
	ctx context.Context,
	candidateID, contactID string,
) (bool, error) {
	var ended bool

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM `+r.messagesTable+`
			WHERE candidate_id = $1 AND contact_id = $2 AND status IN ('REPLIED', 'BOUNCED'))
	`, candidateID, contactID).Scan(&ended)
	if err != nil {
		return false, fmt.Errorf("check conversation: %w", err)
	}

	return ended, nil
}

func (r *MessageRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
	filter outreachTypes.MessageFilter,
	page sharedTypes.PageParams,
) ([]*outreachModel.Message, int, error) {
	condition := `candidate_id = $1 AND ($2 = '' OR status::text = $2)`

	var total int

	if err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.messagesTable+` WHERE `+condition,
		candidateID, filter.Status,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count outreach messages: %w", err)
	}

	messages, err := r.list(ctx, condition+` ORDER BY created_at DESC, id LIMIT $3 OFFSET $4`,
		candidateID, filter.Status, page.PerPage, page.Offset())

	return messages, total, err
}

func (r *MessageRepo) ClaimDue(ctx context.Context, now time.Time, limit int) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		UPDATE `+r.messagesTable+` SET status = 'SENDING', attempt_count = attempt_count + 1,
			updated_at = $1
		WHERE id IN (
			SELECT id FROM `+r.messagesTable+`
			WHERE status = 'PENDING' AND scheduled_at <= $1 AND channel = 'EMAIL'
			ORDER BY scheduled_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id
	`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("claim due messages: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("claim due messages: %w", err)
	}

	return ids, nil
}

func (r *MessageRepo) ReleaseClaim(ctx context.Context, messageID string, now time.Time) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.messagesTable+` SET status = 'PENDING', attempt_count = attempt_count - 1,
			updated_at = $2
		WHERE id = $1 AND status = 'SENDING' AND sent_at IS NULL
	`, messageID, now)
	if err != nil {
		return fmt.Errorf("release message claim: %w", err)
	}

	return nil
}

// The advisory lock makes the daily cap hold across workers.
func (r *MessageRepo) ReserveSend(
	ctx context.Context,
	message *outreachModel.Message,
	dailyLimit int,
	now time.Time,
) (outreachTypes.SendReservation, error) {
	executor := r.postgres.Executor(ctx)

	if _, err := executor.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`,
		*message.ChannelAccountID); err != nil {
		return outreachTypes.SendReservation{}, fmt.Errorf("reserve send: %w", err)
	}

	var sentToday int

	if err := executor.QueryRow(ctx, `
		SELECT COUNT(*) FROM `+r.messagesTable+` WHERE channel_account_id = $1 AND sent_at > $2
	`, *message.ChannelAccountID, now.Add(-config.GetSendAttemptWindow())).Scan(&sentToday); err != nil {
		return outreachTypes.SendReservation{}, fmt.Errorf("reserve send: %w", err)
	}

	if sentToday >= dailyLimit {
		return outreachTypes.SendReservation{CapReached: true}, nil
	}

	tag, err := executor.Exec(ctx, `
		UPDATE `+r.messagesTable+` SET sent_at = $2, provider_message_id = $3, in_reply_to = $4,
			tracking_token = $5, updated_at = $2
		WHERE id = $1 AND status = 'SENDING' AND sent_at IS NULL
	`, message.ID, now, message.ProviderMessageID, message.InReplyTo, message.TrackingToken)
	if err != nil {
		return outreachTypes.SendReservation{}, fmt.Errorf("reserve send: %w", err)
	}

	return outreachTypes.SendReservation{Reserved: tag.RowsAffected() == 1}, nil
}

func (r *MessageRepo) OldestSendSince(
	ctx context.Context,
	channelAccountID string,
	since time.Time,
) (time.Time, error) {
	var oldest time.Time

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT MIN(sent_at) FROM `+r.messagesTable+` WHERE channel_account_id = $1 AND sent_at > $2
	`, channelAccountID, since).Scan(&oldest)
	if err != nil {
		return time.Time{}, fmt.Errorf("oldest send: %w", err)
	}

	return oldest, nil
}

func (r *MessageRepo) PauseForContact(
	ctx context.Context,
	candidateID, contactID string,
	now time.Time,
) error {
	return r.transition(ctx, `candidate_id = $2 AND contact_id = $3 AND status = 'PENDING'`,
		"PAUSED", now, candidateID, contactID)
}

func (r *MessageRepo) CancelForContact(
	ctx context.Context,
	candidateID, contactID string,
	now time.Time,
) error {
	return r.transition(
		ctx,
		`candidate_id = $2 AND contact_id = $3 AND status IN ('PENDING', 'PAUSED')`,
		"CANCELLED",
		now,
		candidateID,
		contactID,
	)
}

func (r *MessageRepo) SetCampaignPaused(
	ctx context.Context,
	campaignID string,
	now time.Time,
) error {
	return r.transition(ctx, `campaign_id = $2 AND status = 'PENDING'`, "PAUSED", now, campaignID)
}

func (r *MessageRepo) SetCampaignResumed(
	ctx context.Context,
	campaignID string,
	now time.Time,
) error {
	return r.transition(ctx, `campaign_id = $2 AND status = 'PAUSED' AND NOT EXISTS (
			SELECT 1 FROM `+r.messagesTable+` AS ended
			WHERE ended.candidate_id = `+r.messagesTable+`.candidate_id
				AND ended.contact_id = `+r.messagesTable+`.contact_id
				AND ended.status IN ('REPLIED', 'BOUNCED'))`, "PENDING", now, campaignID)
}

func (r *MessageRepo) SetCampaignCancelled(
	ctx context.Context,
	campaignID string,
	now time.Time,
) error {
	return r.transition(ctx, `campaign_id = $2 AND status IN ('PENDING', 'PAUSED')`,
		"CANCELLED", now, campaignID)
}

func (r *MessageRepo) FailStaleSending(
	ctx context.Context,
	staleBefore, now time.Time,
) (int, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.messagesTable+`
		SET status = CASE WHEN sent_at IS NULL THEN 'PENDING'::outreach_status ELSE 'FAILED'::outreach_status END,
			failure_reason = CASE WHEN sent_at IS NULL THEN failure_reason ELSE $3 END,
			updated_at = $2
		WHERE status = 'SENDING' AND updated_at < $1
	`, staleBefore, now, config.ErrSendOutcomeUnknown.Error())
	if err != nil {
		return 0, fmt.Errorf("fail stale sending messages: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

func (r *MessageRepo) transition(
	ctx context.Context,
	condition, status string,
	now time.Time,
	arguments ...any,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.messagesTable+` SET status = '`+status+`', updated_at = $1 WHERE `+condition,
		append([]any{now}, arguments...)...)
	if err != nil {
		return fmt.Errorf("update outreach messages to %s: %w", status, err)
	}

	return nil
}

func (r *MessageRepo) findOne(
	ctx context.Context,
	condition string,
	arguments ...any,
) (*outreachModel.Message, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+messageColumns+` FROM `+r.messagesTable+` WHERE `+condition, arguments...)

	return db.FindOne(row, scanMessage, "find outreach message")
}

func (r *MessageRepo) list(
	ctx context.Context,
	condition string,
	arguments ...any,
) ([]*outreachModel.Message, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+messageColumns+` FROM `+r.messagesTable+` WHERE `+condition, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list outreach messages: %w", err)
	}

	messages, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.Message, error) {
			return scanMessage(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list outreach messages: %w", err)
	}

	return messages, nil
}

func messageArguments(message *outreachModel.Message) []any {
	return []any{
		message.ID, message.CandidateID, message.ContactID, message.JobID, message.CampaignID,
		message.ChannelAccountID, message.TemplateID, message.StepNumber, string(message.Channel),
		string(message.Status), message.Subject, message.BodyContent, message.LinkedinNote,
		message.ScheduledAt, message.ProviderMessageID, message.InReplyTo, message.TrackingToken,
		message.AttemptCount, message.SentAt, message.RepliedAt, message.FailureReason,
		message.CreatedAt, message.UpdatedAt,
	}
}

func scanMessage(row db.RowScanner) (*outreachModel.Message, error) {
	message := outreachModel.Message{}

	err := row.Scan(
		&message.ID,
		&message.CandidateID,
		&message.ContactID,
		&message.JobID,
		&message.CampaignID,
		&message.ChannelAccountID,
		&message.TemplateID,
		&message.StepNumber,
		&message.Channel,
		&message.Status,
		&message.Subject,
		&message.BodyContent,
		&message.LinkedinNote,
		&message.ScheduledAt,
		&message.ProviderMessageID,
		&message.InReplyTo,
		&message.TrackingToken,
		&message.AttemptCount,
		&message.SentAt,
		&message.RepliedAt,
		&message.FailureReason,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
