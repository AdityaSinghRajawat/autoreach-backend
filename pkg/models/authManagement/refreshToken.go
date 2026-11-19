package authManagement

import (
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type RefreshToken struct {
	ID        string     `db:"id"         validate:"required,uuid"`
	UserID    string     `db:"user_id"    validate:"required,uuid"`
	FamilyID  string     `db:"family_id"  validate:"required,uuid"`
	TokenHash string     `db:"token_hash" validate:"required,len=64,hexadecimal"`
	ExpiresAt time.Time  `db:"expires_at" validate:"required,gtfield=CreatedAt"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at" validate:"required"`
}

func NewRefreshToken(lifetime time.Duration) *RefreshToken {
	now := utils.GetCurrTimeStamp()

	return &RefreshToken{
		ID:        utils.GenerateUUID(),
		FamilyID:  utils.GenerateUUID(),
		ExpiresAt: now.Add(lifetime),
		CreatedAt: now,
	}
}

func (r *RefreshToken) Validate() error {
	if err := validations.Validate(r); err != nil {
		return fmt.Errorf("refresh token validation failed: %w", err)
	}

	return nil
}

func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

func (r *RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(r.ExpiresAt)
}
