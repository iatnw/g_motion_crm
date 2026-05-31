const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

type RequestOptions = RequestInit & { skipAuthRedirect?: boolean };

async function request<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { skipAuthRedirect, ...fetchOptions } = options;
  const headers: Record<string, string> = {
    ...(fetchOptions.headers as Record<string, string>),
  };
  if (!(fetchOptions.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...fetchOptions,
    headers,
    credentials: "include",
    cache: "no-store",
  });

  if (res.status === 401 && !skipAuthRedirect) {
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
  setupRequired: () => request<{ required: boolean }>("/auth/setup-required"),
  setup: (password: string) =>
    request<void>("/auth/setup", {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
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
      skipAuthRedirect: true,
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
type DealsPaginatedResponse = PaginatedResponse<import("./types").Deal> & {
  counts?: { all: number; new: number; pending: number; paid: number; cancelled: number };
};

async function presignedUpload(
  file: File,
  documentType: string,
  onProgress: ((pct: number) => void) | undefined,
  ids: { client_id?: string; deal_id?: string },
): Promise<import("./types").Document> {
  const qs = new URLSearchParams({
    filename: file.name,
    content_type: file.type || "application/octet-stream",
    size: String(file.size),
    document_type: documentType,
  });
  if (ids.client_id) qs.set("client_id", ids.client_id);
  if (ids.deal_id) qs.set("deal_id", ids.deal_id);
  const { upload_url, upload_token } = await request<{ upload_url: string; upload_token: string }>(
    `/documents/presign?${qs}`
  );
  await new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", upload_url);
    xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");
    if (onProgress) xhr.upload.onprogress = (e) => { if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100)); };
    xhr.onload = () => (xhr.status >= 200 && xhr.status < 300 ? resolve() : reject(new Error(`S3 upload failed: ${xhr.status}`)));
    xhr.onerror = () => reject(new Error("S3 upload network error"));
    xhr.send(file);
  });
  return request<import("./types").Document>("/documents/confirm", {
    method: "POST",
    body: JSON.stringify({ upload_token }),
  });
}

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
  list: (params: Record<string, string> = {}) => fetchAll<import("./types").Deal>("/deals", params),
  listPaged: (params: { limit: number; offset: number; status?: string; search?: string; vin?: "present" | "missing" }) => {
    const qs = new URLSearchParams({ limit: String(params.limit), offset: String(params.offset) });
    if (params.status) qs.set("status", params.status);
    if (params.search) qs.set("search", params.search);
    if (params.vin) qs.set("vin", params.vin);
    return request<DealsPaginatedResponse>(`/deals?${qs}`);
  },
  listByClient: (clientId: string) => fetchAll<import("./types").Deal>("/deals", { client_id: clientId }),
  listByBrokerPaged: (brokerId: string, limit: number, offset: number) => {
    const qs = new URLSearchParams({ broker_id: brokerId, limit: String(limit), offset: String(offset) });
    return request<PaginatedResponse<import("./types").Deal>>(`/deals?${qs}`);
  },
  listByDealerPaged: (dealerId: string, limit: number, offset: number) => {
    const qs = new URLSearchParams({ dealer_id: dealerId, limit: String(limit), offset: String(offset) });
    return request<PaginatedResponse<import("./types").Deal>>(`/deals?${qs}`);
  },
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
  updateBrokerCommissionStatus: (id: string, status: string, payment_method?: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/broker-commission-status`,
      { method: "PATCH", body: JSON.stringify({ status, payment_method }) }
    ),
  markPaymentReminderEmailSent: (id: string) =>
    request<import("./types").Deal>(
      `/deals/${id}/payment-reminder-email-sent`,
      { method: "PATCH" }
    ),
};

// Referrals
export const referrals = {
  list: () => request<import("./types").Referral[]>("/referrals"),
  listByDeals: (dealIds: string[]) => {
    if (dealIds.length === 0) return Promise.resolve([] as import("./types").Referral[]);
    const qs = new URLSearchParams({ deal_ids: dealIds.join(",") });
    return request<import("./types").Referral[]>(`/referrals?${qs}`);
  },
  listByDeal: (dealId: string) =>
    request<import("./types").Referral[]>(`/deals/${dealId}/referrals`),
  create: (data: object) =>
    request<import("./types").Referral>(`/referrals`, {
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
  updateAmount: (id: string, amount?: number) =>
    request<import("./types").Referral>(`/referrals/${id}/amount`, {
      method: "PATCH",
      body: JSON.stringify({ amount }),
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
  uploadForClient: (clientId: string, file: File, documentType: string, onProgress?: (pct: number) => void) =>
    presignedUpload(file, documentType, onProgress, { client_id: clientId }),
  uploadForDeal: (dealId: string, file: File, documentType: string, onProgress?: (pct: number) => void) =>
    presignedUpload(file, documentType, onProgress, { deal_id: dealId }),
};
