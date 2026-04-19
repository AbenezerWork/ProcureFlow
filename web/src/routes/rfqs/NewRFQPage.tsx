import { FormEvent, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { compactRecord, formString } from "@/shared/lib/form";
import type { ProcurementRequest } from "@/shared/types/api";

export function NewRFQPage() {
  const tenant = useTenantParams();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listProcurementRequests(token, organizationId, "approved")).procurement_requests,
    [],
  );
  const { data: requests } = useTenantResource<ProcurementRequest[]>(loader);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      const rfq = await api.createRFQ(
        tenant.token,
        tenant.tenantId,
        compactRecord({
          procurement_request_id: formString(formData, "procurement_request_id") ?? "",
          reference_number: formString(formData, "reference_number"),
          title: formString(formData, "title"),
          description: formString(formData, "description"),
        }) as {
          procurement_request_id: string;
          reference_number?: string;
          title?: string;
          description?: string;
        },
      );
      navigate(`/app/rfqs/${rfq.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create RFQ");
    }
  }

  return (
    <>
      <PageHeader eyebrow="Sourcing" title="New RFQ" description="Create an RFQ from an approved procurement request." />
      {error ? <Notice title="Unable to create RFQ" tone="danger">{error}</Notice> : null}
      <form className="panel form-grid" onSubmit={handleSubmit}>
        <label className="wide-field">
          <span>Approved request</span>
          <select name="procurement_request_id" required>
            <option value="">Select a request</option>
            {(requests ?? []).map((request) => (
              <option key={request.id} value={request.id}>
                {request.title}
              </option>
            ))}
          </select>
        </label>
        <label>
          <span>Reference</span>
          <input name="reference_number" />
        </label>
        <label>
          <span>Title</span>
          <input name="title" />
        </label>
        <label className="wide-field">
          <span>Description</span>
          <input name="description" />
        </label>
        <div className="form-actions wide-field">
          <Button type="submit">Create RFQ</Button>
        </div>
      </form>
    </>
  );
}
