import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import type { RFQ } from "@/shared/types/api";

export function RFQsPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listRFQs(token, organizationId)).rfqs,
    [],
  );
  const { data, isLoading, error } = useTenantResource<RFQ[]>(loader);
  const rfqs = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Sourcing"
        title="RFQs"
        description="Create RFQs from approved requests, attach vendors, publish, compare, and award."
      />

      {error ? <Notice title="Unable to load RFQs" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>RFQ pipeline</h2>
          <span>{isLoading ? "Loading" : `${rfqs.length} events`}</span>
        </div>
        <DataTable
          rows={rfqs}
          getRowKey={(row) => row.id}
          emptyLabel="No RFQs have been created for this organization."
          columns={[
            { key: "title", header: "Title", render: (row) => row.title },
            { key: "reference", header: "Reference", render: (row) => row.reference_number ?? "Not set" },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
      </section>
    </>
  );
}
