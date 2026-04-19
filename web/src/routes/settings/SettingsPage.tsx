import { FormEvent, useEffect, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";
import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { compactRecord, formString } from "@/shared/lib/form";
import { formatDateTime } from "@/shared/lib/format";

export function SettingsPage() {
  const { token } = useAuth();
  const { activeOrganization, refreshOrganizations } = useOrganization();
  const [loadedOrganizationName, setLoadedOrganizationName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadOrganizationDetails() {
      if (!token || !activeOrganization) return;
      try {
        const details = await api.getOrganization(token, activeOrganization.organization.id);
        setLoadedOrganizationName(details.organization.name);
      } catch {
        setLoadedOrganizationName(null);
      }
    }

    void loadOrganizationDetails();
  }, [activeOrganization, token]);

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !activeOrganization) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.updateOrganization(
        token,
        activeOrganization.organization.id,
        compactRecord({
          name: formString(formData, "name"),
          slug: formString(formData, "slug"),
        }),
      );
      await refreshOrganizations();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to update organization");
    }
  }

  async function handleTransfer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !activeOrganization) return;
    const formData = new FormData(event.currentTarget);
    setError(null);
    try {
      await api.transferOrganizationOwnership(token, activeOrganization.organization.id, {
        target_user_id: formString(formData, "target_user_id") ?? "",
        current_owner_new_role: formString(formData, "current_owner_new_role"),
      });
      await refreshOrganizations();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to transfer ownership");
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Workspace"
        title="Settings"
        description="Review the active organization context used for tenant-scoped API requests."
      />

      {!activeOrganization ? (
        <Notice title="No active organization">Create or select an organization before configuring workspace settings.</Notice>
      ) : (
        <>
          {error ? <Notice title="Organization action failed" tone="danger">{error}</Notice> : null}
          <section className="content-grid">
            <form className="panel form-stack span-4" onSubmit={handleUpdate}>
              <div className="panel-heading">
                <h2>Edit organization</h2>
                <StatusBadge status={activeOrganization.status} />
              </div>
              <label>
                <span>Name</span>
                <input name="name" defaultValue={activeOrganization.organization.name} />
              </label>
              <label>
                <span>Slug</span>
                <input name="slug" defaultValue={activeOrganization.organization.slug} />
              </label>
              <Button type="submit">Save organization</Button>
            </form>

            <form className="panel form-stack span-4" onSubmit={handleTransfer}>
              <div className="panel-heading">
                <h2>Transfer ownership</h2>
              </div>
              <label>
                <span>Target user ID</span>
                <input name="target_user_id" required />
              </label>
              <label>
                <span>Your new role</span>
                <select name="current_owner_new_role" defaultValue="admin">
                  <option value="admin">admin</option>
                  <option value="requester">requester</option>
                  <option value="approver">approver</option>
                  <option value="procurement_officer">procurement_officer</option>
                  <option value="viewer">viewer</option>
                </select>
              </label>
              <Button variant="danger" type="submit">Transfer ownership</Button>
            </form>

            <div className="panel detail-list span-4">
              <div className="panel-heading">
                <h2>{loadedOrganizationName ?? activeOrganization.organization.name}</h2>
              </div>
              <dl>
                <div><dt>Organization ID</dt><dd>{activeOrganization.organization.id}</dd></div>
                <div><dt>Slug</dt><dd>{activeOrganization.organization.slug}</dd></div>
                <div><dt>Your role</dt><dd>{activeOrganization.role}</dd></div>
                <div><dt>Created</dt><dd>{formatDateTime(activeOrganization.organization.created_at)}</dd></div>
              </dl>
            </div>
          </section>
        </>
      )}
    </>
  );
}
