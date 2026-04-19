import { useMemo, useState } from "react";
import { Link } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import type { ProcurementRequest } from "@/shared/types/api";

export function RequestsPage() {
  const tenant = useTenantParams();
  const [actionError, setActionError] = useState<string | null>(null);
  const [submittingRequestId, setSubmittingRequestId] = useState<string | null>(null);
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listProcurementRequests(token, organizationId)).procurement_requests,
    [],
  );
  const { data, isLoading, error, reload } = useTenantResource<ProcurementRequest[]>(loader);
  const requests = data ?? [];

  async function submitDraft(requestId: string) {
    if (!tenant) {
      return;
    }

    setActionError(null);
    setSubmittingRequestId(requestId);

    try {
      await api.submitProcurementRequest(tenant.token, tenant.tenantId, requestId);
      await reload();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Unable to submit request");
    } finally {
      setSubmittingRequestId(null);
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Demand intake"
        title="Procurement requests"
        description="Draft, submit, review, and track internal purchasing demand."
        action={<Link className="button button-primary link-button" to="/app/requests/new">New request</Link>}
      />

      {error ? <Notice title="Unable to load requests" tone="danger">{error}</Notice> : null}
      {actionError ? <Notice title="Request action failed" tone="danger">{actionError}</Notice> : null}

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
            { key: "title", header: "Title", render: (row) => <Link to={`/app/requests/${row.id}`}>{row.title}</Link> },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            {
              key: "amount",
              header: "Estimate",
              render: (row) => formatAmount(row.estimated_total_amount, row.currency_code),
            },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
            {
              key: "actions",
              header: "Actions",
              render: (row) =>
                row.status === "draft" ? (
                  <div className="row-actions">
                    <Link className="button button-secondary link-button" to={`/app/requests/${row.id}`}>
                      Resume editing
                    </Link>
                    <Button
                      disabled={submittingRequestId === row.id}
                      onClick={() => void submitDraft(row.id)}
                      type="button"
                    >
                      {submittingRequestId === row.id ? "Submitting" : "Submit"}
                    </Button>
                  </div>
                ) : (
                  <Link className="button button-secondary link-button" to={`/app/requests/${row.id}`}>
                    View
                  </Link>
                ),
            },
          ]}
        />
      </section>
    </>
  );
}
