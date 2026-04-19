import { FormEvent, useState } from "react";

import { useOrganization } from "@/features/organizations/organization-context";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { StatusBadge } from "@/shared/components/ui/StatusBadge";
import { formatDateTime } from "@/shared/lib/format";

export function OrganizationsPage() {
  const {
    organizations,
    activeOrganizationId,
    selectOrganization,
    createOrganization,
    error,
    isLoading,
  } = useOrganization();
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError(null);
    setIsSubmitting(true);

    const formData = new FormData(event.currentTarget);

    try {
      await createOrganization({
        name: String(formData.get("name") ?? ""),
        slug: String(formData.get("slug") ?? "") || undefined,
      });
      event.currentTarget.reset();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "Unable to create organization");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Tenant setup"
        title="Organizations"
        description="Create workspaces and switch the tenant context used by organization-scoped API calls."
      />

      {error ? <Notice title="Unable to load organizations" tone="danger">{error}</Notice> : null}
      {formError ? <Notice title="Unable to create organization" tone="danger">{formError}</Notice> : null}

      <section className="content-grid">
        <div className="panel span-8">
          <div className="panel-heading">
            <h2>My organizations</h2>
            <span>{isLoading ? "Loading" : `${organizations.length} available`}</span>
          </div>
          <DataTable
            rows={organizations}
            getRowKey={(row) => row.organization.id}
            emptyLabel="Create an organization to begin using ProcureFlow."
            columns={[
              { key: "name", header: "Name", render: (row) => row.organization.name },
              { key: "role", header: "Role", render: (row) => <StatusBadge status={row.role} /> },
              { key: "status", header: "Status", render: (row) => <StatusBadge status={row.status} /> },
              {
                key: "created",
                header: "Created",
                render: (row) => formatDateTime(row.organization.created_at),
              },
              {
                key: "active",
                header: "Tenant",
                render: (row) =>
                  row.organization.id === activeOrganizationId ? (
                    <strong>Selected</strong>
                  ) : (
                    <button type="button" onClick={() => selectOrganization(row.organization.id)}>
                      Select
                    </button>
                  ),
              },
            ]}
          />
        </div>

        <form className="panel form-stack span-4" onSubmit={handleSubmit}>
          <div className="panel-heading">
            <h2>Create organization</h2>
          </div>
          <label>
            <span>Name</span>
            <input name="name" required />
          </label>
          <label>
            <span>Slug</span>
            <input name="slug" placeholder="global-ops" />
          </label>
          <Button disabled={isSubmitting} type="submit">
            {isSubmitting ? "Creating" : "Create organization"}
          </Button>
        </form>
      </section>
    </>
  );
}
