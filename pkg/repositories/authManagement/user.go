package authManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
)

const userColumns = `id, email, password_hash, roles, is_active, email_verified_at, created_at,
	updated_at`

type UserRepo struct {
	postgres   *db.PostgresConn
	usersTable string
}

func NewUserRepo(postgres *db.PostgresConn) (*UserRepo, error) {
	if postgres == nil {
		return nil, fmt.Errorf("new user repo: %w", config.ErrMissingDependency)
	}

	return &UserRepo{
		postgres:   postgres,
		usersTable: config.GetTableName(config.TableUsers),
	}, nil
}

func (r *UserRepo) Upsert(ctx context.Context, user *authModel.User) (*authModel.User, error) {
	_, err := r.postgres.Executor(ctx).Exec(ctx, `
		INSERT INTO `+r.usersTable+` (`+userColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			roles = EXCLUDED.roles,
			is_active = EXCLUDED.is_active,
			email_verified_at = EXCLUDED.email_verified_at,
			updated_at = EXCLUDED.updated_at
	`, user.ID, user.Email, user.PasswordHash, user.Roles, user.IsActive, user.EmailVerifiedAt,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if db.IsUniqueViolation(err) {
			return nil, fmt.Errorf("upsert user: %w", config.ErrDuplicateRecord)
		}

		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*authModel.User, error) {
	return r.findOne(ctx, `LOWER(email) = LOWER($1)`, email, "find user by email")
}

func (r *UserRepo) FindByID(ctx context.Context, userID string) (*authModel.User, error) {
	return r.findOne(ctx, `id = $1`, userID, "find user by id")
}

func (r *UserRepo) findOne(
	ctx context.Context,
	condition string,
	argument any,
	operation string,
) (*authModel.User, error) {
	row := r.postgres.Executor(ctx).QueryRow(ctx, `
		SELECT `+userColumns+` FROM `+r.usersTable+` WHERE `+condition+` LIMIT 1
	`, argument)

	return db.FindOne(row, scanUser, operation)
}

func scanUser(row db.RowScanner) (*authModel.User, error) {
	user := authModel.User{}

	err := row.Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Roles,
		&user.IsActive, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// MarkEmailVerified keeps the first verification time and reports whether the user exists.
func (r *UserRepo) MarkEmailVerified(
	ctx context.Context,
	userID string,
	now time.Time,
) (bool, error) {
	tag, err := r.postgres.Executor(ctx).Exec(ctx, `
		UPDATE `+r.usersTable+`
		SET email_verified_at = COALESCE(email_verified_at, $2), updated_at = $2
		WHERE id = $1
	`, userID, now)
	if err != nil {
		return false, fmt.Errorf("mark email verified: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
