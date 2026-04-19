import { FormEvent, useCallback, useMemo, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";
import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { useTenantResource } from "@/shared/hooks/useTenantResource";
import { formatDateTime } from "@/shared/lib/format";
import { formString } from "@/shared/lib/form";
import type { OrganizationMember } from "@/shared/types/api";

export function MembersPage() {
  const { token } = useAuth();
  const { activeOrganizationId } = useOrganization();
  const loader = useMemo(
    () => async (token: string, organizationId: string) =>
      (await api.listMemberships(token, organizationId)).memberships,
    [],
  );
  const { data, isLoading, error } = useTenantResource<OrganizationMember[]>(loader);
  const [actionError, setActionError] = useState<string | null>(null);
  const members = data ?? [];

  const reload = useCallback(() => {
    window.location.reload();
  }, []);

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !activeOrganizationId) return;
    const formData = new FormData(event.currentTarget);
    setActionError(null);
    try {
      await api.createMembership(token, activeOrganizationId, {
        email: formString(formData, "email"),
        user_id: formString(formData, "user_id"),
        role: formString(formData, "role") ?? "viewer",
        status: formString(formData, "status"),
      });
      event.currentTarget.reset();
      reload();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Unable to create membership");
    }
  }

  async function handleUpdate(event: FormEvent<HTMLFormElement>, userId: string) {
    event.preventDefault();
    if (!token || !activeOrganizationId) return;
    const formData = new FormData(event.currentTarget);
    setActionError(null);
    try {
      await api.updateMembership(token, activeOrganizationId, userId, {
        role: formString(formData, "role"),
        status: formString(formData, "status"),
      });
      reload();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Unable to update membership");
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Access"
        title="Members"
        description="Review the active organization membership roster, roles, and access status."
      />

      {error ? <Notice title="Unable to load members" tone="danger">{error}</Notice> : null}
      {actionError ? <Notice title="Membership action failed" tone="danger">{actionError}</Notice> : null}

      <section className="content-grid">
        <div className="panel span-8">
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
              {
                key: "edit",
                header: "Edit",
                render: (row) => (
                  <form className="inline-edit" onSubmit={(event) => void handleUpdate(event, row.user.id)}>
                    <select name="role" defaultValue={row.membership.role}>
                      <option value="owner">owner</option>
                      <option value="admin">admin</option>
                      <option value="requester">requester</option>
                      <option value="approver">approver</option>
                      <option value="procurement_officer">procurement_officer</option>
                      <option value="viewer">viewer</option>
                    </select>
                    <select name="status" defaultValue={row.membership.status}>
                      <option value="invited">invited</option>
                      <option value="active">active</option>
                      <option value="suspended">suspended</option>
                      <option value="removed">removed</option>
                    </select>
                    <button type="submit">Save</button>
                  </form>
                ),
              },
            ]}
          />
        </div>
        <form className="panel form-stack span-4" onSubmit={handleCreate}>
          <div className="panel-heading">
            <h2>Add member</h2>
          </div>
          <label>
            <span>Email</span>
            <input name="email" type="email" />
          </label>
          <label>
            <span>User ID</span>
            <input name="user_id" placeholder="Optional existing user UUID" />
          </label>
          <label>
            <span>Role</span>
            <select name="role" defaultValue="viewer">
              <option value="admin">admin</option>
              <option value="requester">requester</option>
              <option value="approver">approver</option>
              <option value="procurement_officer">procurement_officer</option>
              <option value="viewer">viewer</option>
            </select>
          </label>
          <label>
            <span>Status</span>
            <select name="status" defaultValue="active">
              <option value="invited">invited</option>
              <option value="active">active</option>
            </select>
          </label>
          <Button type="submit">Add member</Button>
        </form>
      </section>
    </>
  );
}
