import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import type { ProcurementRequest } from "@/shared/types/api";

export function ApprovalsPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listApprovalInbox(token, organizationId)).procurement_requests,
    [],
  );
  const { data, isLoading, error } = useTenantResource<ProcurementRequest[]>(loader);
  const approvals = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Approval control"
        title="Approvals"
        description="Review submitted procurement requests that are waiting for a decision."
      />

      {error ? <Notice title="Unable to load approval inbox" tone="danger">{error}</Notice> : null}

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
      </section>
    </>
  );
}
