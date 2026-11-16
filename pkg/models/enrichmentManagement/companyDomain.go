package enrichmentManagement

import (
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

// CompanyDomain caches a company name's website domain; a nil Domain records a failed lookup.
type CompanyDomain struct {
	NormalizedName string    `db:"normalized_name" json:"normalizedName" validate:"required,max=255"`
	CompanyName    string    `db:"company_name"    json:"companyName"    validate:"required,max=255"`
	Domain         *string   `db:"company_domain"  json:"companyDomain"  validate:"omitnil,fqdn,max=255"`
	Provider       *string   `db:"provider"        json:"provider"       validate:"omitnil,max=30"`
	CreatedAt      time.Time `db:"created_at"      json:"createdAt"      validate:"required"`
	UpdatedAt      time.Time `db:"updated_at"      json:"updatedAt"      validate:"required"`
}

func NewCompanyDomain() *CompanyDomain {
	now := utils.GetCurrTimeStamp()

	return &CompanyDomain{CreatedAt: now, UpdatedAt: now}
}

func (d *CompanyDomain) UpdateTimestamp() {
	d.UpdatedAt = utils.GetCurrTimeStamp()
}

func (d *CompanyDomain) Validate() error {
	if err := validations.Validate(d); err != nil {
		return fmt.Errorf("company domain validation failed: %w", err)
	}

	return nil
}
