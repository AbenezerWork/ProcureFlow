//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	applicationactivitylog "github.com/AbenezerWork/ProcureFlow/internal/application/activitylog"
	applicationaward "github.com/AbenezerWork/ProcureFlow/internal/application/award"
	applicationidentity "github.com/AbenezerWork/ProcureFlow/internal/application/identity"
	applicationorganization "github.com/AbenezerWork/ProcureFlow/internal/application/organization"
	applicationprocurement "github.com/AbenezerWork/ProcureFlow/internal/application/procurement"
	applicationquotation "github.com/AbenezerWork/ProcureFlow/internal/application/quotation"
	applicationrfq "github.com/AbenezerWork/ProcureFlow/internal/application/rfq"
	applicationvendor "github.com/AbenezerWork/ProcureFlow/internal/application/vendor"
	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
	domainprocurement "github.com/AbenezerWork/ProcureFlow/internal/domain/procurement"
	domainquotation "github.com/AbenezerWork/ProcureFlow/internal/domain/quotation"
	domainrfq "github.com/AbenezerWork/ProcureFlow/internal/domain/rfq"
	authinfra "github.com/AbenezerWork/ProcureFlow/internal/infrastructure/auth"
	"github.com/AbenezerWork/ProcureFlow/internal/infrastructure/config"
	dbrepositories "github.com/AbenezerWork/ProcureFlow/internal/infrastructure/database/repositories"
	"github.com/google/uuid"
)

