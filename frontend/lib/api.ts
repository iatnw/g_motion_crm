const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };
  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (res.status === 401) {
    window.location.href = "/login";
    throw new Error("Unauthorized");
  }

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

// Auth
export const auth = {
  me: () => request<void>("/auth/me"),
  login: (password: string) =>
    request<{ totp_required: boolean }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
  verifyTOTP: (code: string) =>
    request<void>("/auth/totp/verify", {
      method: "POST",
      body: JSON.stringify({ code }),
    }),
  logout: () => request<void>("/auth/logout", { method: "POST" }),
  setPassword: (password: string, currentPassword = "") =>
    request<void>("/auth/password", {
      method: "POST",
      body: JSON.stringify({ password, current_password: currentPassword }),
    }),
  setupTOTP: () =>
    request<{ secret: string; key_uri: string }>("/auth/totp/setup", {
      method: "POST",
    }),
  enableTOTP: (secret: string, code: string) =>
    request<void>("/auth/totp/enable", {
      method: "POST",
      body: JSON.stringify({ secret, code }),
    }),
  disableTOTP: () =>
    request<void>("/auth/totp/disable", { method: "POST" }),
};

type PaginatedResponse<T> = { data: T[]; limit: number; offset: number; has_more: boolean };

// fetchAll pages a paginated endpoint until has_more is false.
async function fetchAll<T>(path: string, params: Record<string, string> = {}): Promise<T[]> {
  const results: T[] = [];
  let offset = 0;
  const limit = 200; // matches backend maxLimit
  while (true) {
    const qs = new URLSearchParams({ ...params, limit: String(limit), offset: String(offset) });
    const res = await request<PaginatedResponse<T>>(`${path}?${qs}`);
    results.push(...res.data);
    if (!res.has_more) break;
    offset += limit;
  }
  return results;
}

// Clients
export const clients = {
  list: () => fetchAll<import("./types").Client>("/clients"),
  get: (id: string) => request<import("./types").Client>(`/clients/${id}`),
  create: (data: object) =>
    request<import("./types").Client>("/clients", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: string, data: object) =>
    request<import("./types").Client>(`/clients/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
};

// Deals
export const deals = {
  nextNumber: () => request<{ deal_number: string }>("/deals/next-number").then((r) => r.deal_number),
  list: () => fetchAll<import("./types").Deal>("/deals"),
  listByClient: (clientId: string) => fetchAll<import("./types").Deal>("/deals", { client_id: clientId }),
  get: (id: string) => request<import("./types").Deal>(`/deals/${id}`),
  getShow: (id: string) => request<import("./types").DealShow>(`/deals/${id}/show`),
  create: (data: object) =>
    request<import("./types").Deal>("/deals", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: string, data: object) =>
    request<import("./types").Deal>(`/deals/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  updateStatus: (id: string, status: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/status`,
      { method: "PATCH", body: JSON.stringify({ status }) }
    ),
  updateBrokerFeeStatus: (id: string, status: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/broker-fee-status`,
      { method: "PATCH", body: JSON.stringify({ status }) }
    ),
  updateBrokerCommissionStatus: (id: string, status: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/broker-commission-status`,
      { method: "PATCH", body: JSON.stringify({ status }) }
    ),
  markPaymentReminderEmailSent: (id: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/payment-reminder-email-sent`,
      { method: "PATCH" }
    ),
};

// Referrals
export const referrals = {
  listByDeal: (dealId: string) =>
    request<import("./types").Referral[]>(`/deals/${dealId}/referrals`),
  create: (dealId: string, data: object) =>
    request<import("./types").Referral>(`/deals/${dealId}/referrals`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: string, data: object) =>
    request<import("./types").Referral>(`/referrals/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  updateStatus: (id: string, status: string) =>
    request<import("./types").Referral>(`/referrals/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    }),
  delete: (id: string) => request<void>(`/referrals/${id}`, { method: "DELETE" }),
};

// Dashboard
export const dashboard = {
  summary: (params: { from?: string; to?: string; limit?: number } = {}) => {
    const qs = new URLSearchParams();
    if (params.from) qs.set("from", params.from);
    if (params.to) qs.set("to", params.to);
    if (params.limit) qs.set("limit", String(params.limit));
    const query = qs.toString();
    return request<import("./types").DashboardSummary>(
      `/dashboard/summary${query ? `?${query}` : ""}`
    );
  },
};

// Brokers
export const brokers = {
  list: () => fetchAll<import("./types").Broker>("/brokers"),
  get: (id: string) => request<import("./types").Broker>(`/brokers/${id}`),
  create: (data: object) =>
    request<import("./types").Broker>("/brokers", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: string, data: object) =>
    request<import("./types").Broker>(`/brokers/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: string) => request<void>(`/brokers/${id}`, { method: "DELETE" }),
};

// Dealers
export const dealers = {
  list: () => fetchAll<import("./types").Dealer>("/dealers"),
  get: (id: string) => request<import("./types").Dealer>(`/dealers/${id}`),
  create: (data: object) =>
    request<import("./types").Dealer>("/dealers", {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (id: string, data: object) =>
    request<import("./types").Dealer>(`/dealers/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: string) => request<void>(`/dealers/${id}`, { method: "DELETE" }),
};

// Documents
export const documents = {
  listByClient: (clientId: string) =>
    request<import("./types").Document[]>(`/clients/${clientId}/documents`),
  listByDeal: (dealId: string) =>
    request<import("./types").Document[]>(`/deals/${dealId}/documents`),
  getURL: (id: string) =>
    request<{ url: string }>(`/documents/${id}/url`),
  delete: (id: string) =>
    request<void>(`/documents/${id}`, { method: "DELETE" }),
  uploadForClient: (clientId: string, file: File, documentType: string) => {
    const form = new FormData();
    form.append("file", file);
    form.append("document_type", documentType);
    return request<import("./types").Document>(
      `/clients/${clientId}/documents`,
      { method: "POST", body: form }
    );
  },
  uploadForDeal: (dealId: string, file: File, documentType: string) => {
    const form = new FormData();
    form.append("file", file);
    form.append("document_type", documentType);
    return request<import("./types").Document>(
      `/deals/${dealId}/documents`,
      { method: "POST", body: form }
    );
  },
};
