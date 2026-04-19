import type {
  ActivityLog,
  ApiList,
  ID,
  OrganizationMember,
  Organization,
  ProcurementRequest,
  ProcurementRequestStatus,
  Quotation,
  RFQ,
  RFQStatus,
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

  listMemberships(token: string, tenantId: ID) {
    return request<ApiList<OrganizationMember, "memberships">>(
      `/api/v1/organizations/${tenantId}/memberships`,
      { token, tenantId },
    );
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
