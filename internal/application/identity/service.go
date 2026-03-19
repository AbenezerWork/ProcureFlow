package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FullName     string
}

type UserRepository interface {
	CreateUser(ctx context.Context, params CreateUserParams) (domainidentity.User, error)
	GetUserByEmail(ctx context.Context, email string) (domainidentity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domainidentity.User, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenManager interface {
	Issue(claims domainidentity.Claims) (domainidentity.Token, error)
	Verify(token string) (domainidentity.Claims, error)
}

type RegisterInput struct {
	Email    string
	Password string
	FullName string
}

type LoginInput struct {
	Email    string
	Password string
}

type Service struct {
	repo     UserRepository
	password PasswordHasher
	tokens   TokenManager
}

func NewService(repo UserRepository, password PasswordHasher, tokens TokenManager) Service {
	return Service{
		repo:     repo,
		password: password,
		tokens:   tokens,
	}
}

func (s Service) Register(ctx context.Context, input RegisterInput) (domainidentity.Session, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	fullName := strings.TrimSpace(input.FullName)
	password := strings.TrimSpace(input.Password)
	if email == "" || fullName == "" || password == "" {
		return domainidentity.Session{}, ErrInvalidCredentials
	}

	passwordHash, err := s.password.Hash(password)
	if err != nil {
		return domainidentity.Session{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
	})
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return domainidentity.Session{}, err
		}

		return domainidentity.Session{}, fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokens.Issue(domainidentity.Claims{UserID: user.ID})
	if err != nil {
		return domainidentity.Session{}, fmt.Errorf("issue token: %w", err)
	}

	return domainidentity.Session{User: user, Token: token}, nil
}

func (s Service) Login(ctx context.Context, input LoginInput) (domainidentity.Session, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return domainidentity.Session{}, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domainidentity.Session{}, ErrInvalidCredentials
		}

		return domainidentity.Session{}, fmt.Errorf("load user by email: %w", err)
	}

	if !user.IsActive {
		return domainidentity.Session{}, ErrUserInactive
	}

	if err := s.password.Compare(user.PasswordHash, password); err != nil {
		return domainidentity.Session{}, ErrInvalidCredentials
	}

	if err := s.repo.UpdateUserLastLogin(ctx, user.ID); err != nil {
		return domainidentity.Session{}, fmt.Errorf("update user last login: %w", err)
	}

	user, err = s.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domainidentity.Session{}, ErrInvalidCredentials
		}

		return domainidentity.Session{}, fmt.Errorf("reload user after login: %w", err)
	}

	token, err := s.tokens.Issue(domainidentity.Claims{UserID: user.ID})
	if err != nil {
		return domainidentity.Session{}, fmt.Errorf("issue token: %w", err)
	}

	return domainidentity.Session{User: user, Token: token}, nil
}

func (s Service) CurrentUser(ctx context.Context, userID uuid.UUID) (domainidentity.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domainidentity.User{}, err
		}

		return domainidentity.User{}, fmt.Errorf("load current user: %w", err)
	}

	if !user.IsActive {
		return domainidentity.User{}, ErrUserInactive
	}

	return user, nil
}

func (s Service) VerifyToken(token string) (domainidentity.Claims, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return domainidentity.Claims{}, ErrInvalidToken
	}

	return claims, nil
}
