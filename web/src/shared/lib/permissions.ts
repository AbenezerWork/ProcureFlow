import type { MembershipRole } from "@/shared/types/api";

export function canAccessApprovals(role?: MembershipRole | null) {
  return role === "owner" || role === "admin" || role === "approver";
}
