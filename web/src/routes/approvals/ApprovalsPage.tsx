import { useMemo } from "react";
import { Link } from "react-router-dom";

import { useOrganization } from "@/features/organizations/organization-context";
import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import { canAccessApprovals } from "@/shared/lib/permissions";
import type { ProcurementRequest } from "@/shared/types/api";

export function ApprovalsPage() {
  const { activeOrganization } = useOrganization();
  const canApprove = canAccessApprovals(activeOrganization?.role);
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listApprovalInbox(token, organizationId)).procurement_requests,
    [],
  );
  const { data, isLoading, error } = useTenantResource<ProcurementRequest[]>(loader, canApprove);
  const approvals = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Approval control"
        title="Approvals"
        description="Review submitted procurement requests that are waiting for a decision."
      />

      {!canApprove ? (
        <Notice title="Approval access is restricted">
          Your current organization role cannot review or decide procurement approvals.
        </Notice>
      ) : null}
      {canApprove && error ? <Notice title="Unable to load approval inbox" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>Inbox</h2>
          <span>{isLoading ? "Loading" : `${approvals.length} pending`}</span>
        </div>
        <DataTable
          rows={approvals}
          getRowKey={(row) => row.id}
          emptyLabel="No submitted requests are waiting for approval."
          columns={[
            { key: "title", header: "Request", render: (row) => <Link to={`/app/requests/${row.id}`}>{row.title}</Link> },
            {
              key: "amount",
              header: "Estimate",
              render: (row) => formatAmount(row.estimated_total_amount, row.currency_code),
            },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
      </section>
    </>
  );
}
