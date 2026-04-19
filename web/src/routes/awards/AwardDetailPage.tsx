import { FormEvent, useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { formString } from "@/shared/lib/form";
import { formatDateTime } from "@/shared/lib/format";
import type { Award, Quotation } from "@/shared/types/api";

export function AwardDetailPage() {
  const { rfqId = "" } = useParams();
  const tenant = useTenantParams();
  const [award, setAward] = useState<Award | null>(null);
  const [quotations, setQuotations] = useState<Quotation[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!tenant || !rfqId) return;
    setError(null);
    try {
      const [awardResult, quotationResult] = await Promise.allSettled([
        api.getAward(tenant.token, tenant.tenantId, rfqId),
        api.listQuotations(tenant.token, tenant.tenantId, rfqId),
      ]);
      if (awardResult.status === "fulfilled") setAward(awardResult.value);
      if (quotationResult.status === "fulfilled") setQuotations(quotationResult.value.quotations);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load award");
    }
  }, [rfqId, tenant]);

  useEffect(() => {
    void load();
  }, [load]);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      setAward(
        await api.createAward(tenant.token, tenant.tenantId, rfqId, {
          quotation_id: formString(formData, "quotation_id") ?? "",
          reason: formString(formData, "reason") ?? "",
        }),
      );
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create award");
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Award decisions"
        title="RFQ award"
        description={award ? `Awarded ${formatDateTime(award.awarded_at)}` : "Create or review the award decision for this RFQ."}
      />
      {error ? <Notice title="Award action failed" tone="danger">{error}</Notice> : null}

      {award ? (
        <section className="panel detail-list">
          <div className="panel-heading">
            <h2>Award record</h2>
          </div>
          <dl>
            <div><dt>Quotation ID</dt><dd>{award.quotation_id}</dd></div>
            <div><dt>Reason</dt><dd>{award.reason}</dd></div>
            <div><dt>Awarded at</dt><dd>{formatDateTime(award.awarded_at)}</dd></div>
          </dl>
        </section>
      ) : (
        <form className="panel form-stack" onSubmit={handleCreate}>
          <div className="panel-heading">
            <h2>Create award</h2>
          </div>
          <label>
            <span>Submitted quotation</span>
            <select name="quotation_id" required>
              <option value="">Select quotation</option>
              {quotations.filter((quotation) => quotation.status === "submitted").map((quotation) => (
                <option key={quotation.id} value={quotation.id}>
                  {quotation.vendor_name ?? quotation.id}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Reason</span>
            <input name="reason" required />
          </label>
          <Button type="submit">Create award</Button>
        </form>
      )}

      <section className="panel">
        <div className="panel-heading">
          <h2>Eligible quotations</h2>
        </div>
        <DataTable
          rows={quotations.filter((quotation) => quotation.status === "submitted")}
          getRowKey={(row) => row.id}
          emptyLabel="No submitted quotations are available."
          columns={[
            { key: "vendor", header: "Vendor", render: (row) => row.vendor_name ?? "Not set" },
            { key: "currency", header: "Currency", render: (row) => row.currency_code },
            { key: "lead", header: "Lead time", render: (row) => row.lead_time_days ?? "Not set" },
          ]}
        />
      </section>
    </>
  );
}