func TestProcurementRequestToAwardFlow(t *testing.T) {
	store, cleanup := openTestStore(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	identityRepository := dbrepositories.NewIdentityRepository(store)
	organizationRepository := dbrepositories.NewOrganizationRepository(store)
	vendorRepository := dbrepositories.NewVendorRepository(store)
	procurementRepository := dbrepositories.NewProcurementRepository(store)
	rfqRepository := dbrepositories.NewRFQRepository(store)
	quotationRepository := dbrepositories.NewQuotationRepository(store)
	awardRepository := dbrepositories.NewAwardRepository(store)
	activityLogRepository := dbrepositories.NewActivityLogRepository(store)

	identityService := applicationidentity.NewService(
		identityRepository,
		authinfra.NewPasswordHasher(),
		authinfra.NewTokenManager(config.AuthConfig{
			JWTIssuer:      "procureflow-integration-test",
			JWTSecret:      "integration-test-secret-with-enough-length",
			AccessTokenTTL: time.Hour,
		}),
	)
	organizationService := applicationorganization.NewService(organizationRepository, organizationRepository, identityRepository)
	vendorService := applicationvendor.NewService(vendorRepository, vendorRepository)
	procurementService := applicationprocurement.NewService(procurementRepository, procurementRepository)
	rfqService := applicationrfq.NewService(rfqRepository, rfqRepository)
	quotationService := applicationquotation.NewService(quotationRepository, quotationRepository)
	awardService := applicationaward.NewService(awardRepository, awardRepository)
	activityLogService := applicationactivitylog.NewService(activityLogRepository)

	unique := uuid.NewString()
	session, err := identityService.Register(ctx, applicationidentity.RegisterInput{
		Email:    "owner-" + unique + "@example.test",
		Password: "secret123",
		FullName: "Integration Owner",
	})
	if err != nil {
		t.Fatalf("register owner: %v", err)
	}
	userID := session.User.ID

	createdOrg, err := organizationService.Create(ctx, applicationorganization.CreateInput{
		Name:        "Integration Org " + unique,
		Slug:        "integration-" + unique,
		CurrentUser: userID,
	})
	if err != nil {
		t.Fatalf("create organization: %v", err)
	}
	orgID := createdOrg.Organization.ID

	createdVendor, err := vendorService.Create(ctx, applicationvendor.CreateInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		Name:           "Blue Nile Supplies " + unique,
		Email:          stringPtr("sales-" + unique + "@vendor.test"),
		Country:        stringPtr("ET"),
	})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}

	createdRequest, err := procurementService.CreateRequest(ctx, applicationprocurement.CreateRequestInput{
		OrganizationID:       orgID,
		CurrentUser:          userID,
		Title:                "Office Chairs",
		Justification:        stringPtr("Team expansion"),
		CurrencyCode:         stringPtr("ETB"),
		EstimatedTotalAmount: stringPtr("10000.00"),
	})
	if err != nil {
		t.Fatalf("create procurement request: %v", err)
	}

	createdItem, err := procurementService.CreateItem(ctx, applicationprocurement.CreateItemInput{
		OrganizationID:     orgID,
		RequestID:          createdRequest.ID,
		CurrentUser:        userID,
		ItemName:           "Ergonomic chair",
		Quantity:           "10",
		Unit:               "pcs",
		EstimatedUnitPrice: stringPtr("950.00"),
	})
	if err != nil {
		t.Fatalf("create procurement item: %v", err)
	}
	if createdItem.LineNumber != 1 {
		t.Fatalf("expected first request item line number 1, got %d", createdItem.LineNumber)
	}

	submittedRequest, err := procurementService.SubmitRequest(ctx, applicationprocurement.SubmitRequestInput{
		OrganizationID: orgID,
		RequestID:      createdRequest.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("submit procurement request: %v", err)
	}
	if submittedRequest.Status != domainprocurement.RequestStatusSubmitted {
		t.Fatalf("expected submitted request, got %s", submittedRequest.Status)
	}

	approvedRequest, err := procurementService.ApproveRequest(ctx, applicationprocurement.DecisionInput{
		OrganizationID:  orgID,
		RequestID:       createdRequest.ID,
		CurrentUser:     userID,
		DecisionComment: stringPtr("Approved for sourcing"),
	})
	if err != nil {
		t.Fatalf("approve procurement request: %v", err)
	}
	if approvedRequest.Status != domainprocurement.RequestStatusApproved {
		t.Fatalf("expected approved request, got %s", approvedRequest.Status)
	}

	createdRFQ, err := rfqService.Create(ctx, applicationrfq.CreateInput{
		OrganizationID:       orgID,
		ProcurementRequestID: approvedRequest.ID,
		CurrentUser:          userID,
		ReferenceNumber:      stringPtr("RFQ-" + unique),
	})
	if err != nil {
		t.Fatalf("create rfq: %v", err)
	}
	rfqItems, err := rfqService.ListItems(ctx, orgID, createdRFQ.ID, userID)
	if err != nil {
		t.Fatalf("list rfq items: %v", err)
	}
	if len(rfqItems) != 1 || rfqItems[0].SourceRequestItemID == nil || *rfqItems[0].SourceRequestItemID != createdItem.ID {
		t.Fatalf("expected rfq item snapshot from procurement item, got %#v", rfqItems)
	}

	attachedVendor, err := rfqService.AttachVendor(ctx, applicationrfq.AttachVendorInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		VendorID:       createdVendor.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("attach rfq vendor: %v", err)
	}

	publishedRFQ, err := rfqService.Publish(ctx, applicationrfq.TransitionInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("publish rfq: %v", err)
	}
	if publishedRFQ.Status != domainrfq.StatusPublished {
		t.Fatalf("expected published rfq, got %s", publishedRFQ.Status)
	}

	createdQuotation, err := quotationService.Create(ctx, applicationquotation.CreateInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		RFQVendorID:    attachedVendor.ID,
		CurrentUser:    userID,
		CurrencyCode:   stringPtr("ETB"),
		LeadTimeDays:   int32Ptr(7),
		PaymentTerms:   stringPtr("Net 30"),
	})
	if err != nil {
		t.Fatalf("create quotation: %v", err)
	}

	quotationItems, err := quotationService.ListItems(ctx, orgID, createdRFQ.ID, createdQuotation.ID, userID)
	if err != nil {
		t.Fatalf("list quotation items: %v", err)
	}
	if len(quotationItems) != 1 || quotationItems[0].RFQItemID != rfqItems[0].ID || quotationItems[0].UnitPrice != "0" {
		t.Fatalf("expected quotation item snapshot with zero pricing, got %#v", quotationItems)
	}

	updatedQuotationItem, err := quotationService.UpdateItem(ctx, applicationquotation.UpdateItemInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		QuotationID:    createdQuotation.ID,
		ItemID:         quotationItems[0].ID,
		CurrentUser:    userID,
		UnitPrice:      stringPtr("900.00"),
		DeliveryDays:   int32Ptr(5),
		Notes:          stringPtr("In stock"),
	})
	if err != nil {
		t.Fatalf("update quotation item: %v", err)
	}
	if updatedQuotationItem.UnitPrice != "900.00" {
		t.Fatalf("expected updated unit price, got %s", updatedQuotationItem.UnitPrice)
	}

	submittedQuotation, err := quotationService.Submit(ctx, applicationquotation.TransitionInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		QuotationID:    createdQuotation.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("submit quotation: %v", err)
	}
	if submittedQuotation.Status != domainquotation.StatusSubmitted {
		t.Fatalf("expected submitted quotation, got %s", submittedQuotation.Status)
	}

	comparison, err := quotationService.Compare(ctx, applicationquotation.ListInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("compare quotations: %v", err)
	}
	if len(comparison.Comparison.Quotations) != 1 {
		t.Fatalf("expected 1 compared quotation, got %d", len(comparison.Comparison.Quotations))
	}
	if comparison.Comparison.Quotations[0].TotalAmount != "9000.00" {
		t.Fatalf("expected quotation total 9000.00, got %s", comparison.Comparison.Quotations[0].TotalAmount)
	}
	if len(comparison.Comparison.LineItems) != 1 || len(comparison.Comparison.LineItems[0].Prices) != 1 {
		t.Fatalf("expected one line item price in comparison, got %#v", comparison.Comparison.LineItems)
	}

	closedRFQ, err := rfqService.Close(ctx, applicationrfq.TransitionInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("close rfq: %v", err)
	}
	if closedRFQ.Status != domainrfq.StatusClosed {
		t.Fatalf("expected closed rfq, got %s", closedRFQ.Status)
	}

	evaluatedRFQ, err := rfqService.Evaluate(ctx, applicationrfq.TransitionInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		CurrentUser:    userID,
	})
	if err != nil {
		t.Fatalf("evaluate rfq: %v", err)
	}
	if evaluatedRFQ.Status != domainrfq.StatusEvaluated {
		t.Fatalf("expected evaluated rfq, got %s", evaluatedRFQ.Status)
	}

	createdAward, err := awardService.Create(ctx, applicationaward.CreateInput{
		OrganizationID: orgID,
		RFQID:          createdRFQ.ID,
		QuotationID:    createdQuotation.ID,
		CurrentUser:    userID,
		Reason:         "Best commercial value",
	})
	if err != nil {
		t.Fatalf("create award: %v", err)
	}
	if createdAward.QuotationID != createdQuotation.ID {
		t.Fatalf("expected award for quotation %s, got %s", createdQuotation.ID, createdAward.QuotationID)
	}

	awardedRFQ, err := rfqService.Get(ctx, orgID, createdRFQ.ID, userID)
	if err != nil {
		t.Fatalf("get awarded rfq: %v", err)
	}
	if awardedRFQ.Status != domainrfq.StatusAwarded {
		t.Fatalf("expected awarded rfq, got %s", awardedRFQ.Status)
	}

	loadedAward, err := awardService.GetByRFQ(ctx, orgID, createdRFQ.ID, userID)
	if err != nil {
		t.Fatalf("get award: %v", err)
	}
	if loadedAward.ID != createdAward.ID {
		t.Fatalf("expected loaded award %s, got %s", createdAward.ID, loadedAward.ID)
	}

	rfqLogs, err := activityLogService.ListByEntity(ctx, applicationactivitylog.ListByEntityInput{
		OrganizationID: orgID,
		CurrentUser:    userID,
		EntityType:     string(domainactivitylog.EntityTypeRFQ),
		EntityID:       createdRFQ.ID,
	})
	if err != nil {
		t.Fatalf("list rfq activity logs: %v", err)
	}
	assertHasAction(t, rfqLogs, domainactivitylog.ActionRFQCreated)
	assertHasAction(t, rfqLogs, domainactivitylog.ActionRFQPublished)
	assertHasAction(t, rfqLogs, domainactivitylog.ActionRFQClosed)
	assertHasAction(t, rfqLogs, domainactivitylog.ActionRFQEvaluated)
	assertHasAction(t, rfqLogs, domainactivitylog.ActionRFQAwarded)
}

func assertHasAction(t *testing.T, entries []domainactivitylog.Entry, action string) {
	t.Helper()

	for _, entry := range entries {
		if entry.Action == action {
			return
		}
	}

	t.Fatalf("expected activity log action %q in %#v", action, entries)
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
