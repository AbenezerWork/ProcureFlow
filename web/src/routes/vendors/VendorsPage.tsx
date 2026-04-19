import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import type { Vendor } from "@/shared/types/api";

export function VendorsPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listVendors(token, organizationId)).vendors,
    [],
  );
  const { data, isLoading, error } = useTenantResource<Vendor[]>(loader);
  const vendors = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Supplier base"
        title="Vendors"
        description="Manage vendor records that can be attached to sourcing events."
      />

      {error ? <Notice title="Unable to load vendors" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>Vendor directory</h2>
          <span>{isLoading ? "Loading" : `${vendors.length} vendors`}</span>
        </div>
        <DataTable
          rows={vendors}
          getRowKey={(row) => row.id}
          emptyLabel="No vendors have been added for this organization."
          columns={[
            { key: "name", header: "Vendor", render: (row) => row.name },
            { key: "contact", header: "Contact", render: (row) => row.contact_name ?? "Not set" },
            { key: "email", header: "Email", render: (row) => row.email ?? "Not set" },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
            { key: "updated", header: "Updated", render: (row) => formatDateTime(row.updated_at) },
          ]}
        />
      </section>
    </>
  );
}
