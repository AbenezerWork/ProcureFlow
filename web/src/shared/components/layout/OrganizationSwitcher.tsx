import { useOrganization } from "@/features/organizations/organization-context";

export function OrganizationSwitcher() {
  const { organizations, activeOrganizationId, selectOrganization, isLoading } = useOrganization();

  if (isLoading) {
    return <div className="org-switcher is-loading">Loading organizations</div>;
  }

  if (organizations.length === 0) {
    return <div className="org-switcher">No organization selected</div>;
  }

  return (
    <label className="org-switcher">
      <span>Organization</span>
      <select
        value={activeOrganizationId ?? ""}
        onChange={(event) => selectOrganization(event.target.value)}
      >
        {organizations.map((entry) => (
          <option key={entry.organization.id} value={entry.organization.id}>
            {entry.organization.name}
          </option>
        ))}
      </select>
    </label>
  );
}
