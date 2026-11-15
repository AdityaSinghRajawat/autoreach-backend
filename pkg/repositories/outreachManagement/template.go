package outreachManagement

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

const templateColumns = `id, candidate_id, name, channel, subject, body, status, approved_at,
	created_at, updated_at`

type TemplateRepo struct {
	postgres       *db.PostgresConn
	templatesTable string
}

func NewTemplateRepo(postgres *db.PostgresConn) (*TemplateRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new template repo: %w", config.ErrMissingDependency)
	}

	return &TemplateRepo{
		postgres:       postgres,
		templatesTable: config.GetTableName(config.TableOutreachTemplates),
	}, nil
}

func (r *TemplateRepo) Upsert(
	ctx context.Context,
	template *outreachModel.Template,
) (*outreachModel.Template, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.templatesTable+` (`+templateColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			subject = EXCLUDED.subject,
			body = EXCLUDED.body,
			status = EXCLUDED.status,
			approved_at = EXCLUDED.approved_at,
			updated_at = EXCLUDED.updated_at
	`,
		template.ID, template.CandidateID, template.Name, string(template.Channel),
		template.Subject, template.Body, string(template.Status), template.ApprovedAt,
		template.CreatedAt, template.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert template: %w", err)
	}

	return template, nil
}

func (r *TemplateRepo) FindByCandidate(
	ctx context.Context,
	candidateID, templateID string,
) (*outreachModel.Template, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+templateColumns+` FROM `+r.templatesTable+` WHERE candidate_id = $1 AND id = $2
	`, candidateID, templateID)

	return db.FindOne(row, scanTemplate, "find template")
}

func (r *TemplateRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
) ([]*outreachModel.Template, error) {
	return r.list(ctx, `WHERE candidate_id = $1 ORDER BY created_at DESC`, candidateID)
}

func (r *TemplateRepo) ListByIDs(
	ctx context.Context,
	candidateID string,
	templateIDs []string,
) ([]*outreachModel.Template, error) {
	return r.list(ctx, `WHERE candidate_id = $1 AND id = ANY($2)`, candidateID, templateIDs)
}

func (r *TemplateRepo) Delete(ctx context.Context, candidateID, templateID string) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		DELETE FROM `+r.templatesTable+` WHERE candidate_id = $1 AND id = $2
	`, candidateID, templateID)
	if db.IsForeignKeyViolation(err) {
		return false, fmt.Errorf("delete template: %w", config.ErrTemplateInUse)
	}

	if err != nil {
		return false, fmt.Errorf("delete template: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *TemplateRepo) list(
	ctx context.Context,
	condition string,
	arguments ...any,
) ([]*outreachModel.Template, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+templateColumns+` FROM `+r.templatesTable+` `+condition, arguments...)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	templates, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*outreachModel.Template, error) {
			return scanTemplate(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	return templates, nil
}

func scanTemplate(row db.RowScanner) (*outreachModel.Template, error) {
	template := outreachModel.Template{}

	err := row.Scan(
		&template.ID, &template.CandidateID, &template.Name, &template.Channel, &template.Subject,
		&template.Body, &template.Status, &template.ApprovedAt, &template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &template, nil
}
