package authManagement

import (
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type APIClient struct {
	ID         string     `db:"id"           json:"clientId"   validate:"required,uuid"`
	Name       string     `db:"name"         json:"name"       validate:"required,max=100"`
	KeyPrefix  string     `db:"key_prefix"   json:"keyPrefix"  validate:"required,max=32"`
	KeyHash    string     `db:"key_hash"     json:"-"          validate:"required,len=64,hexadecimal"`
	Scopes     []string   `db:"scopes"       json:"scopes"     validate:"required,min=1,unique,dive,validate_api_client_grant"`
	LastUsedAt *time.Time `db:"last_used_at" json:"lastUsedAt"`
	RevokedAt  *time.Time `db:"revoked_at"   json:"revokedAt"`
	CreatedAt  time.Time  `db:"created_at"   json:"createdAt"  validate:"required"`
	UpdatedAt  time.Time  `db:"updated_at"   json:"-"          validate:"required"`
}

func NewAPIClient() *APIClient {
	now := utils.GetCurrTimeStamp()

	return &APIClient{
		ID:        utils.GenerateUUID(),
		Scopes:    []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *APIClient) UpdateTimestamp() {
	c.UpdatedAt = utils.GetCurrTimeStamp()
}

func (c *APIClient) Validate() error {
	if err := validations.Validate(c); err != nil {
		return fmt.Errorf("api client validation failed: %w", err)
	}

	return nil
}
