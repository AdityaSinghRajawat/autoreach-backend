package authManagement

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type User struct {
	ID           string                `db:"id"            json:"userId"    validate:"required,uuid"`
	Email        string                `db:"email"         json:"email"     validate:"required,email,max=255"`
	PasswordHash string                `db:"password_hash" json:"-"         validate:"required"`
	Role         modelHelpers.UserRole `db:"role"          json:"role"      validate:"required,oneof=CANDIDATE ADMIN"`
	IsActive     bool                  `db:"is_active"     json:"isActive"`
	CreatedAt    time.Time             `db:"created_at"    json:"createdAt" validate:"required"`
	UpdatedAt    time.Time             `db:"updated_at"    json:"updatedAt" validate:"required"`
}

func NewUser() *User {
	now := utils.GetCurrTimeStamp()

	return &User{
		ID:        uuid.NewString(),
		Role:      modelHelpers.UserRoleCandidate,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (u *User) UpdateTimestamp() {
	u.UpdatedAt = utils.GetCurrTimeStamp()
}

func (u *User) Validate() error {
	if err := validations.Validate(u); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	return nil
}
