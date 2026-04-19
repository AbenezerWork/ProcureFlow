import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";

export function useTenantParams() {
  const { token } = useAuth();
  const { activeOrganizationId } = useOrganization();

  if (!token || !activeOrganizationId) {
    return null;
  }

  return { token, tenantId: activeOrganizationId };
}
