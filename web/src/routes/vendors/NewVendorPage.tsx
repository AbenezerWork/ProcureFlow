import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formString } from "@/shared/lib/form";

export function NewVendorPage() {
  const tenant = useTenantParams();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      const vendor = await api.createVendor(
        tenant.token,
        tenant.tenantId,
        compactRecord({
          name: formString(formData, "name") ?? "",
          legal_name: formString(formData, "legal_name"),
          contact_name: formString(formData, "contact_name"),
          email: formString(formData, "email"),
          phone: formString(formData, "phone"),
          country: formString(formData, "country"),
          notes: formString(formData, "notes"),
        }) as Record<string, string>,
      );
      navigate(`/app/vendors/${vendor.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to create vendor");
    }
  }

  return (
    <>
      <PageHeader eyebrow="Supplier base" title="New vendor" description="Create a vendor record for sourcing events." />
      {error ? <Notice title="Unable to create vendor" tone="danger">{error}</Notice> : null}
      <VendorForm onSubmit={handleSubmit} submitLabel="Create vendor" />
    </>
  );
}

export function VendorForm({
  onSubmit,
  submitLabel,
  defaults = {},
}: {
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  submitLabel: string;
  defaults?: Record<string, string | null | undefined>;
}) {
  return (
    <form className="panel form-grid" onSubmit={onSubmit}>
      <label>
        <span>Name</span>
        <input name="name" defaultValue={defaults.name ?? ""} required />
      </label>
      <label>
        <span>Legal name</span>
        <input name="legal_name" defaultValue={defaults.legal_name ?? ""} />
      </label>
      <label>
        <span>Contact</span>
        <input name="contact_name" defaultValue={defaults.contact_name ?? ""} />
      </label>
      <label>
        <span>Email</span>
        <input name="email" type="email" defaultValue={defaults.email ?? ""} />
      </label>
      <label>
        <span>Phone</span>
        <input name="phone" defaultValue={defaults.phone ?? ""} />
      </label>
      <label>
        <span>Country</span>
        <input name="country" defaultValue={defaults.country ?? ""} />
      </label>
      <label className="wide-field">
        <span>Notes</span>
        <input name="notes" defaultValue={defaults.notes ?? ""} />
      </label>
      <div className="form-actions wide-field">
        <Button type="submit">{submitLabel}</Button>
      </div>
    </form>
  );
}
