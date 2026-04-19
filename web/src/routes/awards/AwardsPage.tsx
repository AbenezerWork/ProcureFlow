import { useMemo } from "react";
import { Link } from "react-router-dom";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import type { RFQ } from "@/shared/types/api";

export function AwardsPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listRFQs(token, organizationId)).rfqs.filter((rfq) => rfq.status === "awarded"),
    [],
  );
  const { data, isLoading, error } = useTenantResource<RFQ[]>(loader);
  const awardedRFQs = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Award decisions"
        title="Awards"
        description="Award records are created from submitted quotations on eligible RFQs."
      />

      {error ? <Notice title="Unable to load awards" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>Awarded RFQs</h2>
          <span>{isLoading ? "Loading" : `${awardedRFQs.length} awarded`}</span>
        </div>
        <DataTable
          rows={awardedRFQs}
          getRowKey={(row) => row.id}
          emptyLabel="No RFQs have been awarded yet."
          columns={[
            { key: "title", header: "RFQ", render: (row) => <Link to={`/app/rfqs/${row.id}/award`}>{row.title}</Link> },
            { key: "reference", header: "Reference", render: (row) => row.reference_number ?? "Not set" },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
      </section>
    </>
  );
}
