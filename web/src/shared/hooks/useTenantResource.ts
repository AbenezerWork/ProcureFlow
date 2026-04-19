import { useCallback, useEffect, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { useOrganization } from "@/features/organizations/organization-context";

type ResourceState<T> = {
  data: T | null;
  isLoading: boolean;
  error: string | null;
  reload: () => Promise<void>;
};

export function useTenantResource<T>(
  loader: (token: string, organizationId: string) => Promise<T>,
  enabled = true,
): ResourceState<T> {
  const { token } = useAuth();
  const { activeOrganizationId } = useOrganization();
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    if (!enabled || !token || !activeOrganizationId) {
      setData(null);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      setData(await loader(token, activeOrganizationId));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load data");
    } finally {
      setIsLoading(false);
    }
  }, [activeOrganizationId, enabled, loader, token]);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { data, isLoading, error, reload };
}
