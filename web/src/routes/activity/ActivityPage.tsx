import { FormEvent, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";
import { api } from "@/shared/api/client";
import { Button } from "@/shared/components/ui/Button";
import { DataTable } from "@/shared/components/ui/DataTable";
import { Notice } from "@/shared/components/ui/Notice";
import { PageHeader } from "@/shared/components/ui/PageHeader";
import { formatDateTime } from "@/shared/lib/format";
import type { ActivityLog } from "@/shared/types/api";

export function ActivityPage() {
  const { token } = useAuth();
  const { activeOrganizationId } = useOrganization();
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !activeOrganizationId) {
      return;
    }

    const formData = new FormData(event.currentTarget);
    setError(null);
    setIsLoading(true);

    try {
      const response = await api.listActivityLogs(
        token,
        activeOrganizationId,
        String(formData.get("entity_type") ?? ""),
        String(formData.get("entity_id") ?? ""),
      );
      setLogs(response.activity_logs);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load activity");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Audit trail"
        title="Activity"
        description="Load an entity timeline by entity type and entity ID."
      />

      {error ? <Notice title="Unable to load activity" tone="danger">{error}</Notice> : null}

      <section className="content-grid">
        <form className="panel form-stack span-4" onSubmit={handleSubmit}>
          <div className="panel-heading">
            <h2>Timeline query</h2>
          </div>
          <label>
            <span>Entity type</span>
            <input name="entity_type" placeholder="procurement_request" required />
          </label>
          <label>
            <span>Entity ID</span>
            <input name="entity_id" placeholder="UUID" required />
          </label>
          <Button disabled={isLoading} type="submit">
            {isLoading ? "Loading" : "Load activity"}
          </Button>
        </form>

        <div className="panel span-8">
          <div className="panel-heading">
            <h2>Timeline</h2>
            <span>{logs.length} events</span>
          </div>
          <DataTable
            rows={logs}
            getRowKey={(row) => row.id}
            emptyLabel="Submit a query to load an activity timeline."
            columns={[
              { key: "action", header: "Action", render: (row) => row.action },
              { key: "summary", header: "Summary", render: (row) => row.summary ?? "Not set" },
              { key: "when", header: "Occurred", render: (row) => formatDateTime(row.occurred_at) },
            ]}
          />
        </div>
      </section>
    </>
  );
}
