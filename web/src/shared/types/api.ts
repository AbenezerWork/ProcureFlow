export type ID = string;

export type User = {
  id: ID;
  email: string;
  full_name: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
};

export type Session = {
  access_token: string;
  token_type: "Bearer" | string;
  expires_at: string;
  user: User;
};

export type MembershipRole =
  | "owner"
  | "admin"
  | "requester"
  | "approver"
  | "procurement_officer"
  | "viewer";

export type MembershipStatus = "invited" | "active" | "suspended" | "removed";

export type Organization = {
  id: ID;
  name: string;
  slug: string;
  created_by_user_id: ID;
  created_at: string;
  updated_at: string;
  archived_at?: string;
};

export type UserOrganization = {
  organization: Organization;
  role: MembershipRole;
  status: MembershipStatus;
};

export type Membership = {
  id: ID;
  organization_id: ID;
  user_id: ID;
  role: MembershipRole;
  status: MembershipStatus;
  created_by_user_id: ID;
  invited_at: string;
  activated_at?: string;
  suspended_at?: string;
  removed_at?: string;
  created_at: string;
  updated_at: string;
};

export type OrganizationMember = {
  user: User;
  membership: Membership;
};

export type ProcurementRequestStatus =
  | "draft"
  | "submitted"
  | "approved"
  | "rejected"
  | "cancelled";

export type ProcurementRequest = {
  id: ID;
  organization_id: ID;
  requester_user_id: ID;
  title: string;
  description?: string | null;
  justification?: string | null;
  status: ProcurementRequestStatus;
  currency_code: string;
  estimated_total_amount?: string | null;
  decision_comment?: string | null;
  created_at: string;
  updated_at: string;
};

export type RFQStatus = "draft" | "published" | "closed" | "evaluated" | "awarded" | "cancelled";

export type RFQ = {
  id: ID;
  organization_id: ID;
  procurement_request_id: ID;
  reference_number?: string | null;
  title: string;
  description?: string | null;
  status: RFQStatus;
  created_by_user_id: ID;
  created_at: string;
  updated_at: string;
};

export type Vendor = {
  id: ID;
  organization_id: ID;
  name: string;
  legal_name?: string | null;
  contact_name?: string | null;
  email?: string | null;
  phone?: string | null;
  status: "active" | "archived";
  created_at: string;
  updated_at: string;
};

export type Quotation = {
  id: ID;
  organization_id: ID;
  rfq_id: ID;
  rfq_vendor_id: ID;
  status: "draft" | "submitted" | "rejected";
  currency_code: string;
  lead_time_days?: number | null;
  payment_terms?: string | null;
  notes?: string | null;
  vendor_name?: string | null;
  created_at: string;
  updated_at: string;
};

export type ActivityLog = {
  id: ID;
  organization_id: ID;
  actor_user_id?: ID;
  entity_type: string;
  entity_id: ID;
  action: string;
  summary?: string;
  metadata: Record<string, unknown>;
  occurred_at: string;
};

export type ApiList<T, K extends string> = Record<K, T[]>;
