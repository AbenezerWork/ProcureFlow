import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import type { ReactNode } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { api } from "@/shared/api/client";
import type { ID, UserOrganization } from "@/shared/types/api";

const activeOrganizationStorageKey = "procureflow.activeOrganizationId";

type OrganizationContextValue = {
  organizations: UserOrganization[];
  activeOrganization: UserOrganization | null;
  activeOrganizationId: ID | null;
  isLoading: boolean;
  error: string | null;
  refreshOrganizations: () => Promise<void>;
  selectOrganization: (organizationId: ID) => void;
  createOrganization: (input: { name: string; slug?: string }) => Promise<void>;
};

const OrganizationContext = createContext<OrganizationContextValue | null>(null);

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const { token, isAuthenticated } = useAuth();
  const [organizations, setOrganizations] = useState<UserOrganization[]>([]);
  const [activeOrganizationId, setActiveOrganizationId] = useState<ID | null>(() =>
    window.localStorage.getItem(activeOrganizationStorageKey),
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectOrganization = useCallback((organizationId: ID) => {
    window.localStorage.setItem(activeOrganizationStorageKey, organizationId);
    setActiveOrganizationId(organizationId);
  }, []);

  const refreshOrganizations = useCallback(async () => {
    if (!token) {
      setOrganizations([]);
      setActiveOrganizationId(null);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const response = await api.listOrganizations(token);
      setOrganizations(response.organizations);

      const active = response.organizations.find(
        (entry) => entry.organization.id === activeOrganizationId,
      );
      const firstActive =
        response.organizations.find((entry) => entry.status === "active") ??
        response.organizations[0] ??
        null;

      if (!active && firstActive) {
        selectOrganization(firstActive.organization.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load organizations");
    } finally {
      setIsLoading(false);
    }
  }, [activeOrganizationId, selectOrganization, token]);

  const createOrganization = useCallback(
    async (input: { name: string; slug?: string }) => {
      if (!token) {
        throw new Error("Authentication required");
      }

      await api.createOrganization(token, input);
      await refreshOrganizations();
    },
    [refreshOrganizations, token],
  );

  useEffect(() => {
    if (isAuthenticated) {
      void refreshOrganizations();
    } else {
      setOrganizations([]);
      setActiveOrganizationId(null);
      window.localStorage.removeItem(activeOrganizationStorageKey);
    }
  }, [isAuthenticated, refreshOrganizations]);

  const activeOrganization = useMemo(
    () =>
      organizations.find((entry) => entry.organization.id === activeOrganizationId) ??
      null,
    [activeOrganizationId, organizations],
  );

  const value = useMemo<OrganizationContextValue>(
    () => ({
      organizations,
      activeOrganization,
      activeOrganizationId: activeOrganization?.organization.id ?? null,
      isLoading,
      error,
      refreshOrganizations,
      selectOrganization,
      createOrganization,
    }),
    [
      activeOrganization,
      createOrganization,
      error,
      isLoading,
      organizations,
      refreshOrganizations,
      selectOrganization,
    ],
  );

  return <OrganizationContext.Provider value={value}>{children}</OrganizationContext.Provider>;
}

export function useOrganization() {
  const value = useContext(OrganizationContext);
  if (!value) {
    throw new Error("useOrganization must be used inside OrganizationProvider");
  }

  return value;
}
