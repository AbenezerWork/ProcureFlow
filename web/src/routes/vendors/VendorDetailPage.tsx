import { FormEvent, useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantParams } from "@/shared/hooks/useTenantParams";
import { compactRecord, formString } from "@/shared/lib/form";
import { formatDateTime } from "@/shared/lib/format";
import type { Vendor } from "@/shared/types/api";
import { VendorForm } from "@/routes/vendors/NewVendorPage";

export function VendorDetailPage() {
  const { vendorId = "" } = useParams();
  const tenant = useTenantParams();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!tenant || !vendorId) return;
    setError(null);
    try {
      setVendor(await api.getVendor(tenant.token, tenant.tenantId, vendorId));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load vendor");
    }
  }, [tenant, vendorId]);

  useEffect(() => {
    void load();
  }, [load]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant || !vendor) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      setVendor(
        await api.updateVendor(
          tenant.token,
          tenant.tenantId,
          vendor.id,
          compactRecord({
            name: formString(formData, "name"),
            legal_name: formString(formData, "legal_name"),
            contact_name: formString(formData, "contact_name"),
            email: formString(formData, "email"),
            phone: formString(formData, "phone"),
            country: formString(formData, "country"),
            notes: formString(formData, "notes"),
          }) as Record<string, string>,
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update vendor");
    }
  }

  async function archive() {
    if (!tenant || !vendor) return;
    setError(null);
    try {
      setVendor(await api.archiveVendor(tenant.token, tenant.tenantId, vendor.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to archive vendor");
    }
  }

  if (!vendor) {
    return (
      <>
        <PageHeader title="Vendor detail" description="Loading vendor." />
        {error ? <Notice title="Unable to load vendor" tone="danger">{error}</Notice> : null}
      </>
    );
  }

  return (
    <>
      <PageHeader
        eyebrow="Supplier base"
        title={vendor.name}
        description={`Updated ${formatDateTime(vendor.updated_at)}`}
        action={<StatusBadge status={vendor.status} />}
      />
      {error ? <Notice title="Vendor action failed" tone="danger">{error}</Notice> : null}
      <section className="content-grid">
        <div className="span-8">
          <VendorForm onSubmit={handleSubmit} submitLabel="Save vendor" defaults={vendor} />
        </div>
        <div className="panel form-stack span-4">
          <div className="panel-heading">
            <h2>Vendor actions</h2>
          </div>
          <Button variant="danger" onClick={() => void archive()}>
            Archive vendor
          </Button>
        </div>
      </section>
    </>
  );
}
