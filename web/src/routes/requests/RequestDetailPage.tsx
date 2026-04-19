import { FormEvent, useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formString } from "@/shared/lib/form";
import { formatAmount, formatDateTime } from "@/shared/lib/format";
import type { ProcurementItem, ProcurementRequest } from "@/shared/types/api";

export function RequestDetailPage() {
  const { requestId = "" } = useParams();
  const tenant = useTenantParams();
  const [request, setRequest] = useState<ProcurementRequest | null>(null);
  const [items, setItems] = useState<ProcurementItem[]>([]);
  const [decisionComment, setDecisionComment] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const load = useCallback(async () => {
    if (!tenant || !requestId) return;
    setIsLoading(true);
    setError(null);
    try {
      const [nextRequest, nextItems] = await Promise.all([
        api.getProcurementRequest(tenant.token, tenant.tenantId, requestId),
        api.listProcurementRequestItems(tenant.token, tenant.tenantId, requestId),
      ]);
      setRequest(nextRequest);
      setItems(nextItems.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load request");
    } finally {
      setIsLoading(false);
    }
  }, [requestId, tenant]);

  useEffect(() => {
    void load();
  }, [load]);

  async function mutate(action: () => Promise<ProcurementRequest>) {
    setError(null);
    try {
      setRequest(await action());
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Action failed");
    }
  }

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !request) return;
    const formData = new FormData(event.currentTarget);
    await mutate(() =>
      api.updateProcurementRequest(
        tenant.token,
        tenant.tenantId,
        request.id,
        compactRecord({
          title: formString(formData, "title"),
          description: formString(formData, "description"),
          justification: formString(formData, "justification"),
          currency_code: formString(formData, "currency_code"),
          estimated_total_amount: formString(formData, "estimated_total_amount"),
        }),
      ),
    );
  }

  async function handleCreateItem(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !request) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.createProcurementRequestItem(tenant.token, tenant.tenantId, request.id, {
        item_name: formString(formData, "item_name") ?? "",
        quantity: formString(formData, "quantity") ?? "",
        unit: formString(formData, "unit") ?? "",
        description: formString(formData, "description"),
        estimated_unit_price: formString(formData, "estimated_unit_price"),
        needed_by_date: formString(formData, "needed_by_date"),
      });
      event.currentTarget.reset();
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to add item");
    }
  }

  async function deleteItem(itemId: string) {
    if (!tenant || !request) return;
    setError(null);
    try {
      await api.deleteProcurementRequestItem(tenant.token, tenant.tenantId, request.id, itemId);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to delete item");
    }
  }

  async function updateItem(event: FormEvent<HTMLFormElement>, itemId: string) {
    event.preventDefault();
    if (!tenant || !request) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.updateProcurementRequestItem(
        tenant.token,
        tenant.tenantId,
        request.id,
        itemId,
        compactRecord({
          item_name: formString(formData, "item_name"),
          quantity: formString(formData, "quantity"),
          unit: formString(formData, "unit"),
          estimated_unit_price: formString(formData, "estimated_unit_price"),
          needed_by_date: formString(formData, "needed_by_date"),
        }),
      );
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update item");
    }
  }

  if (!request) {
    return (
      <>
        <PageHeader title="Request detail" description={isLoading ? "Loading request." : "No request loaded."} />
        {error ? <Notice title="Unable to load request" tone="danger">{error}</Notice> : null}
      </>
    );
  }

  return (
    <>
      <PageHeader
        eyebrow="Demand intake"
        title={request.title}
        description={`${formatAmount(request.estimated_total_amount, request.currency_code)} · Updated ${formatDateTime(request.updated_at)}`}
        action={<StatusBadge status={request.status} />}
      />

      {error ? <Notice title="Request action failed" tone="danger">{error}</Notice> : null}

      <section className="content-grid">
        <form className="panel form-grid span-8" onSubmit={handleUpdate}>
          <div className="panel-heading wide-field">
            <h2>Request details</h2>
          </div>
          <label>
            <span>Title</span>
            <input name="title" defaultValue={request.title} />
          </label>
          <label>
            <span>Currency</span>
            <input name="currency_code" defaultValue={request.currency_code} />
          </label>
          <label>
            <span>Estimated total</span>
            <input name="estimated_total_amount" defaultValue={request.estimated_total_amount ?? ""} />
          </label>
          <label className="wide-field">
            <span>Description</span>
            <input name="description" defaultValue={request.description ?? ""} />
          </label>
          <label className="wide-field">
            <span>Justification</span>
            <input name="justification" defaultValue={request.justification ?? ""} />
          </label>
          <div className="form-actions wide-field">
            <Button type="submit">Save request</Button>
          </div>
        </form>

        <div className="panel form-stack span-4">
          <div className="panel-heading">
            <h2>Lifecycle</h2>
          </div>
          <Button variant="secondary" onClick={() => void mutate(() => api.submitProcurementRequest(tenant!.token, tenant!.tenantId, request.id))}>
            Submit
          </Button>
          <input
            value={decisionComment}
            onChange={(event) => setDecisionComment(event.target.value)}
            placeholder="Decision comment"
          />
          <Button variant="secondary" onClick={() => void mutate(() => api.approveProcurementRequest(tenant!.token, tenant!.tenantId, request.id, decisionComment || undefined))}>
            Approve
          </Button>
          <Button variant="danger" onClick={() => void mutate(() => api.rejectProcurementRequest(tenant!.token, tenant!.tenantId, request.id, decisionComment || undefined))}>
            Reject
          </Button>
          <Button variant="danger" onClick={() => void mutate(() => api.cancelProcurementRequest(tenant!.token, tenant!.tenantId, request.id))}>
            Cancel
          </Button>
        </div>
      </section>

      <section className="panel">
        <div className="panel-heading">
          <h2>Request items</h2>
          <span>{items.length} lines</span>
        </div>
        <DataTable
          rows={items}
          getRowKey={(row) => row.id}
          emptyLabel="No items have been added."
          columns={[
            { key: "line", header: "Line", render: (row) => row.line_number },
            { key: "name", header: "Item", render: (row) => row.item_name },
            { key: "qty", header: "Quantity", render: (row) => `${row.quantity} ${row.unit}` },
            { key: "price", header: "Unit estimate", render: (row) => row.estimated_unit_price ?? "Not set" },
            { key: "date", header: "Needed by", render: (row) => row.needed_by_date ?? "Not set" },
            {
              key: "edit",
              header: "Edit",
              render: (row) => (
                <form className="inline-edit request-item-edit" onSubmit={(event) => void updateItem(event, row.id)}>
                  <input name="item_name" defaultValue={row.item_name} aria-label="Item name" />
                  <input name="quantity" defaultValue={row.quantity} aria-label="Quantity" />
                  <input name="unit" defaultValue={row.unit} aria-label="Unit" />
                  <input name="estimated_unit_price" defaultValue={row.estimated_unit_price ?? ""} aria-label="Estimated unit price" />
                  <input name="needed_by_date" type="date" defaultValue={row.needed_by_date ?? ""} aria-label="Needed by" />
                  <button type="submit">Save</button>
                  <button type="button" onClick={() => void deleteItem(row.id)}>Delete</button>
                </form>
              ),
            },
          ]}
        />
      </section>

      <form className="panel form-grid" onSubmit={handleCreateItem}>
        <div className="panel-heading wide-field">
          <h2>Add item</h2>
        </div>
        <label>
          <span>Item name</span>
          <input name="item_name" required />
        </label>
        <label>
          <span>Quantity</span>
          <input name="quantity" required />
        </label>
        <label>
          <span>Unit</span>
          <input name="unit" required />
        </label>
        <label>
          <span>Estimated unit price</span>
          <input name="estimated_unit_price" />
        </label>
        <label>
          <span>Needed by</span>
          <input name="needed_by_date" type="date" />
        </label>
        <label className="wide-field">
          <span>Description</span>
          <input name="description" />
        </label>
        <div className="form-actions wide-field">
          <Button type="submit">Add item</Button>
        </div>
      </form>
    </>
  );
}
