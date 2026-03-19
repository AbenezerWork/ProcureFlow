package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/google/uuid"
)

type fakeUserRepository struct {
	createUserFn          func(context.Context, CreateUserParams) (domainidentity.User, error)
	getUserByEmailFn      func(context.Context, string) (domainidentity.User, error)
	getUserByIDFn         func(context.Context, uuid.UUID) (domainidentity.User, error)
	updateUserLastLoginFn func(context.Context, uuid.UUID) error
}

func (f fakeUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (domainidentity.User, error) {
	return f.createUserFn(ctx, params)
}

func (f fakeUserRepository) GetUserByEmail(ctx context.Context, email string) (domainidentity.User, error) {
	return f.getUserByEmailFn(ctx, email)
}

func (f fakeUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domainidentity.User, error) {
	return f.getUserByIDFn(ctx, id)
}

func (f fakeUserRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	return f.updateUserLastLoginFn(ctx, id)
}

type fakePasswordHasher struct {
	hashFn    func(string) (string, error)
	compareFn func(string, string) error
}

func (f fakePasswordHasher) Hash(password string) (string, error) {
	return f.hashFn(password)
}

func (f fakePasswordHasher) Compare(hash, password string) error {
	return f.compareFn(hash, password)
}

type fakeTokenManager struct {
	issueFn  func(domainidentity.Claims) (domainidentity.Token, error)
	verifyFn func(string) (domainidentity.Claims, error)
}

func (f fakeTokenManager) Issue(claims domainidentity.Claims) (domainidentity.Token, error) {
	return f.issueFn(claims)
}

func (f fakeTokenManager) Verify(token string) (domainidentity.Claims, error) {
	return f.verifyFn(token)
}

func TestServiceRegister(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := NewService(
		fakeUserRepository{
			createUserFn: func(_ context.Context, params CreateUserParams) (domainidentity.User, error) {
				if params.Email != "user@example.com" {
					t.Fatalf("unexpected email: %s", params.Email)
				}
				if params.PasswordHash != "hashed-password" {
					t.Fatalf("unexpected password hash: %s", params.PasswordHash)
				}

				now := time.Now().UTC()
				return domainidentity.User{
					ID:           userID,
					Email:        params.Email,
					PasswordHash: params.PasswordHash,
					FullName:     params.FullName,
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
		},
		fakePasswordHasher{
			hashFn: func(password string) (string, error) {
				if password != "password123" {
					t.Fatalf("unexpected password: %s", password)
				}
				return "hashed-password", nil
			},
			compareFn: func(string, string) error { return nil },
		},
		fakeTokenManager{
			issueFn: func(claims domainidentity.Claims) (domainidentity.Token, error) {
				if claims.UserID != userID {
					t.Fatalf("unexpected user ID in claims: %s", claims.UserID)
				}
				return domainidentity.Token{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
			},
			verifyFn: func(string) (domainidentity.Claims, error) { return domainidentity.Claims{}, nil },
		},
	)

	session, err := service.Register(context.Background(), RegisterInput{
		Email:    "User@Example.com",
		Password: "password123",
		FullName: "Ada Lovelace",
	})
	if err != nil {
		t.Fatalf("register returned error: %v", err)
	}

	if session.User.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %s", session.User.Email)
	}

	if session.Token.AccessToken != "token" {
		t.Fatalf("expected token to be issued")
	}
}

func TestServiceLoginInvalidPassword(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Now().UTC()
	service := NewService(
		fakeUserRepository{
			getUserByEmailFn: func(_ context.Context, _ string) (domainidentity.User, error) {
				return domainidentity.User{
					ID:           userID,
					Email:        "user@example.com",
					PasswordHash: "stored-hash",
					FullName:     "Ada Lovelace",
					IsActive:     true,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			getUserByIDFn: func(_ context.Context, _ uuid.UUID) (domainidentity.User, error) {
				return domainidentity.User{}, errors.New("unexpected call")
			},
			updateUserLastLoginFn: func(_ context.Context, _ uuid.UUID) error {
				return errors.New("unexpected call")
			},
		},
		fakePasswordHasher{
			hashFn: func(password string) (string, error) { return password, nil },
			compareFn: func(hash, password string) error {
				if hash != "stored-hash" {
					t.Fatalf("unexpected hash: %s", hash)
				}
				if password != "wrong-password" {
					t.Fatalf("unexpected password: %s", password)
				}
				return errors.New("mismatch")
			},
		},
		fakeTokenManager{
			issueFn: func(domainidentity.Claims) (domainidentity.Token, error) {
				return domainidentity.Token{}, errors.New("unexpected call")
			},
			verifyFn: func(string) (domainidentity.Claims, error) { return domainidentity.Claims{}, nil },
		},
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}
