package outreachManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

const channelAccountColumns = `id, candidate_id, channel_type, identifier, display_name, smtp_host,
	smtp_port, smtp_security, imap_host, imap_port, imap_security, encrypted_credentials, key_version,
	status, daily_limit, last_verified_at, last_error, imap_uid_validity, imap_last_uid, created_at,
	updated_at`

type ChannelAccountRepo struct {
	postgres             *db.PostgresConn
	channelAccountsTable string
}

func NewChannelAccountRepo(postgres *db.PostgresConn) (*ChannelAccountRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new channel account repo: %w", config.ErrMissingDependency)
	}

	return &ChannelAccountRepo{
		postgres:             postgres,
		channelAccountsTable: config.GetTableName(config.TableChannelAccounts),
	}, nil
}

func (r *ChannelAccountRepo) Upsert(
	ctx context.Context,
	account *outreachModel.ChannelAccount,
) (*outreachModel.ChannelAccount, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.channelAccountsTable+` (`+channelAccountColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21)
		ON CONFLICT (id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			encrypted_credentials = EXCLUDED.encrypted_credentials,
			key_version = EXCLUDED.key_version,
			status = EXCLUDED.status,
			daily_limit = EXCLUDED.daily_limit,
			last_verified_at = EXCLUDED.last_verified_at,
			last_error = EXCLUDED.last_error,
			imap_uid_validity = EXCLUDED.imap_uid_validity,
			imap_last_uid = EXCLUDED.imap_last_uid,
			updated_at = EXCLUDED.updated_at
	`,
		account.ID, account.CandidateID, string(account.ChannelType), account.EmailAddress,
		account.DisplayName, account.SMTPHost, account.SMTPPort, string(account.SMTPSecurity),
		account.IMAPHost, account.IMAPPort, string(account.IMAPSecurity),
		account.EncryptedCredentials, account.KeyVersion, string(account.Status),
		account.DailyLimit, account.LastVerifiedAt, account.LastError, account.IMAPUIDValidity,
		account.IMAPLastUID, account.CreatedAt, account.UpdatedAt,
	)

	if db.IsUniqueViolation(err) {
		return nil, fmt.Errorf("upsert channel account: %w", config.ErrDuplicateRecord)
	}

	if err != nil {
		return nil, fmt.Errorf("upsert channel account: %w", err)
	}

	return account, nil
}

func (r *ChannelAccountRepo) FindByCandidate(
	ctx context.Context,
	candidateID, accountID string,
) (*outreachModel.ChannelAccount, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+channelAccountColumns+` FROM `+r.channelAccountsTable+`
		WHERE candidate_id = $1 AND id = $2
	`, candidateID, accountID)

	return db.FindOne(row, scanChannelAccount, "find channel account")
}

func (r *ChannelAccountRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
) ([]*outreachModel.ChannelAccount, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+channelAccountColumns+` FROM `+r.channelAccountsTable+`
		WHERE candidate_id = $1
		ORDER BY created_at
	`, candidateID)
	if err != nil {
		return nil, fmt.Errorf("list channel accounts: %w", err)
	}

	accounts, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.ChannelAccount, error) {
			return scanChannelAccount(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list channel accounts: %w", err)
	}

	return accounts, nil
}

func (r *ChannelAccountRepo) FindByID(
	ctx context.Context,
	accountID string,
) (*outreachModel.ChannelAccount, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+channelAccountColumns+` FROM `+r.channelAccountsTable+` WHERE id = $1
	`, accountID)

	return db.FindOne(row, scanChannelAccount, "find channel account")
}

func (r *ChannelAccountRepo) FindFirstActive(
	ctx context.Context,
	candidateID string,
) (*outreachModel.ChannelAccount, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+channelAccountColumns+` FROM `+r.channelAccountsTable+`
		WHERE candidate_id = $1 AND status = `+sqlChannelAccountStatusActive+` ORDER BY created_at LIMIT 1
	`, candidateID)

	return db.FindOne(row, scanChannelAccount, "find active channel account")
}

func (r *ChannelAccountRepo) ListActiveIDs(ctx context.Context) ([]string, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT id FROM `+r.channelAccountsTable+` WHERE status = `+sqlChannelAccountStatusActive+` ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list active channel accounts: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list active channel accounts: %w", err)
	}

	return ids, nil
}

func (r *ChannelAccountRepo) Delete(
	ctx context.Context,
	candidateID, accountID string,
) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.channelAccountsTable+` WHERE candidate_id = $1 AND id = $2
	`, candidateID, accountID)
	if err != nil {
		return false, fmt.Errorf("delete channel account: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func scanChannelAccount(row db.RowScanner) (*outreachModel.ChannelAccount, error) {
	account := outreachModel.ChannelAccount{}

	err := row.Scan(
		&account.ID, &account.CandidateID, &account.ChannelType, &account.EmailAddress,
		&account.DisplayName, &account.SMTPHost, &account.SMTPPort, &account.SMTPSecurity,
		&account.IMAPHost, &account.IMAPPort, &account.IMAPSecurity,
		&account.EncryptedCredentials, &account.KeyVersion, &account.Status, &account.DailyLimit,
		&account.LastVerifiedAt, &account.LastError, &account.IMAPUIDValidity,
		&account.IMAPLastUID, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// AdvanceCursor records how far the inbox was read. It touches nothing else, and nothing at all once
// the mailbox was paused or deleted while it was being read.
func (r *ChannelAccountRepo) AdvanceCursor(
	ctx context.Context,
	accountID string,
	uidValidity, lastUID int64,
	now time.Time,
) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.channelAccountsTable+`
		SET imap_uid_validity = $2, imap_last_uid = $3, last_verified_at = $4, last_error = NULL, updated_at = $4
		WHERE id = $1 AND status = `+sqlChannelAccountStatusActive+`
	`, accountID, uidValidity, lastUID, now)
	if err != nil {
		return fmt.Errorf("advance mailbox cursor: %w", err)
	}

	return nil
}
