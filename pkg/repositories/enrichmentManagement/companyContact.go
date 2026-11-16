package enrichmentManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

const (
	companyContactColumns = `id, company_domain, company_name, full_name, role_type, title,
	work_email, phone_number, linkedin_url, enrichment_provider, provider_person_id,
	is_email_verified, email_verification_status, enrichment_metadata, created_at, updated_at,
	last_interaction_at, email_confirmed_at, linkedin_confirmed_at, anonymized_at`

	qualifiedContactColumns = `contact.id, contact.company_domain, contact.company_name,
	contact.full_name, contact.role_type::text, contact.title, contact.work_email,
	contact.phone_number, contact.linkedin_url, contact.enrichment_provider,
	contact.provider_person_id, contact.is_email_verified, contact.email_verification_status,
	contact.enrichment_metadata, contact.created_at, contact.updated_at, contact.last_interaction_at,
	contact.email_confirmed_at, contact.linkedin_confirmed_at, contact.anonymized_at`
)

type CompanyContactRepo struct {
	postgres              *db.PostgresConn
	companyContactsTable  string
	jobContactsTable      string
	jobMatchesTable       string
	outreachMessagesTable string
}

func NewCompanyContactRepo(postgres *db.PostgresConn) (*CompanyContactRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new company contact repo: %w", config.ErrMissingDependency)
	}

	return &CompanyContactRepo{
		postgres:              postgres,
		companyContactsTable:  config.GetTableName(config.TableCompanyContacts),
		jobContactsTable:      config.GetTableName(config.TableJobContacts),
		jobMatchesTable:       config.GetTableName(config.TableJobMatches),
		outreachMessagesTable: config.GetTableName(config.TableOutreachMessages),
	}, nil
}

func (r *CompanyContactRepo) Upsert(
	ctx context.Context,
	contact *enrichmentModel.CompanyContact,
) (*enrichmentModel.CompanyContact, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.companyContactsTable+` (`+companyContactColumns+`)
		VALUES (
			$1, $2, $3, $4, $5::contact_role, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb,
			$15, $16, $17, $18, $19, $20
		)
		ON CONFLICT (id) DO UPDATE SET
			company_domain = EXCLUDED.company_domain,
			company_name = EXCLUDED.company_name,
			full_name = EXCLUDED.full_name,
			role_type = EXCLUDED.role_type,
			title = EXCLUDED.title,
			work_email = EXCLUDED.work_email,
			phone_number = EXCLUDED.phone_number,
			linkedin_url = EXCLUDED.linkedin_url,
			enrichment_provider = EXCLUDED.enrichment_provider,
			provider_person_id = EXCLUDED.provider_person_id,
			is_email_verified = EXCLUDED.is_email_verified,
			email_verification_status = EXCLUDED.email_verification_status,
			enrichment_metadata = EXCLUDED.enrichment_metadata,
			last_interaction_at = EXCLUDED.last_interaction_at,
			email_confirmed_at = EXCLUDED.email_confirmed_at,
			linkedin_confirmed_at = EXCLUDED.linkedin_confirmed_at,
			anonymized_at = EXCLUDED.anonymized_at,
			updated_at = EXCLUDED.updated_at
	`,
		contact.ID, contact.CompanyDomain, contact.CompanyName, contact.FullName,
		string(contact.RoleType), contact.Title, contact.WorkEmail, contact.PhoneNumber,
		contact.LinkedinURL, contact.EnrichmentProvider, contact.ProviderPersonID,
		contact.IsEmailVerified, string(contact.EmailVerificationStatus),
		string(contact.EnrichmentMetadata), contact.CreatedAt, contact.UpdatedAt,
		contact.LastInteractionAt, contact.EmailConfirmedAt, contact.LinkedinConfirmedAt,
		contact.AnonymizedAt,
	)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, fmt.Errorf("upsert company contact: %w", config.ErrDuplicateRecord)
		}

		return nil, fmt.Errorf("upsert company contact: %w", err)
	}

	return contact, nil
}

