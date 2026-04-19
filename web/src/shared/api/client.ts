import type {
  ActivityLog,
  Award,
  ApiList,
  Health,
  ID,
  OrganizationMember,
  Organization,
  ProcurementItem,
  ProcurementRequest,
  ProcurementRequestStatus,
  Quotation,
  QuotationItem,
  RFQ,
  RFQItem,
  RFQStatus,
  RFQQuotationComparison,
  RFQVendor,
  Session,
  User,
  UserOrganization,
  Vendor,
} from "@/shared/types/api";

type QueryValue = string | number | boolean | null | undefined;

type ApiRequestOptions = {
  method?: "GET" | "POST" | "PATCH" | "DELETE";
  body?: unknown;
  query?: Record<string, QueryValue>;
  token?: string | null;
  tenantId?: string | null;
};

export class ApiError extends Error {
  readonly status: number;
  readonly payload: unknown;

  constructor(status: number, message: string, payload: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.payload = payload;
  }
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "";

function buildUrl(path: string, query?: Record<string, QueryValue>) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  const base = apiBaseUrl || window.location.origin;
  const url = new URL(normalizedPath, base);

  Object.entries(query ?? {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      url.searchParams.set(key, String(value));
    }
  });

  if (!apiBaseUrl) {
    return `${url.pathname}${url.search}`;
  }

  return url.toString();
}

