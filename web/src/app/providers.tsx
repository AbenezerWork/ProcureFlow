import type { ReactNode } from "react";

import { AuthProvider } from "@/features/auth/auth-context";
import { OrganizationProvider } from "@/features/organizations/organization-context";

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <AuthProvider>
      <OrganizationProvider>{children}</OrganizationProvider>
    </AuthProvider>
  );
}
