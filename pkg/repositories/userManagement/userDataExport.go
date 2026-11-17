package userManagement

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

// UserDataExportRepo reads everything a user's candidate profile owns, each table joined through
// the candidate so one indexed lookup on user_id drives every query.
type UserDataExportRepo struct {
	postgres        *db.PostgresConn
	candidatesTable string
}

func NewUserDataExportRepo(postgres *db.PostgresConn) (*UserDataExportRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new user data export repo: %w", config.ErrMissingDependency)
	}

	return &UserDataExportRepo{
		postgres:        postgres,
		candidatesTable: config.GetTableName(config.TableCandidates),
	}, nil
}

func (r *UserDataExportRepo) ListResumes(
	ctx context.Context,
	userID string,
) ([]*candidateModel.Resume, error) {
	return listOwnedByUser[candidateModel.Resume](ctx, r, config.TableCandidateResumes, userID)
}

func (r *UserDataExportRepo) ListJobMatches(
	ctx context.Context,
	userID string,
) ([]*matchingModel.JobMatch, error) {
	return listOwnedByUser[matchingModel.JobMatch](ctx, r, config.TableJobMatches, userID)
}

func (r *UserDataExportRepo) ListApplications(
	ctx context.Context,
	userID string,
) ([]*applicationModel.Application, error) {
	return listOwnedByUser[applicationModel.Application](ctx, r, config.TableApplications, userID)
}

func (r *UserDataExportRepo) ListChannelAccounts(
	ctx context.Context,
	userID string,
) ([]*outreachModel.ChannelAccount, error) {
	return listOwnedByUser[outreachModel.ChannelAccount](
		ctx,
		r,
		config.TableChannelAccounts,
		userID,
	)
}

func (r *UserDataExportRepo) ListOutreachTemplates(
	ctx context.Context,
	userID string,
) ([]*outreachModel.Template, error) {
	return listOwnedByUser[outreachModel.Template](ctx, r, config.TableOutreachTemplates, userID)
}

func (r *UserDataExportRepo) ListOutreachCampaigns(
	ctx context.Context,
	userID string,
) ([]*outreachModel.Campaign, error) {
	return listOwnedByUser[outreachModel.Campaign](ctx, r, config.TableOutreachCampaigns, userID)
}

func (r *UserDataExportRepo) ListOutreachMessages(
	ctx context.Context,
	userID string,
) ([]*outreachModel.Message, error) {
	return listOwnedByUser[outreachModel.Message](ctx, r, config.TableOutreachMessages, userID)
}

func (r *UserDataExportRepo) ListSuppressions(
	ctx context.Context,
	userID string,
) ([]*outreachModel.SuppressionEntry, error) {
	return listOwnedByUser[outreachModel.SuppressionEntry](
		ctx,
		r,
		config.TableSuppressionEntries,
		userID,
	)
}

func (r *UserDataExportRepo) ListContactSignals(
	ctx context.Context,
	userID string,
) ([]*enrichmentModel.ContactSignal, error) {
	return listOwnedByUser[enrichmentModel.ContactSignal](
		ctx,
		r,
		config.TableContactSignals,
		userID,
	)
}

func listOwnedByUser[T any](
	ctx context.Context,
	r *UserDataExportRepo,
	table config.Table,
	userID string,
) ([]*T, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+db.Columns[T]("owned")+`
		FROM `+config.GetTableName(table)+` AS owned
		JOIN `+r.candidatesTable+` AS candidate ON candidate.id = owned.candidate_id
		WHERE candidate.user_id = $1
		ORDER BY owned.created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", table, err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", table, err)
	}

	return records, nil
}