func (r *CompanyContactRepo) FindByID(
	ctx context.Context,
	contactID string,
) (*enrichmentModel.CompanyContact, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE contact.id = $1
		LIMIT 1
	`, contactID)

	return db.FindOne(row, scanCompanyContact, "find company contact")
}

func (r *CompanyContactRepo) ListByIDs(
	ctx context.Context,
	contactIDs []string,
) ([]*enrichmentModel.CompanyContact, error) {
	return r.list(ctx, "list contacts by id", `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE contact.id = ANY($1)
	`, contactIDs)
}

func (r *CompanyContactRepo) FindExisting(
	ctx context.Context,
	contact *enrichmentModel.CompanyContact,
) (*enrichmentModel.CompanyContact, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE (contact.enrichment_provider = $1 AND contact.provider_person_id = $2)
			OR LOWER(contact.linkedin_url) = LOWER($3)
			OR (contact.company_domain = $4 AND LOWER(contact.work_email) = LOWER($5))
		ORDER BY contact.created_at
		LIMIT 1
	`,
		contact.EnrichmentProvider, contact.ProviderPersonID, contact.LinkedinURL,
		contact.CompanyDomain, contact.WorkEmail,
	)

	return db.FindOne(row, scanCompanyContact, "find existing company contact")
}

func (r *CompanyContactRepo) ListFreshByCompany(
	ctx context.Context,
	companyDomain string,
	since time.Time,
) ([]*enrichmentModel.CompanyContact, error) {
	return r.list(ctx, "list fresh company contacts", `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE contact.company_domain = $1 AND contact.updated_at >= $2 AND contact.anonymized_at IS NULL
		ORDER BY contact.is_email_verified DESC, contact.updated_at DESC
	`, companyDomain, since)
}

func (r *CompanyContactRepo) ListByJob(
	ctx context.Context,
	jobID string,
) ([]*enrichmentModel.CompanyContact, error) {
	return r.list(ctx, "list job contacts", `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		JOIN `+r.jobContactsTable+` AS link ON link.contact_id = contact.id
		WHERE link.job_id = $1
		ORDER BY contact.is_email_verified DESC, contact.full_name
	`, jobID)
}

func (r *CompanyContactRepo) ListByCandidate(
	ctx context.Context,
	candidateID string,
	page sharedTypes.PageParams,
) ([]*enrichmentModel.CompanyContact, error) {
	return r.list(ctx, "list candidate contacts", `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE contact.id IN (`+r.candidateContactIDs()+`)
		ORDER BY contact.created_at DESC, contact.id
		LIMIT $2 OFFSET $3
	`, candidateID, page.PerPage, page.Offset())
}

