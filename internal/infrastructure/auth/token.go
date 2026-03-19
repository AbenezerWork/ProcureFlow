package auth

import (
	"errors"
	"fmt"
	"time"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	issuer string
	secret []byte
	ttl    time.Duration
}

type tokenClaims struct {
	jwt.RegisteredClaims
}

func NewTokenManager(cfg config.AuthConfig) TokenManager {
	return TokenManager{
		issuer: cfg.JWTIssuer,
		secret: []byte(cfg.JWTSecret),
		ttl:    cfg.AccessTokenTTL,
	}
}

func (m TokenManager) Issue(claims domainidentity.Claims) (domainidentity.Token, error) {
	if claims.UserID == uuid.Nil {
		return domainidentity.Token{}, errors.New("issue token: missing user ID")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.UserID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return domainidentity.Token{}, fmt.Errorf("sign token: %w", err)
	}

	return domainidentity.Token{
		AccessToken: signed,
		ExpiresAt:   expiresAt,
	}, nil
}

func (m TokenManager) Verify(token string) (domainidentity.Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(received *jwt.Token) (any, error) {
		if received.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", received.Method.Alg())
		}

		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return domainidentity.Claims{}, err
	}

	claims, ok := parsed.Claims.(*tokenClaims)
	if !ok || !parsed.Valid {
		return domainidentity.Claims{}, errors.New("invalid token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return domainidentity.Claims{}, fmt.Errorf("parse token subject: %w", err)
	}

	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time.UTC()
	}

	return domainidentity.Claims{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}
