package auth

import (
	"testing"
	"time"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	"github.com/google/uuid"
)

func TestTokenManagerIssueAndVerify(t *testing.T) {
	t.Parallel()

	manager := NewTokenManager(config.AuthConfig{
		JWTIssuer:      "procureflow-test",
		JWTSecret:      "test-secret",
		AccessTokenTTL: time.Hour,
	})

	userID := uuid.New()
	token, err := manager.Issue(domainidentity.Claims{UserID: userID})
	if err != nil {
		t.Fatalf("issue returned error: %v", err)
	}

	claims, err := manager.Verify(token.AccessToken)
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.ExpiresAt.IsZero() {
		t.Fatal("expected expiry to be set")
	}
}