async function request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const headers = new Headers({
    Accept: "application/json",
  });

  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  if (options.token) {
    headers.set("Authorization", `Bearer ${options.token}`);
  }

  if (options.tenantId) {
    headers.set("X-Tenant-ID", options.tenantId);
  }

  const response = await fetch(buildUrl(path, options.query), {
    method: options.method ?? "GET",
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  const contentType = response.headers.get("content-type") ?? "";
  const payload = contentType.includes("application/json") ? await response.json() : null;

  if (!response.ok) {
    const message =
      payload && typeof payload === "object" && "error" in payload
        ? String((payload as { error: unknown }).error)
        : `Request failed with status ${response.status}`;

    throw new ApiError(response.status, message, payload);
  }

  return payload as T;
}

export const api = {
  health(tenantId?: string | null) {
    return request<Health>("/healthz", { tenantId });
  },

  login(input: { email: string; password: string }) {
    return request<Session>("/api/v1/auth/login", {
      method: "POST",
      body: input,
    });
  },

  register(input: { email: string; password: string; full_name: string }) {
    return request<Session>("/api/v1/auth/register", {
      method: "POST",
      body: input,
    });
  },

  currentUser(token: string) {
    return request<User>("/api/v1/auth/me", { token });
  },

  listOrganizations(token: string) {
    return request<ApiList<UserOrganization, "organizations">>("/api/v1/organizations/", { token });
  },

  createOrganization(token: string, input: { name: string; slug?: string }) {
    return request<{ organization: Organization; role: string; status: string }>(
      "/api/v1/organizations/",
      {
        method: "POST",
        token,
        body: input,
      },
    );
  },

  getOrganization(token: string, tenantId: ID) {
    return request<{ organization: Organization; membership: unknown }>(
      `/api/v1/organizations/${tenantId}`,
      { token, tenantId },
    );
  },

  updateOrganization(token: string, tenantId: ID, input: { name?: string; slug?: string }) {
    return request<{ organization: Organization; membership: unknown }>(
      `/api/v1/organizations/${tenantId}`,
      {
        method: "PATCH",
        token,
        tenantId,
        body: input,
      },
    );
  },

  transferOrganizationOwnership(
    token: string,
    tenantId: ID,
    input: { target_user_id: ID; current_owner_new_role?: string },
  ) {
    return request<unknown>(`/api/v1/organizations/${tenantId}/ownership-transfer`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  listMemberships(token: string, tenantId: ID) {
    return request<ApiList<OrganizationMember, "memberships">>(
      `/api/v1/organizations/${tenantId}/memberships`,
      { token, tenantId },
    );
  },

  createMembership(
    token: string,
    tenantId: ID,
    input: { user_id?: ID; email?: string; role: string; status?: string },
  ) {
    return request<OrganizationMember>(`/api/v1/organizations/${tenantId}/memberships`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  updateMembership(
    token: string,
    tenantId: ID,
    userId: ID,
    input: { role?: string; status?: string },
  ) {
    return request<OrganizationMember>(
      `/api/v1/organizations/${tenantId}/memberships/${userId}`,
      {
        method: "PATCH",
        token,
        tenantId,
        body: input,
      },
    );
  },

  createVendor(token: string, tenantId: ID, input: Record<string, string>) {
    return request<Vendor>(`/api/v1/organizations/${tenantId}/vendors`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  getVendor(token: string, tenantId: ID, vendorId: ID) {
    return request<Vendor>(`/api/v1/organizations/${tenantId}/vendors/${vendorId}`, {
      token,
      tenantId,
    });
  },

  updateVendor(token: string, tenantId: ID, vendorId: ID, input: Record<string, string>) {
    return request<Vendor>(`/api/v1/organizations/${tenantId}/vendors/${vendorId}`, {
      method: "PATCH",
      token,
      tenantId,
      body: input,
    });
  },

  archiveVendor(token: string, tenantId: ID, vendorId: ID) {
    return request<Vendor>(`/api/v1/organizations/${tenantId}/vendors/${vendorId}/archive`, {
      method: "POST",
      token,
      tenantId,
    });
  },

  listProcurementRequests(
    token: string,
    tenantId: ID,
    status?: ProcurementRequestStatus,
  ) {
    return request<ApiList<ProcurementRequest, "procurement_requests">>(
      `/api/v1/organizations/${tenantId}/procurement-requests`,
      { token, tenantId, query: { status } },
    );
  },

  createProcurementRequest(
    token: string,
    tenantId: ID,
    input: {
      title: string;
      description?: string;
      justification?: string;
      currency_code?: string;
      estimated_total_amount?: string;
    },
  ) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests`,
      {
        method: "POST",
        token,
        tenantId,
        body: input,
      },
    );
  },

  getProcurementRequest(token: string, tenantId: ID, requestId: ID) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}`,
      { token, tenantId },
    );
  },

  updateProcurementRequest(
    token: string,
    tenantId: ID,
    requestId: ID,
    input: Partial<Pick<ProcurementRequest, "title" | "description" | "justification" | "currency_code" | "estimated_total_amount">>,
  ) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}`,
      {
        method: "PATCH",
        token,
        tenantId,
        body: input,
      },
    );
  },

  submitProcurementRequest(token: string, tenantId: ID, requestId: ID) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/submit`,
      { method: "POST", token, tenantId },
    );
  },

  approveProcurementRequest(token: string, tenantId: ID, requestId: ID, decision_comment?: string) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/approve`,
      { method: "POST", token, tenantId, body: { decision_comment } },
    );
  },

  rejectProcurementRequest(token: string, tenantId: ID, requestId: ID, decision_comment?: string) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/reject`,
      { method: "POST", token, tenantId, body: { decision_comment } },
    );
  },

  cancelProcurementRequest(token: string, tenantId: ID, requestId: ID) {
    return request<ProcurementRequest>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/cancel`,
      { method: "POST", token, tenantId },
    );
  },

  listProcurementRequestItems(token: string, tenantId: ID, requestId: ID) {
    return request<ApiList<ProcurementItem, "items">>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/items`,
      { token, tenantId },
    );
  },

  createProcurementRequestItem(
    token: string,
    tenantId: ID,
    requestId: ID,
    input: {
      item_name: string;
      description?: string;
      quantity: string;
      unit: string;
      estimated_unit_price?: string;
      needed_by_date?: string;
    },
  ) {
    return request<ProcurementItem>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/items`,
      { method: "POST", token, tenantId, body: input },
    );
  },

  updateProcurementRequestItem(
    token: string,
    tenantId: ID,
    requestId: ID,
    itemId: ID,
    input: Partial<Pick<ProcurementItem, "item_name" | "description" | "quantity" | "unit" | "estimated_unit_price" | "needed_by_date">>,
  ) {
    return request<ProcurementItem>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/items/${itemId}`,
      { method: "PATCH", token, tenantId, body: input },
    );
  },

  deleteProcurementRequestItem(token: string, tenantId: ID, requestId: ID, itemId: ID) {
    return request<null>(
      `/api/v1/organizations/${tenantId}/procurement-requests/${requestId}/items/${itemId}`,
      { method: "DELETE", token, tenantId },
    );
  },

  listApprovalInbox(token: string, tenantId: ID) {
    return request<ApiList<ProcurementRequest, "procurement_requests">>(
      `/api/v1/organizations/${tenantId}/approvals/procurement-requests`,
      { token, tenantId },
    );
  },

  listRFQs(token: string, tenantId: ID, status?: RFQStatus) {
    return request<ApiList<RFQ, "rfqs">>(`/api/v1/organizations/${tenantId}/rfqs`, {
      token,
      tenantId,
      query: { status },
    });
  },

  createRFQ(
    token: string,
    tenantId: ID,
    input: { procurement_request_id: ID; reference_number?: string; title?: string; description?: string },
  ) {
    return request<RFQ>(`/api/v1/organizations/${tenantId}/rfqs`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  getRFQ(token: string, tenantId: ID, rfqId: ID) {
    return request<RFQ>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}`, {
      token,
      tenantId,
    });
  },

  updateRFQ(token: string, tenantId: ID, rfqId: ID, input: { reference_number?: string; title?: string; description?: string }) {
    return request<RFQ>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}`, {
      method: "PATCH",
      token,
      tenantId,
      body: input,
    });
  },

  transitionRFQ(token: string, tenantId: ID, rfqId: ID, action: "publish" | "close" | "evaluate" | "cancel") {
    return request<RFQ>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/${action}`, {
      method: "POST",
      token,
      tenantId,
    });
  },

  listRFQItems(token: string, tenantId: ID, rfqId: ID) {
    return request<ApiList<RFQItem, "items">>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/items`, {
      token,
      tenantId,
    });
  },

  listRFQVendors(token: string, tenantId: ID, rfqId: ID) {
    return request<ApiList<RFQVendor, "vendors">>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/vendors`, {
      token,
      tenantId,
    });
  },

  attachRFQVendor(token: string, tenantId: ID, rfqId: ID, vendor_id: ID) {
    return request<RFQVendor>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/vendors`, {
      method: "POST",
      token,
      tenantId,
      body: { vendor_id },
    });
  },

  removeRFQVendor(token: string, tenantId: ID, rfqId: ID, vendorId: ID) {
    return request<null>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/vendors/${vendorId}`, {
      method: "DELETE",
      token,
      tenantId,
    });
  },

  compareRFQQuotations(token: string, tenantId: ID, rfqId: ID) {
    return request<RFQQuotationComparison>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/comparison`,
      { token, tenantId },
    );
  },

  listVendors(token: string, tenantId: ID) {
    return request<ApiList<Vendor, "vendors">>(`/api/v1/organizations/${tenantId}/vendors`, {
      token,
      tenantId,
    });
  },

  listQuotations(token: string, tenantId: ID, rfqId: ID) {
    return request<ApiList<Quotation, "quotations">>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations`,
      { token, tenantId },
    );
  },

  createQuotation(
    token: string,
    tenantId: ID,
    rfqId: ID,
    input: { rfq_vendor_id: ID; currency_code?: string; lead_time_days?: number; payment_terms?: string; notes?: string },
  ) {
    return request<Quotation>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  getQuotation(token: string, tenantId: ID, rfqId: ID, quotationId: ID) {
    return request<Quotation>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}`,
      { token, tenantId },
    );
  },

  updateQuotation(
    token: string,
    tenantId: ID,
    rfqId: ID,
    quotationId: ID,
    input: { currency_code?: string; lead_time_days?: number; payment_terms?: string; notes?: string },
  ) {
    return request<Quotation>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}`,
      { method: "PATCH", token, tenantId, body: input },
    );
  },

  submitQuotation(token: string, tenantId: ID, rfqId: ID, quotationId: ID) {
    return request<Quotation>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}/submit`,
      { method: "POST", token, tenantId },
    );
  },

  rejectQuotation(token: string, tenantId: ID, rfqId: ID, quotationId: ID, rejection_reason?: string) {
    return request<Quotation>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}/reject`,
      { method: "POST", token, tenantId, body: { rejection_reason } },
    );
  },

  listQuotationItems(token: string, tenantId: ID, rfqId: ID, quotationId: ID) {
    return request<ApiList<QuotationItem, "items">>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}/items`,
      { token, tenantId },
    );
  },

  updateQuotationItem(
    token: string,
    tenantId: ID,
    rfqId: ID,
    quotationId: ID,
    itemId: ID,
    input: { unit_price?: string; delivery_days?: number; notes?: string },
  ) {
    return request<QuotationItem>(
      `/api/v1/organizations/${tenantId}/rfqs/${rfqId}/quotations/${quotationId}/items/${itemId}`,
      { method: "PATCH", token, tenantId, body: input },
    );
  },

  getAward(token: string, tenantId: ID, rfqId: ID) {
    return request<Award>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/award`, {
      token,
      tenantId,
    });
  },

  createAward(token: string, tenantId: ID, rfqId: ID, input: { quotation_id: ID; reason: string }) {
    return request<Award>(`/api/v1/organizations/${tenantId}/rfqs/${rfqId}/award`, {
      method: "POST",
      token,
      tenantId,
      body: input,
    });
  },

  listActivityLogs(token: string, tenantId: ID, entityType: string, entityId: ID) {
    return request<ApiList<ActivityLog, "activity_logs">>(
      `/api/v1/organizations/${tenantId}/activity-logs`,
      {
        token,
        tenantId,
        query: {
          entity_type: entityType,
          entity_id: entityId,
        },
      },
    );
  },
};
