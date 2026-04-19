import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { formatAmount } from "@/shared/lib/format";
import type { RFQQuotationComparison } from "@/shared/types/api";

export function RFQComparisonPage() {
  const { rfqId = "" } = useParams();
  const tenant = useTenantParams();
  const [comparison, setComparison] = useState<RFQQuotationComparison | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!tenant || !rfqId) return;
    setError(null);
    try {
      setComparison(await api.compareRFQQuotations(tenant.token, tenant.tenantId, rfqId));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load comparison");
    }
  }, [rfqId, tenant]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <>
      <PageHeader
        eyebrow="Sourcing"
        title={comparison?.rfq.title ?? "RFQ comparison"}
        description="Compare submitted quotation totals and line-item pricing."
      />
      {error ? <Notice title="Unable to load comparison" tone="danger">{error}</Notice> : null}
      <section className="panel">
        <div className="panel-heading">
          <h2>Submitted quotations</h2>
          <span>{comparison?.quotations.length ?? 0} quotes</span>
        </div>
        <DataTable
          rows={comparison?.quotations ?? []}
          getRowKey={(row) => row.quotation_id}
          emptyLabel="No submitted quotations are available for comparison."
          columns={[
            { key: "vendor", header: "Vendor", render: (row) => row.vendor_name },
            { key: "total", header: "Total", render: (row) => formatAmount(row.total_amount, row.currency_code) },
            { key: "lead", header: "Lead time", render: (row) => row.lead_time_days ?? "Not set" },
            { key: "terms", header: "Payment terms", render: (row) => row.payment_terms ?? "Not set" },
          ]}
        />
      </section>
      <section className="panel">
        <div className="panel-heading">
          <h2>Line items</h2>
        </div>
        <div className="stack-list">
          {(comparison?.line_items ?? []).map((item) => (
            <article className="stack-list-item comparison-item" key={item.rfq_item_id}>
              <div>
                <strong>{item.line_number}. {item.item_name}</strong>
                <span>{item.quantity} {item.unit}</span>
              </div>
              <div className="comparison-prices">
                {item.prices.map((price) => (
                  <span key={price.quotation_item_id}>
                    {price.vendor_name}: {price.unit_price} / {price.line_total}
                  </span>
                ))}
              </div>
            </article>
          ))}
        </div>
      </section>
    </>
  );
}
