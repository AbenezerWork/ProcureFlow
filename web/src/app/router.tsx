import type { ReactNode } from "react";
import { Navigate, Outlet, createBrowserRouter } from "react-router-dom";

import { AppProviders } from "@/app/providers";
import { useAuth } from "@/features/auth/auth-context";
import { AppShell } from "@/shared/components/layout/AppShell";
import { LoginPage } from "@/routes/auth/LoginPage";
import { RegisterPage } from "@/routes/auth/RegisterPage";
import { DashboardPage } from "@/routes/dashboard/DashboardPage";
import { OrganizationsPage } from "@/routes/organizations/OrganizationsPage";
import { MembersPage } from "@/routes/members/MembersPage";
import { VendorsPage } from "@/routes/vendors/VendorsPage";
import { NewVendorPage } from "@/routes/vendors/NewVendorPage";
import { VendorDetailPage } from "@/routes/vendors/VendorDetailPage";
import { RequestsPage } from "@/routes/requests/RequestsPage";
import { NewRequestPage } from "@/routes/requests/NewRequestPage";
import { RequestDetailPage } from "@/routes/requests/RequestDetailPage";
import { ApprovalsPage } from "@/routes/approvals/ApprovalsPage";
import { RFQsPage } from "@/routes/rfqs/RFQsPage";
import { NewRFQPage } from "@/routes/rfqs/NewRFQPage";
import { RFQDetailPage } from "@/routes/rfqs/RFQDetailPage";
import { RFQComparisonPage } from "@/routes/rfqs/RFQComparisonPage";
import { QuotationsPage } from "@/routes/quotations/QuotationsPage";
import { QuotationDetailPage } from "@/routes/quotations/QuotationDetailPage";
import { AwardsPage } from "@/routes/awards/AwardsPage";
import { AwardDetailPage } from "@/routes/awards/AwardDetailPage";
import { ActivityPage } from "@/routes/activity/ActivityPage";
import { SettingsPage } from "@/routes/settings/SettingsPage";

function RootLayout() {
  return (
    <AppProviders>
      <Outlet />
    </AppProviders>
  );
}

function RequireAuth() {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <AppShell /> : <Navigate to="/auth/login" replace />;
}

function PublicOnly({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <Navigate to="/app/dashboard" replace /> : children;
}

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { index: true, element: <Navigate to="/app/dashboard" replace /> },
      {
        path: "auth/login",
        element: (
          <PublicOnly>
            <LoginPage />
          </PublicOnly>
        ),
      },
      {
        path: "auth/register",
        element: (
          <PublicOnly>
            <RegisterPage />
          </PublicOnly>
        ),
      },
      {
        path: "app",
        element: <RequireAuth />,
        children: [
          { index: true, element: <Navigate to="/app/dashboard" replace /> },
          { path: "dashboard", element: <DashboardPage /> },
          { path: "organizations", element: <OrganizationsPage /> },
          { path: "members", element: <MembersPage /> },
          { path: "vendors", element: <VendorsPage /> },
          { path: "vendors/new", element: <NewVendorPage /> },
          { path: "vendors/:vendorId", element: <VendorDetailPage /> },
          { path: "requests", element: <RequestsPage /> },
          { path: "requests/new", element: <NewRequestPage /> },
          { path: "requests/:requestId", element: <RequestDetailPage /> },
          { path: "approvals", element: <ApprovalsPage /> },
          { path: "rfqs", element: <RFQsPage /> },
          { path: "rfqs/new", element: <NewRFQPage /> },
          { path: "rfqs/:rfqId", element: <RFQDetailPage /> },
          { path: "rfqs/:rfqId/comparison", element: <RFQComparisonPage /> },
          { path: "rfqs/:rfqId/quotations/:quotationId", element: <QuotationDetailPage /> },
          { path: "rfqs/:rfqId/award", element: <AwardDetailPage /> },
          { path: "quotations", element: <QuotationsPage /> },
          { path: "awards", element: <AwardsPage /> },
          { path: "activity", element: <ActivityPage /> },
          { path: "settings", element: <SettingsPage /> },
        ],
      },
      { path: "*", element: <Navigate to="/app/dashboard" replace /> },
    ],
  },
]);