func (r *CompanyContactRepo) CountByCandidate(
	ctx context.Context,
	candidateID string,
) (int, error) {
	var total int

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM (`+r.candidateContactIDs()+`) AS contact_ids
	`, candidateID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count candidate contacts: %w", err)
	}

	return total, nil
}

func (r *CompanyContactRepo) LinkToJob(
	ctx context.Context,
	jobID string,
	contactIDs []string,
) error {
	if len(contactIDs) == 0 {
		return nil
	}

	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.jobContactsTable+` (job_id, contact_id)
		SELECT $1, contact_id FROM UNNEST($2::uuid[]) AS contact_id
		ON CONFLICT (job_id, contact_id) DO NOTHING
	`, jobID, contactIDs)
	if err != nil {
		return fmt.Errorf("link contacts to job: %w", err)
	}

	return nil
}

func (r *CompanyContactRepo) candidateContactIDs() string {
	return `
		SELECT DISTINCT link.contact_id
		FROM ` + r.jobContactsTable + ` AS link
		JOIN ` + r.jobMatchesTable + ` AS match ON match.job_id = link.job_id
		WHERE match.candidate_id = $1`
}

func (r *CompanyContactRepo) list(
	ctx context.Context,
	operation, statement string,
	arguments ...any,
) ([]*enrichmentModel.CompanyContact, error) {
	rows, err := r.postgres.Executor(ctx).Query(ctx, statement, arguments...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	contacts, err := pgx.CollectRows(
		rows,
		func(row pgx.CollectableRow) (*enrichmentModel.CompanyContact, error) {
			return scanCompanyContact(row)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return contacts, nil
}

func scanCompanyContact(row db.RowScanner) (*enrichmentModel.CompanyContact, error) {
	contact := enrichmentModel.CompanyContact{}

	var metadata []byte

	err := row.Scan(
		&contact.ID,
		&contact.CompanyDomain,
		&contact.CompanyName,
		&contact.FullName,
		&contact.RoleType,
		&contact.Title,
		&contact.WorkEmail,
		&contact.PhoneNumber,
		&contact.LinkedinURL,
		&contact.EnrichmentProvider,
		&contact.ProviderPersonID,
		&contact.IsEmailVerified,
		&contact.EmailVerificationStatus,
		&metadata,
		&contact.CreatedAt,
		&contact.UpdatedAt,
		&contact.LastInteractionAt,
		&contact.EmailConfirmedAt,
		&contact.LinkedinConfirmedAt,
		&contact.AnonymizedAt,
	)
	if err != nil {
		return nil, err
	}

	contact.EnrichmentMetadata = metadata

	return &contact, nil
}

// ListByIdentity returns every contact with the email address or LinkedIn URL.
func (r *CompanyContactRepo) ListByIdentity(
	ctx context.Context,
	email, linkedinURL string,
) ([]*enrichmentModel.CompanyContact, error) {
	return r.list(ctx, "list contacts by identity", `
		SELECT `+qualifiedContactColumns+`
		FROM `+r.companyContactsTable+` AS contact
		WHERE ($1 <> '' AND LOWER(contact.work_email) = LOWER($1))
			OR ($2 <> '' AND LOWER(contact.linkedin_url) = LOWER($2))
		ORDER BY contact.created_at, contact.id
	`, email, linkedinURL)
}

// Delete removes the contact; their job links, messages and events go with it.
func (r *CompanyContactRepo) Delete(ctx context.Context, contactID string) (bool, error) {
	tag, err := r.postgres.Executor(ctx).
		Exec(ctx, `DELETE FROM `+r.companyContactsTable+` WHERE id = $1`, contactID)
	if err != nil {
		return false, fmt.Errorf("delete contact: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

// AnonymizeStale blanks the personal details of contacts not refreshed, contacted or heard from
// since the cutoff, and the text of the messages sent to them, keeping the rows so outreach history
// and metrics stay intact.
func (r *CompanyContactRepo) AnonymizeStale(
	ctx context.Context,
	cutoff, now time.Time,
) (int, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		WITH stale AS (
			SELECT contact.id FROM `+r.companyContactsTable+` contact
			WHERE contact.anonymized_at IS NULL
				AND contact.updated_at < $1
				AND (contact.last_interaction_at IS NULL OR contact.last_interaction_at < $1)
				AND NOT EXISTS (
					SELECT 1 FROM `+r.outreachMessagesTable+` message
					WHERE message.contact_id = contact.id AND message.created_at >= $1
				)
		),
		redacted_messages AS (
			UPDATE `+r.outreachMessagesTable+` SET subject = NULL, body_content = '', linkedin_note = NULL
			WHERE contact_id IN (SELECT id FROM stale)
		)
		UPDATE `+r.companyContactsTable+` SET
			full_name = $3, work_email = NULL, phone_number = NULL, linkedin_url = NULL,
			provider_person_id = NULL, is_email_verified = FALSE, email_verification_status = 'UNVERIFIED',
			enrichment_metadata = '{}'::jsonb, anonymized_at = $2, updated_at = $2
		WHERE id IN (SELECT id FROM stale)
	`, cutoff, now, config.GetAnonymizedContactName())
	if err != nil {
		return 0, fmt.Errorf("anonymize stale contacts: %w", err)
	}

	return int(tag.RowsAffected()), nil
}

// EmailInUse reports whether another contact at the company already has the address.
func (r *CompanyContactRepo) EmailInUse(
	ctx context.Context,
	companyDomain, address, exceptID string,
) (bool, error) {
	var inUse bool

	err := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM `+r.companyContactsTable+`
			WHERE company_domain = $1 AND LOWER(work_email) = LOWER($2) AND id <> $3)
	`, companyDomain, address, exceptID).Scan(&inUse)
	if err != nil {
		return false, fmt.Errorf("check contact email: %w", err)
	}

	return inUse, nil
}
