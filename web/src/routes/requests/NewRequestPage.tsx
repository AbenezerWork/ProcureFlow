import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formString } from "@/shared/lib/form";

export function NewRequestPage() {
  const tenant = useTenantParams();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant) return;

    setError(null);
    setIsSubmitting(true);
    const formData = new FormData(event.currentTarget);

    try {
      const request = await api.createProcurementRequest(
        tenant.token,
        tenant.tenantId,
        compactRecord({
          title: formString(formData, "title") ?? "",
          description: formString(formData, "description"),
          justification: formString(formData, "justification"),
          currency_code: formString(formData, "currency_code"),
          estimated_total_amount: formString(formData, "estimated_total_amount"),
        }) as {
          title: string;
          description?: string;
          justification?: string;
          currency_code?: string;
          estimated_total_amount?: string;
        },
      );
      navigate(`/app/requests/${request.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create request");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Demand intake"
        title="New request"
        description="Create a draft procurement request before adding line items and submitting it."
      />
      {error ? <Notice title="Unable to create request" tone="danger">{error}</Notice> : null}
      <form className="panel form-grid" onSubmit={handleSubmit}>
        <label>
          <span>Title</span>
          <input name="title" required />
        </label>
        <label>
          <span>Currency</span>
          <input name="currency_code" placeholder="USD" />
        </label>
        <label>
          <span>Estimated total</span>
          <input name="estimated_total_amount" placeholder="10000.00" />
        </label>
        <label className="wide-field">
          <span>Description</span>
          <input name="description" />
        </label>
        <label className="wide-field">
          <span>Justification</span>
          <input name="justification" />
        </label>
        <div className="form-actions">
          <Button disabled={isSubmitting} type="submit">
            {isSubmitting ? "Creating" : "Create request"}
          </Button>
        </div>
      </form>
    </>
  );
}
