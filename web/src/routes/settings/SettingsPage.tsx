import { useOrganization } from "@/features/organizations/organization-context";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { formatDateTime } from "@/shared/lib/format";

export function SettingsPage() {
  const { activeOrganization } = useOrganization();

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
        <section className="panel detail-list">
          <div className="panel-heading">
            <h2>{activeOrganization.organization.name}</h2>
            <StatusBadge status={activeOrganization.status} />
          </div>
          <dl>
            <div>
              <dt>Organization ID</dt>
              <dd>{activeOrganization.organization.id}</dd>
            </div>
            <div>
              <dt>Slug</dt>
              <dd>{activeOrganization.organization.slug}</dd>
            </div>
            <div>
              <dt>Your role</dt>
              <dd>{activeOrganization.role}</dd>
            </div>
            <div>
              <dt>Created</dt>
              <dd>{formatDateTime(activeOrganization.organization.created_at)}</dd>
            </div>
          </dl>
        </section>
      )}
    </>
  );
}
