package repositories

import (
	"context"
	"errors"

	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type IdentityRepository struct {
	store *database.Store
}

func NewIdentityRepository(store *database.Store) *IdentityRepository {
	return &IdentityRepository{store: store}
}

func (r *IdentityRepository) CreateUser(ctx context.Context, params applicationidentity.CreateUserParams) (domainidentity.User, error) {
	user, err := r.store.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		FullName:     params.FullName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainidentity.User{}, applicationidentity.ErrEmailAlreadyExists
		}

		return domainidentity.User{}, err
	}

	return mapUser(user), nil
}

func (r *IdentityRepository) GetUserByEmail(ctx context.Context, email string) (domainidentity.User, error) {
	user, err := r.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainidentity.User{}, applicationidentity.ErrUserNotFound
		}

		return domainidentity.User{}, err
	}

	return mapUser(user), nil
}

func (r *IdentityRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domainidentity.User, error) {
	user, err := r.store.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainidentity.User{}, applicationidentity.ErrUserNotFound
		}

		return domainidentity.User{}, err
	}

	return mapUser(user), nil
}

func (r *IdentityRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.store.UpdateUserLastLogin(ctx, id)
}

func mapUser(user sqlc.User) domainidentity.User {
	return domainidentity.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		IsActive:     user.IsActive,
		LastLoginAt:  optionalTime(user.LastLoginAt),
		CreatedAt:    requiredTime(user.CreatedAt),
		UpdatedAt:    requiredTime(user.UpdatedAt),
	}
}
