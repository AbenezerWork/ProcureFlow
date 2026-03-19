package middleware

import (
	"context"
	"net/http"
	"strings"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	"github.com/google/uuid"
)

type authContextKey struct{}

type TokenVerifier interface {
	VerifyToken(token string) (domainidentity.Claims, error)
}

func RequireAuthentication(tokens TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			claims, err := tokens.VerifyToken(token)
			if err != nil {
				http.Error(w, "invalid bearer token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), authContextKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthenticatedUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(authContextKey{}).(uuid.UUID)
	return userID, ok
}

func bearerToken(header string) string {
	value := strings.TrimSpace(header)
	if value == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}
