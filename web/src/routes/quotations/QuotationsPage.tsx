import { useEffect, useMemo, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";
import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import type { Quotation, RFQ } from "@/shared/types/api";

export function QuotationsPage() {
  const { token } = useAuth();
  const { activeOrganizationId } = useOrganization();
  const rfqLoader = useMemo(
    () => async (tokenValue: string, organizationId: string) =>
      (await api.listRFQs(tokenValue, organizationId)).rfqs,
    [],
  );
  const { data: rfqs, error: rfqError } = useTenantResource<RFQ[]>(rfqLoader);
  const [selectedRFQId, setSelectedRFQId] = useState("");
  const [quotations, setQuotations] = useState<Quotation[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!selectedRFQId && rfqs?.[0]) {
      setSelectedRFQId(rfqs[0].id);
    }
  }, [rfqs, selectedRFQId]);

  useEffect(() => {
    async function loadQuotations() {
      if (!token || !activeOrganizationId || !selectedRFQId) {
        setQuotations([]);
        return;
      }

      setIsLoading(true);
      setError(null);
      try {
        const response = await api.listQuotations(token, activeOrganizationId, selectedRFQId);
        setQuotations(response.quotations);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unable to load quotations");
      } finally {
        setIsLoading(false);
      }
    }

    void loadQuotations();
  }, [activeOrganizationId, selectedRFQId, token]);

  return (
    <>
      <PageHeader
        eyebrow="Vendor responses"
        title="Quotations"
        description="Quotation lists are scoped to a specific RFQ."
      />

      {rfqError ? <Notice title="Unable to load RFQs" tone="danger">{rfqError}</Notice> : null}
      {error ? <Notice title="Unable to load quotations" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>RFQ quotations</h2>
          <label className="inline-control">
            <span>RFQ</span>
            <select value={selectedRFQId} onChange={(event) => setSelectedRFQId(event.target.value)}>
              {(rfqs ?? []).map((rfq) => (
                <option key={rfq.id} value={rfq.id}>
                  {rfq.title}
                </option>
              ))}
            </select>
          </label>
        </div>
        <DataTable
          rows={quotations}
          getRowKey={(row) => row.id}
          emptyLabel={selectedRFQId ? "No quotations are available for this RFQ." : "Create an RFQ before reviewing quotations."}
          columns={[
            { key: "vendor", header: "Vendor", render: (row) => row.vendor_name ?? "Not set" },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            { key: "currency", header: "Currency", render: (row) => row.currency_code },
            { key: "lead", header: "Lead time", render: (row) => row.lead_time_days ?? "Not set" },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
        {isLoading ? <p className="empty-copy">Loading quotations</p> : null}
      </section>
    </>
  );
}
