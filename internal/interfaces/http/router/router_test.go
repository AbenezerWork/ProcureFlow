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

func (stubOrganizationHandler) Get(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubOrganizationHandler) Update(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubOrganizationHandler) ListMemberships(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubOrganizationHandler) CreateMembership(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubOrganizationHandler) UpdateMembership(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubOrganizationHandler) TransferOwnership(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type stubVendorHandler struct{}

func (stubVendorHandler) Create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubVendorHandler) List(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubVendorHandler) Get(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubVendorHandler) Update(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubVendorHandler) Archive(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type stubProcurementHandler struct{}

func (stubProcurementHandler) CreateRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubProcurementHandler) ListRequests(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) ListApprovalInbox(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) GetRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) UpdateRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) SubmitRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) ApproveRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) RejectRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) CancelRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) ListItems(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) CreateItem(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubProcurementHandler) UpdateItem(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubProcurementHandler) DeleteItem(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type stubRFQHandler struct{}

func (stubRFQHandler) Create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubRFQHandler) List(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Get(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Update(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Publish(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Close(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Evaluate(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) Cancel(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) ListItems(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) ListVendors(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubRFQHandler) AttachVendor(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubRFQHandler) RemoveVendor(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

type stubQuotationHandler struct{}

func (stubQuotationHandler) Compare(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) Create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubQuotationHandler) List(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) Get(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) Update(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) Submit(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) Reject(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) ListItems(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (stubQuotationHandler) UpdateItem(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type stubAwardHandler struct{}

func (stubAwardHandler) Create(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

func (stubAwardHandler) Get(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type stubActivityLogHandler struct{}

func (stubActivityLogHandler) ListByEntity(w http.ResponseWriter, _ *http.Request) {
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
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, func(next http.Handler) http.Handler {
		return next
	}, "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestNewRoutesOpenAPISpec(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, func(next http.Handler) http.Handler {
		return next
	}, "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if got, want := recorder.Header().Get("Content-Type"), "application/yaml; charset=utf-8"; got != want {
		t.Fatalf("expected content type %q, got %q", want, got)
	}
}

func TestNewRoutesSwaggerUI(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, func(next http.Handler) http.Handler {
		return next
	}, "X-Tenant-ID")

	request := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if got, want := recorder.Header().Get("Content-Type"), "text/html; charset=utf-8"; got != want {
		t.Fatalf("expected content type %q, got %q", want, got)
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
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, func(next http.Handler) http.Handler {
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
	}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, httpmiddleware.RequireAuthentication(fakeTokenVerifier{
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

func TestNewRoutesQuotationComparison(t *testing.T) {
	t.Parallel()

	router := New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, httpmiddleware.RequireAuthentication(fakeTokenVerifier{
		verifyFn: func(token string) (domainidentity.Claims, error) {
			if token != "token" {
				t.Fatalf("unexpected token: %s", token)
			}

			return domainidentity.Claims{UserID: uuid.New()}, nil
		},
	}), "X-Tenant-ID")

	organizationID := uuid.New()
	rfqID := uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+organizationID.String()+"/rfqs/"+rfqID.String()+"/comparison", nil)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("X-Tenant-ID", organizationID.String())
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
	}), stubAuthHandler{}, stubOrganizationHandler{}, stubVendorHandler{}, stubProcurementHandler{}, stubQuotationHandler{}, stubAwardHandler{}, stubActivityLogHandler{}, stubRFQHandler{}, httpmiddleware.RequireAuthentication(fakeTokenVerifier{
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
