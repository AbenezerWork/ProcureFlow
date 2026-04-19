import { FormEvent, useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formNumber, formString } from "@/shared/lib/form";
import { formatDateTime } from "@/shared/lib/format";
import type { Quotation, QuotationItem } from "@/shared/types/api";

export function QuotationDetailPage() {
  const { rfqId = "", quotationId = "" } = useParams();
  const tenant = useTenantParams();
  const [quotation, setQuotation] = useState<Quotation | null>(null);
  const [items, setItems] = useState<QuotationItem[]>([]);
  const [rejectionReason, setRejectionReason] = useState("");
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!tenant || !rfqId || !quotationId) return;
    setError(null);
    try {
      const [nextQuotation, nextItems] = await Promise.all([
        api.getQuotation(tenant.token, tenant.tenantId, rfqId, quotationId),
        api.listQuotationItems(tenant.token, tenant.tenantId, rfqId, quotationId),
      ]);
      setQuotation(nextQuotation);
      setItems(nextItems.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load quotation");
    }
  }, [quotationId, rfqId, tenant]);

  useEffect(() => {
    void load();
  }, [load]);

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !quotation) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      setQuotation(
        await api.updateQuotation(
          tenant.token,
          tenant.tenantId,
          rfqId,
          quotation.id,
          compactRecord({
            currency_code: formString(formData, "currency_code"),
            lead_time_days: formNumber(formData, "lead_time_days"),
            payment_terms: formString(formData, "payment_terms"),
            notes: formString(formData, "notes"),
          }) as {
            currency_code?: string;
            lead_time_days?: number;
            payment_terms?: string;
            notes?: string;
          },
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update quotation");
    }
  }

  async function submit() {
    if (!tenant || !quotation) return;
    setError(null);
    try {
      setQuotation(await api.submitQuotation(tenant.token, tenant.tenantId, rfqId, quotation.id));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to submit quotation");
    }
  }

  async function reject() {
    if (!tenant || !quotation) return;
    setError(null);
    try {
      setQuotation(await api.rejectQuotation(tenant.token, tenant.tenantId, rfqId, quotation.id, rejectionReason || undefined));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to reject quotation");
    }
  }

  async function updateItem(event: FormEvent<HTMLFormElement>, itemId: string) {
    event.preventDefault();
    if (!tenant || !quotation) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.updateQuotationItem(
        tenant.token,
        tenant.tenantId,
        rfqId,
        quotation.id,
        itemId,
        compactRecord({
          unit_price: formString(formData, "unit_price"),
          delivery_days: formNumber(formData, "delivery_days"),
          notes: formString(formData, "notes"),
        }) as { unit_price?: string; delivery_days?: number; notes?: string },
      );
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update quotation item");
    }
  }

  if (!quotation) {
    return (
      <>
        <PageHeader title="Quotation detail" description="Loading quotation." />
        {error ? <Notice title="Unable to load quotation" tone="danger">{error}</Notice> : null}
      </>
    );
  }

  return (
    <>
      <PageHeader
        eyebrow="Vendor responses"
        title={quotation.vendor_name ?? "Quotation"}
        description={`${quotation.currency_code} · Updated ${formatDateTime(quotation.updated_at)}`}
        action={<StatusBadge status={quotation.status} />}
      />
      {error ? <Notice title="Quotation action failed" tone="danger">{error}</Notice> : null}

      <section className="content-grid">
        <form className="panel form-grid span-8" onSubmit={handleUpdate}>
          <div className="panel-heading wide-field">
            <h2>Quotation details</h2>
          </div>
          <label>
            <span>Currency</span>
            <input name="currency_code" defaultValue={quotation.currency_code} />
          </label>
          <label>
            <span>Lead time days</span>
            <input name="lead_time_days" type="number" min="0" defaultValue={quotation.lead_time_days ?? ""} />
          </label>
          <label>
            <span>Payment terms</span>
            <input name="payment_terms" defaultValue={quotation.payment_terms ?? ""} />
          </label>
          <label className="wide-field">
            <span>Notes</span>
            <input name="notes" defaultValue={quotation.notes ?? ""} />
          </label>
          <div className="form-actions wide-field">
            <Button type="submit">Save quotation</Button>
          </div>
        </form>

        <div className="panel form-stack span-4">
          <div className="panel-heading">
            <h2>Actions</h2>
          </div>
          <Button variant="secondary" onClick={() => void submit()}>Submit</Button>
          <input
            value={rejectionReason}
            onChange={(event) => setRejectionReason(event.target.value)}
            placeholder="Rejection reason"
          />
          <Button variant="danger" onClick={() => void reject()}>Reject</Button>
        </div>
      </section>

      <section className="panel">
        <div className="panel-heading">
          <h2>Pricing</h2>
          <span>{items.length} lines</span>
        </div>
        <DataTable
          rows={items}
          getRowKey={(row) => row.id}
          emptyLabel="No quotation items are available."
          columns={[
            { key: "line", header: "Line", render: (row) => row.line_number },
            { key: "item", header: "Item", render: (row) => row.item_name },
            { key: "qty", header: "Quantity", render: (row) => `${row.quantity} ${row.unit}` },
            {
              key: "pricing",
              header: "Pricing",
              render: (row) => (
                <form className="inline-edit" onSubmit={(event) => void updateItem(event, row.id)}>
                  <input name="unit_price" defaultValue={row.unit_price} aria-label="Unit price" />
                  <input name="delivery_days" type="number" min="0" defaultValue={row.delivery_days ?? ""} aria-label="Delivery days" />
                  <input name="notes" defaultValue={row.notes ?? ""} aria-label="Notes" />
                  <button type="submit">Save</button>
                </form>
              ),
            },
          ]}
        />
      </section>
    </>
  );
}
