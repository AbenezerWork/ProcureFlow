import type { Session } from "@/shared/types/api";

const storageKey = "procureflow.session";

export function loadStoredSession(): Session | null {
  try {
    const raw = window.localStorage.getItem(storageKey);
    return raw ? (JSON.parse(raw) as Session) : null;
  } catch {
    return null;
  }
}

export function storeSession(session: Session) {
  window.localStorage.setItem(storageKey, JSON.stringify(session));
}

export function clearStoredSession() {
  window.localStorage.removeItem(storageKey);
}
