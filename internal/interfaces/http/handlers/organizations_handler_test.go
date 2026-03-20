package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	domainidentity "github.com/AbenezerWork/ProcureFlow/internal/domain/identity"
	domainorganization "github.com/AbenezerWork/ProcureFlow/internal/domain/organization"
	httpmiddleware "github.com/AbenezerWork/ProcureFlow/internal/interfaces/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeOrganizationService struct {
	getFn func(ctx context.Context, organizationID, currentUser uuid.UUID) (applicationorganization.OrganizationDetails, error)
}

func (f fakeOrganizationService) Create(context.Context, applicationorganization.CreateInput) (applicationorganization.CreatedOrganization, error) {
	panic("unexpected Create call")
}

func (f fakeOrganizationService) ListForUser(context.Context, uuid.UUID) ([]domainorganization.UserOrganization, error) {
	panic("unexpected ListForUser call")
}

func (f fakeOrganizationService) Get(ctx context.Context, organizationID, currentUser uuid.UUID) (applicationorganization.OrganizationDetails, error) {
	if f.getFn == nil {
		panic("unexpected Get call")
	}

	return f.getFn(ctx, organizationID, currentUser)
}

func (f fakeOrganizationService) Update(context.Context, applicationorganization.UpdateInput) (applicationorganization.OrganizationDetails, error) {
	panic("unexpected Update call")
}

func (f fakeOrganizationService) ListMemberships(context.Context, uuid.UUID, uuid.UUID) ([]applicationorganization.OrganizationMember, error) {
	panic("unexpected ListMemberships call")
}

func (f fakeOrganizationService) AddMembership(context.Context, applicationorganization.AddMembershipInput) (applicationorganization.OrganizationMember, error) {
	panic("unexpected AddMembership call")
}

func (f fakeOrganizationService) UpdateMembership(context.Context, applicationorganization.UpdateMembershipInput) (applicationorganization.OrganizationMember, error) {
	panic("unexpected UpdateMembership call")
}

func (f fakeOrganizationService) TransferOwnership(context.Context, applicationorganization.TransferOwnershipInput) (applicationorganization.OwnershipTransferResult, error) {
	panic("unexpected TransferOwnership call")
}

type fakeTokenVerifier struct {
	verifyFn func(string) (domainidentity.Claims, error)
}

func (f fakeTokenVerifier) VerifyToken(token string) (domainidentity.Claims, error) {
	return f.verifyFn(token)
}

func TestOrganizationGetRejectsMissingTenantContext(t *testing.T) {
	t.Parallel()

	handler := NewOrganizationHandler(fakeOrganizationService{
		getFn: func(context.Context, uuid.UUID, uuid.UUID) (applicationorganization.OrganizationDetails, error) {
			t.Fatal("service should not be called without tenant context")
			return applicationorganization.OrganizationDetails{}, nil
		},
	})

	request := authenticatedOrganizationRequestForTest(http.MethodGet, uuid.New(), "")
	recorder := httptest.NewRecorder()

	authorizedHandlerForTest(handler.Get).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if got, want := recorder.Body.String(), "{\"error\":\"missing tenant context\"}\n"; got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func TestOrganizationGetRejectsInvalidTenantContext(t *testing.T) {
	t.Parallel()

	handler := NewOrganizationHandler(fakeOrganizationService{
		getFn: func(context.Context, uuid.UUID, uuid.UUID) (applicationorganization.OrganizationDetails, error) {
			t.Fatal("service should not be called with invalid tenant context")
			return applicationorganization.OrganizationDetails{}, nil
		},
	})

	request := authenticatedOrganizationRequestForTest(http.MethodGet, uuid.New(), "not-a-uuid")
	recorder := httptest.NewRecorder()

	authorizedHandlerForTest(handler.Get).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if got, want := recorder.Body.String(), "{\"error\":\"invalid tenant_id\"}\n"; got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func TestOrganizationGetRejectsMismatchedTenantContext(t *testing.T) {
	t.Parallel()

	handler := NewOrganizationHandler(fakeOrganizationService{
		getFn: func(context.Context, uuid.UUID, uuid.UUID) (applicationorganization.OrganizationDetails, error) {
			t.Fatal("service should not be called when tenant does not match organization")
			return applicationorganization.OrganizationDetails{}, nil
		},
	})

	request := authenticatedOrganizationRequestForTest(http.MethodGet, uuid.New(), uuid.New().String())
	recorder := httptest.NewRecorder()

	authorizedHandlerForTest(handler.Get).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}

	if got, want := recorder.Body.String(), "{\"error\":\"tenant does not match organization\"}\n"; got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func TestOrganizationGetAllowsMatchingTenantContext(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	currentUser := uuid.New()

	handler := NewOrganizationHandler(fakeOrganizationService{
		getFn: func(_ context.Context, gotOrganizationID, gotCurrentUser uuid.UUID) (applicationorganization.OrganizationDetails, error) {
			if gotOrganizationID != organizationID {
				t.Fatalf("expected organization %s, got %s", organizationID, gotOrganizationID)
			}

			if gotCurrentUser != currentUser {
				t.Fatalf("expected current user %s, got %s", currentUser, gotCurrentUser)
			}

			now := time.Now().UTC()
			return applicationorganization.OrganizationDetails{
				Organization: domainorganization.Organization{
					ID:              organizationID,
					Name:            "Acme",
					Slug:            "acme",
					CreatedByUserID: currentUser,
					CreatedAt:       now,
					UpdatedAt:       now,
				},
				Membership: domainorganization.Membership{
					ID:              uuid.New(),
					OrganizationID:  organizationID,
					UserID:          currentUser,
					Role:            domainorganization.MembershipRoleOwner,
					Status:          domainorganization.MembershipStatusActive,
					CreatedByUserID: currentUser,
					InvitedAt:       now,
					ActivatedAt:     &now,
					CreatedAt:       now,
					UpdatedAt:       now,
				},
			}, nil
		},
	})

	request := authenticatedOrganizationRequestForTest(http.MethodGet, organizationID, organizationID.String())
	recorder := httptest.NewRecorder()

	authorizedHandlerForTestWithUser(handler.Get, currentUser).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func authenticatedOrganizationRequestForTest(method string, organizationID uuid.UUID, tenantID string) *http.Request {
	request := httptest.NewRequest(method, "/api/v1/organizations/"+organizationID.String(), nil)
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("organizationID", organizationID.String())
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))

	request.Header.Set("Authorization", "Bearer token")
	if tenantID != "" {
		request.Header.Set("X-Tenant-ID", tenantID)
	}

	return request
}

func authorizedHandlerForTest(next http.HandlerFunc) http.Handler {
	return authorizedHandlerForTestWithUser(next, uuid.New())
}

func authorizedHandlerForTestWithUser(next http.HandlerFunc, userID uuid.UUID) http.Handler {
	handler := httpmiddleware.RequireAuthentication(fakeTokenVerifier{
		verifyFn: func(token string) (domainidentity.Claims, error) {
			if token != "token" {
				return domainidentity.Claims{}, context.Canceled
			}

			return domainidentity.Claims{UserID: userID}, nil
		},
	})(next)

	return httpmiddleware.WithTenant(handler, "X-Tenant-ID")
}
