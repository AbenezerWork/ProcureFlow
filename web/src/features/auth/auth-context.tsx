import { createContext, useCallback, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";

import { api } from "@/shared/api/client";
import {
  clearStoredSession,
  loadStoredSession,
  storeSession,
} from "@/shared/auth/session-storage";
import type { Session, User } from "@/shared/types/api";

type AuthContextValue = {
  session: Session | null;
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (input: { email: string; password: string }) => Promise<void>;
  register: (input: { email: string; password: string; full_name: string }) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(() => loadStoredSession());

  const saveSession = useCallback((nextSession: Session) => {
    storeSession(nextSession);
    setSession(nextSession);
  }, []);

  const login = useCallback(
    async (input: { email: string; password: string }) => {
      saveSession(await api.login(input));
    },
    [saveSession],
  );

  const register = useCallback(
    async (input: { email: string; password: string; full_name: string }) => {
      saveSession(await api.register(input));
    },
    [saveSession],
  );

  const logout = useCallback(() => {
    clearStoredSession();
    setSession(null);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      user: session?.user ?? null,
      token: session?.access_token ?? null,
      isAuthenticated: Boolean(session?.access_token),
      login,
      register,
      logout,
    }),
    [login, logout, register, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return value;
}
