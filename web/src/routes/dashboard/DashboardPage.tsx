import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { DataTable } from "@/shared/components/ui/DataTable";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import type { ProcurementRequest, RFQ, Vendor } from "@/shared/types/api";

type DashboardData = {
  requests: ProcurementRequest[];
  approvals: ProcurementRequest[];
  rfqs: RFQ[];
  vendors: Vendor[];
};

export function DashboardPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string): Promise<DashboardData> => {
      const [requests, approvals, rfqs, vendors] = await Promise.all([
        api.listProcurementRequests(token, organizationId),
        api.listApprovalInbox(token, organizationId),
        api.listRFQs(token, organizationId),
        api.listVendors(token, organizationId),
      ]);

      return {
        requests: requests.procurement_requests,
        approvals: approvals.procurement_requests,
        rfqs: rfqs.rfqs,
        vendors: vendors.vendors,
      };
    },
    [],
  );

  const { data, isLoading, error } = useTenantResource(loader);
  const requests = data?.requests ?? [];
  const rfqs = data?.rfqs ?? [];
  const vendors = data?.vendors ?? [];
  const approvals = data?.approvals ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="Dashboard"
        description="Track demand intake, approval pressure, sourcing progress, and active supplier coverage."
      />

      {error ? <Notice title="Dashboard data unavailable" tone="danger">{error}</Notice> : null}

      <section className="metric-grid" aria-label="Operational summary">
        <Metric label="Draft requests" value={requests.filter((item) => item.status === "draft").length} />
        <Metric label="Pending approvals" value={approvals.length} tone="attention" />
        <Metric label="Published RFQs" value={rfqs.filter((item) => item.status === "published").length} />
        <Metric label="Awarded RFQs" value={rfqs.filter((item) => item.status === "awarded").length} />
        <Metric label="Active vendors" value={vendors.filter((item) => item.status === "active").length} />
      </section>

      <section className="content-grid">
        <div className="panel span-8">
          <div className="panel-heading">
            <h2>Requests needing approval</h2>
            <span>{isLoading ? "Loading" : `${approvals.length} open`}</span>
          </div>
          <DataTable
            rows={approvals.slice(0, 6)}
            getRowKey={(row) => row.id}
            emptyLabel="No submitted requests are awaiting approval."
            columns={[
              { key: "title", header: "Request", render: (row) => row.title },
              {
                key: "amount",
                header: "Estimate",
                render: (row) => formatAmount(row.estimated_total_amount, row.currency_code),
              },
              { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
              { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
            ]}
          />
        </div>

        <div className="panel span-4">
          <div className="panel-heading">
            <h2>Recent RFQs</h2>
            <span>{rfqs.length} total</span>
          </div>
          <div className="stack-list">
            {rfqs.slice(0, 5).map((rfq) => (
              <article key={rfq.id} className="stack-list-item">
                <div>
                  <strong>{rfq.title}</strong>
                  <span>{rfq.reference_number ?? "No reference"}</span>
                </div>
                <StatusBadge status={rfq.status} />
              </article>
            ))}
            {rfqs.length === 0 ? <p className="empty-copy">No RFQs have been created yet.</p> : null}
          </div>
        </div>
      </section>
    </>
  );
}

function Metric({ label, value, tone = "default" }: { label: string; value: number; tone?: "default" | "attention" }) {
  return (
    <article className={`metric-card metric-${tone}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  );
}
