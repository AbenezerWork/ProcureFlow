import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import type { ProcurementRequest } from "@/shared/types/api";

export function RequestsPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listProcurementRequests(token, organizationId)).procurement_requests,
    [],
  );
  const { data, isLoading, error } = useTenantResource<ProcurementRequest[]>(loader);
  const requests = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Demand intake"
        title="Procurement requests"
        description="Draft, submit, review, and track internal purchasing demand."
      />

      {error ? <Notice title="Unable to load requests" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>Requests</h2>
          <span>{isLoading ? "Loading" : `${requests.length} total`}</span>
        </div>
        <DataTable
          rows={requests}
          getRowKey={(row) => row.id}
          emptyLabel="No procurement requests are available for this organization."
          columns={[
            { key: "title", header: "Title", render: (row) => row.title },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            {
              key: "amount",
              header: "Estimate",
              render: (row) => formatAmount(row.estimated_total_amount, row.currency_code),
            },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
      </section>
    </>
  );
}
