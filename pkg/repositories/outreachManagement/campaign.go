package outreachManagement

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const (
	campaignColumns     = `id, candidate_id, channel_account_id, name, status, created_at, updated_at`
	campaignStepColumns = `campaign_id, step_number, template_id, delay_days`
)

type CampaignRepo struct {
	postgres       *db.PostgresConn
	campaignsTable string
	stepsTable     string
}

func NewCampaignRepo(postgres *db.PostgresConn) (*CampaignRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new campaign repo: %w", config.ErrMissingDependency)
	}

	return &CampaignRepo{
		postgres:       postgres,
		campaignsTable: config.GetTableName(config.TableOutreachCampaigns),
		stepsTable:     config.GetTableName(config.TableOutreachCampaignSteps),
	}, nil
}

func (r *CampaignRepo) Upsert(
	ctx context.Context,
	campaign *outreachModel.Campaign,
) (*outreachModel.Campaign, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.campaignsTable+` (`+campaignColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			channel_account_id = EXCLUDED.channel_account_id,
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`,
		campaign.ID, campaign.CandidateID, campaign.ChannelAccountID, campaign.Name,
		string(campaign.Status), campaign.CreatedAt, campaign.UpdatedAt,
	)
	if db.IsUniqueViolation(err) {
		return nil, fmt.Errorf("upsert campaign: %w", config.ErrActiveCampaignExists)
	}

	if err != nil {
		return nil, fmt.Errorf("upsert campaign: %w", err)
	}

	return campaign, nil
}

func (r *CampaignRepo) ReplaceSteps(
	ctx context.Context,
	campaignID string,
	steps []*outreachModel.CampaignStep,
) error {
	executor := r.postgres.Executor(ctx)

	if _, err := executor.Exec(
		ctx,
		`DELETE FROM `+r.stepsTable+` WHERE campaign_id = $1`,
		campaignID,
	); err != nil {
		return fmt.Errorf("replace campaign steps: %w", err)
	}

	batch := &pgx.Batch{}
	for _, step := range steps {
		batch.Queue(
			`INSERT INTO `+r.stepsTable+` (`+campaignStepColumns+`) VALUES ($1, $2, $3, $4)`,
			step.CampaignID,
			step.StepNumber,
			step.TemplateID,
			step.DelayDays,
		)
	}

	if err := executor.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("replace campaign steps: %w", err)
	}

	return nil
}

func (r *CampaignRepo) FindByCandidate(
	ctx context.Context,
	candidateID, campaignID string,
) (*outreachModel.Campaign, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+campaignColumns+` FROM `+r.campaignsTable+` WHERE candidate_id = $1 AND id = $2
	`, candidateID, campaignID)

	return utils.ScanOptionalRow(row, scanCampaign, "find campaign")
}

func (r *CampaignRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
) ([]*outreachModel.Campaign, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+campaignColumns+` FROM `+r.campaignsTable+`
		WHERE candidate_id = $1 ORDER BY created_at DESC
	`, candidateID)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}

	campaigns, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.Campaign, error) {
			return scanCampaign(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}

	return campaigns, nil
}

func (r *CampaignRepo) FindActive(
	ctx context.Context,
	candidateID string,
) (*outreachModel.Campaign, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+campaignColumns+` FROM `+r.campaignsTable+` WHERE candidate_id = $1 AND status = `+sqlCampaignStatusActive+`
	`, candidateID)

	return utils.ScanOptionalRow(row, scanCampaign, "find active campaign")
}

func (r *CampaignRepo) ListSteps(
	ctx context.Context,
	campaignIDs []string,
) ([]*outreachModel.CampaignStep, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+campaignStepColumns+` FROM `+r.stepsTable+`
		WHERE campaign_id = ANY($1) ORDER BY campaign_id, step_number
	`, campaignIDs)
	if err != nil {
		return nil, fmt.Errorf("list campaign steps: %w", err)
	}

	steps, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.CampaignStep, error) {
			step := outreachModel.CampaignStep{}
			err := row.Scan(&step.CampaignID, &step.StepNumber, &step.TemplateID, &step.DelayDays)

			return &step, err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list campaign steps: %w", err)
	}

	return steps, nil
}

func scanCampaign(row utils.RowScanner) (*outreachModel.Campaign, error) {
	campaign := outreachModel.Campaign{}

	err := row.Scan(
		&campaign.ID, &campaign.CandidateID, &campaign.ChannelAccountID, &campaign.Name,
		&campaign.Status, &campaign.CreatedAt, &campaign.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

// Runs before deleting a candidate: steps hold templates the cascade would delete first.
func (r *CampaignRepo) DeleteByCandidate(ctx context.Context, candidateID string) error {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.campaignsTable+` WHERE candidate_id = $1
	`, candidateID)
	if err != nil {
		return fmt.Errorf("delete campaigns of candidate: %w", err)
	}

	return nil
}
