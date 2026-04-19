import { useMemo } from "react";

import { api } from "@/shared/api/client";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import type { OrganizationMember } from "@/shared/types/api";

export function MembersPage() {
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listMemberships(token, organizationId)).memberships,
    [],
  );
  const { data, isLoading, error } = useTenantResource<OrganizationMember[]>(loader);
  const members = data ?? [];

  return (
    <>
      <PageHeader
        eyebrow="Access"
        title="Members"
        description="Review the active organization membership roster, roles, and access status."
      />

      {error ? <Notice title="Unable to load members" tone="danger">{error}</Notice> : null}

      <section className="panel">
        <div className="panel-heading">
          <h2>Memberships</h2>
          <span>{isLoading ? "Loading" : `${members.length} members`}</span>
        </div>
        <DataTable
          rows={members}
          getRowKey={(row) => row.membership.id}
          emptyLabel="No members are visible for this organization."
          columns={[
            { key: "name", header: "Name", render: (row) => row.user.full_name },
            { key: "email", header: "Email", render: (row) => row.user.email },
            { key: "role", header: "Role", render: (row) => <StatusBadge status={row.membership.role} /> },
            { key: "status", header: "Status", render: (row) => <StatusBadge status={row.membership.status} /> },
            { key: "joined", header: "Invited", render: (row) => formatDateTime(row.membership.invited_at) },
          ]}
        />
      </section>
    </>
  );
}
