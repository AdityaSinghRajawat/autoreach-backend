package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

const companyDomainColumns = `normalized_name, company_name, company_domain, provider, created_at, updated_at`

type CompanyDomainRepo struct {
	postgres            *db.PostgresConn
	companyDomainsTable string
}

func NewCompanyDomainRepo(postgres *db.PostgresConn) (*CompanyDomainRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new company domain repo: %w", config.ErrMissingDependency)
	}

	return &CompanyDomainRepo{
		postgres:            postgres,
		companyDomainsTable: config.GetTableName(config.TableCompanyDomains),
	}, nil
}

func (r *CompanyDomainRepo) ListByNormalizedNames(
	ctx context.Context,
	normalizedNames []string,
) ([]*enrichmentModel.CompanyDomain, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, `
		SELECT `+companyDomainColumns+` FROM `+r.companyDomainsTable+` WHERE normalized_name = ANY($1)
	`, normalizedNames)
	if err != nil {
		return nil, fmt.Errorf("list company domains: %w", err)
	}

	domains, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*enrichmentModel.CompanyDomain, error) {
			domain := enrichmentModel.CompanyDomain{}
			err := row.Scan(
				&domain.NormalizedName, &domain.CompanyName, &domain.Domain, &domain.Provider,
				&domain.CreatedAt, &domain.UpdatedAt,
			)

			return &domain, err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list company domains: %w", err)
	}

	return domains, nil
}

func (r *CompanyDomainRepo) Upsert(
	ctx context.Context,
	domain *enrichmentModel.CompanyDomain,
) (*enrichmentModel.CompanyDomain, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.companyDomainsTable+` (`+companyDomainColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (normalized_name) DO UPDATE SET
			company_name = EXCLUDED.company_name,
			company_domain = EXCLUDED.company_domain,
			provider = EXCLUDED.provider,
			updated_at = EXCLUDED.updated_at
	`,
		domain.NormalizedName, domain.CompanyName, domain.Domain, domain.Provider,
		domain.CreatedAt, domain.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert company domain: %w", err)
	}

	return domain, nil
}
