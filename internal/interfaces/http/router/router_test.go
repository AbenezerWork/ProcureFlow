package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/google/uuid"
)

type stubAuthHandler struct {
	meFn func(http.ResponseWriter, *http.Request)
}

func (stubAuthHandler) Register(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubAuthHandler) Login(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h stubAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if h.meFn != nil {
		h.meFn(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type stubOrganizationHandler struct{}

func (stubOrganizationHandler) Create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubOrganizationHandler) ListMine(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type fakeTokenVerifier struct {
	verifyFn func(string) (domainidentity.Claims, error)
}

func (f fakeTokenVerifier) VerifyToken(token string) (domainidentity.Claims, error) {
	return f.verifyFn(token)
}

func TestNewRoutesHealthCheck(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, func(next http.Handler) http.Handler {
		return next
	}, "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestNewAppliesTenantMiddleware(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := httpmiddleware.TenantFromContext(r.Context())
		if !ok {
			t.Fatal("expected tenant to be present in context")
		}

		if got, want := tenantID.String(), "tenant-123"; got != want {
			t.Fatalf("expected tenant %q, got %q", want, got)
		}

		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, func(next http.Handler) http.Handler {
		return next
	}, "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Tenant-ID", "tenant-123")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestNewProtectsAuthenticatedRoutes(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{
		meFn: func(w http.ResponseWriter, r *http.Request) {
			if _, ok := httpmiddleware.AuthenticatedUserID(r.Context()); !ok {
				t.Fatal("expected authenticated user")
			}

			w.WriteHeader(http.StatusOK)
		},
	}, stubOrganizationHandler{}, httpmiddleware.RequireAuthentication(fakeTokenVerifier{
		verifyFn: func(token string) (domainidentity.Claims, error) {
			if token != "token" {
				t.Fatalf("unexpected token: %s", token)
			}

			return domainidentity.Claims{UserID: uuid.New()}, nil
		},
	}), "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestNewRejectsMissingBearerToken(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, httpmiddleware.RequireAuthentication(fakeTokenVerifier{
		verifyFn: func(string) (domainidentity.Claims, error) {
			return domainidentity.Claims{UserID: uuid.New()}, nil
		},
	}), "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
