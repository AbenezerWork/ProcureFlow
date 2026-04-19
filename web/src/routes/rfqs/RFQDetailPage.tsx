import { FormEvent, useCallback, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formNumber, formString } from "@/shared/lib/form";
import { formatDateTime } from "@/shared/lib/format";
import type { Quotation, RFQ, RFQItem, RFQVendor, Vendor } from "@/shared/types/api";

export function RFQDetailPage() {
  const { rfqId = "" } = useParams();
  const tenant = useTenantParams();
  const navigate = useNavigate();
  const [rfq, setRFQ] = useState<RFQ | null>(null);
  const [items, setItems] = useState<RFQItem[]>([]);
  const [rfqVendors, setRFQVendors] = useState<RFQVendor[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [quotations, setQuotations] = useState<Quotation[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!tenant || !rfqId) return;
    setError(null);
    try {
      const [nextRFQ, nextItems, nextRFQVendors, nextVendors, nextQuotations] = await Promise.all([
        api.getRFQ(tenant.token, tenant.tenantId, rfqId),
        api.listRFQItems(tenant.token, tenant.tenantId, rfqId),
        api.listRFQVendors(tenant.token, tenant.tenantId, rfqId),
        api.listVendors(tenant.token, tenant.tenantId),
        api.listQuotations(tenant.token, tenant.tenantId, rfqId),
      ]);
      setRFQ(nextRFQ);
      setItems(nextItems.items);
      setRFQVendors(nextRFQVendors.vendors);
      setVendors(nextVendors.vendors);
      setQuotations(nextQuotations.quotations);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load RFQ");
    }
  }, [rfqId, tenant]);

  useEffect(() => {
    void load();
  }, [load]);

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !rfq) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      setRFQ(
        await api.updateRFQ(
          tenant.token,
          tenant.tenantId,
          rfq.id,
          compactRecord({
            reference_number: formString(formData, "reference_number"),
            title: formString(formData, "title"),
            description: formString(formData, "description"),
          }),
        ),
      );
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update RFQ");
    }
  }

  async function transition(action: "publish" | "close" | "evaluate" | "cancel") {
    if (!tenant || !rfq) return;
    setError(null);
    try {
      setRFQ(await api.transitionRFQ(tenant.token, tenant.tenantId, rfq.id, action));
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : `Unable to ${action} RFQ`);
    }
  }

  async function attachVendor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !rfq) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.attachRFQVendor(tenant.token, tenant.tenantId, rfq.id, formString(formData, "vendor_id") ?? "");
      event.currentTarget.reset();
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to attach vendor");
    }
  }

  async function removeVendor(vendorId: string) {
    if (!tenant || !rfq) return;
    setError(null);
    try {
      await api.removeRFQVendor(tenant.token, tenant.tenantId, rfq.id, vendorId);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to remove vendor");
    }
  }

  async function createQuotation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !rfq) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      const quotation = await api.createQuotation(
        tenant.token,
        tenant.tenantId,
        rfq.id,
        compactRecord({
          rfq_vendor_id: formString(formData, "rfq_vendor_id") ?? "",
          currency_code: formString(formData, "currency_code"),
          lead_time_days: formNumber(formData, "lead_time_days"),
          payment_terms: formString(formData, "payment_terms"),
          notes: formString(formData, "notes"),
        }) as {
          rfq_vendor_id: string;
          currency_code?: string;
          lead_time_days?: number;
          payment_terms?: string;
          notes?: string;
        },
      );
      navigate(`/app/rfqs/${rfq.id}/quotations/${quotation.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create quotation");
    }
  }

  if (!rfq) {
    return (
      <>
        <PageHeader title="RFQ detail" description="Loading RFQ." />
        {error ? <Notice title="Unable to load RFQ" tone="danger">{error}</Notice> : null}
      </>
    );
  }

  return (
    <>
      <PageHeader
        eyebrow="Sourcing"
        title={rfq.title}
        description={`${rfq.reference_number ?? "No reference"} · Updated ${formatDateTime(rfq.updated_at)}`}
        action={<StatusBadge status={rfq.status} />}
      />
      {error ? <Notice title="RFQ action failed" tone="danger">{error}</Notice> : null}

      <section className="content-grid">
        <form className="panel form-grid span-8" onSubmit={handleUpdate}>
          <div className="panel-heading wide-field">
            <h2>RFQ details</h2>
          </div>
          <label>
            <span>Reference</span>
            <input name="reference_number" defaultValue={rfq.reference_number ?? ""} />
          </label>
          <label>
            <span>Title</span>
            <input name="title" defaultValue={rfq.title} />
          </label>
          <label className="wide-field">
            <span>Description</span>
            <input name="description" defaultValue={rfq.description ?? ""} />
          </label>
          <div className="form-actions wide-field">
            <Button type="submit">Save RFQ</Button>
          </div>
        </form>

        <div className="panel form-stack span-4">
          <div className="panel-heading">
            <h2>Lifecycle</h2>
          </div>
          <Button variant="secondary" onClick={() => void transition("publish")}>Publish</Button>
          <Button variant="secondary" onClick={() => void transition("close")}>Close</Button>
          <Button variant="secondary" onClick={() => void transition("evaluate")}>Evaluate</Button>
          <Button variant="danger" onClick={() => void transition("cancel")}>Cancel</Button>
          <Link className="button button-secondary link-button" to={`/app/rfqs/${rfq.id}/comparison`}>Compare quotations</Link>
          <Link className="button button-secondary link-button" to={`/app/rfqs/${rfq.id}/award`}>Award</Link>
        </div>
      </section>

      <section className="content-grid">
        <div className="panel span-8">
          <div className="panel-heading">
            <h2>RFQ items</h2>
            <span>{items.length} lines</span>
          </div>
          <DataTable
            rows={items}
            getRowKey={(row) => row.id}
            emptyLabel="No RFQ item snapshot is available."
            columns={[
              { key: "line", header: "Line", render: (row) => row.line_number },
              { key: "name", header: "Item", render: (row) => row.item_name },
              { key: "qty", header: "Quantity", render: (row) => `${row.quantity} ${row.unit}` },
              { key: "date", header: "Target", render: (row) => row.target_date ?? "Not set" },
            ]}
          />
        </div>
        <form className="panel form-stack span-4" onSubmit={attachVendor}>
          <div className="panel-heading">
            <h2>Attach vendor</h2>
          </div>
          <label>
            <span>Vendor</span>
            <select name="vendor_id" required>
              <option value="">Select vendor</option>
              {vendors.map((vendor) => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.name}
                </option>
              ))}
            </select>
          </label>
          <Button type="submit">Attach vendor</Button>
        </form>
      </section>

      <section className="panel">
        <div className="panel-heading">
          <h2>Attached vendors</h2>
          <span>{rfqVendors.length} vendors</span>
        </div>
        <DataTable
          rows={rfqVendors}
          getRowKey={(row) => row.id}
          emptyLabel="No vendors are attached to this RFQ."
          columns={[
            { key: "vendor", header: "Vendor", render: (row) => row.vendor_name },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.vendor_status} /> },
            { key: "attached", header: "Attached", render: (row) => formatDateTime(row.attached_at) },
            { key: "actions", header: "Actions", render: (row) => <button type="button" onClick={() => void removeVendor(row.vendor_id)}>Remove</button> },
          ]}
        />
      </section>

      <section className="content-grid">
        <div className="panel span-8">
          <div className="panel-heading">
            <h2>Quotations</h2>
            <span>{quotations.length} quotes</span>
          </div>
          <DataTable
            rows={quotations}
            getRowKey={(row) => row.id}
            emptyLabel="No quotations have been created for this RFQ."
            columns={[
              { key: "vendor", header: "Vendor", render: (row) => row.vendor_name ?? "Not set" },
              { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
              { key: "currency", header: "Currency", render: (row) => row.currency_code },
              { key: "actions", header: "Actions", render: (row) => <Link to={`/app/rfqs/${rfq.id}/quotations/${row.id}`}>Open</Link> },
            ]}
          />
        </div>
        <form className="panel form-stack span-4" onSubmit={createQuotation}>
          <div className="panel-heading">
            <h2>Create quotation</h2>
          </div>
          <label>
            <span>RFQ vendor</span>
            <select name="rfq_vendor_id" required>
              <option value="">Select vendor</option>
              {rfqVendors.map((vendor) => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.vendor_name}
                </option>
              ))}
            </select>
          </label>
          <label>
            <span>Currency</span>
            <input name="currency_code" placeholder="USD" />
          </label>
          <label>
            <span>Lead time days</span>
            <input name="lead_time_days" type="number" min="0" />
          </label>
          <label>
            <span>Payment terms</span>
            <input name="payment_terms" />
          </label>
          <Button type="submit">Create quotation</Button>
        </form>
      </section>
    </>
  );
}
