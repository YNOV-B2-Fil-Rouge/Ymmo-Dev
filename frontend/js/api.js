// API client: wraps fetch with the base URL, JWT and JSON/error handling.
import { CONFIG } from "./config.js";

const TOKEN_KEY = "ymmo_token";

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}
export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

// HTTP call returning parsed JSON; throws { status, data } on a non-2xx response.
async function request(path, { method = "GET", body, auth = false } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (auth) {
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${CONFIG.API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const isJson = res.headers.get("content-type")?.includes("application/json");
  const data = isJson ? await res.json() : null;

  if (!res.ok) {
    throw { status: res.status, data };
  }
  return data;
}

// Multipart POST for file uploads (Content-Type is set by the browser).
async function requestForm(path, formData) {
  const headers = {};
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${CONFIG.API_BASE}${path}`, { method: "POST", headers, body: formData });
  const isJson = res.headers.get("content-type")?.includes("application/json");
  const data = isJson ? await res.json() : null;
  if (!res.ok) throw { status: res.status, data };
  return data;
}

export const api = {
  register: (payload) => request("/auth/register", { method: "POST", body: payload }),
  login: (payload) => request("/auth/login", { method: "POST", body: payload }),
  me: () => request("/auth/me", { auth: true }),

  listProperties: (queryString = "") => request(`/properties${queryString}`),
  getProperty: (id) => request(`/properties/${id}`),
  myProperties: () => request("/me/properties", { auth: true }),
  createProperty: (payload) => request("/properties", { method: "POST", body: payload, auth: true }),
  updateProperty: (id, payload) => request(`/properties/${id}`, { method: "PUT", body: payload, auth: true }),
  deleteProperty: (id) => request(`/properties/${id}`, { method: "DELETE", auth: true }),

  allProperties: () => request("/management/properties", { auth: true }),
  collaborators: () => request("/management/collaborators", { auth: true }),

  pendingProperties: () => request("/management/pending-properties", { auth: true }),
  validateProperty: (id) => request(`/properties/${id}/validate`, { method: "POST", auth: true }),

  applyAsSeller: (payload) => request("/seller-applications", { method: "POST", body: payload, auth: true }),
  mySellerApplication: () => request("/seller-applications", { auth: true }),
  pendingSellerApplications: () => request("/management/seller-applications", { auth: true }),
  approveSellerApplication: (id) => request(`/management/seller-applications/${id}/approve`, { method: "POST", auth: true }),
  rejectSellerApplication: (id) => request(`/management/seller-applications/${id}/reject`, { method: "POST", auth: true }),

  allUsers: () => request("/management/users", { auth: true }),
  permissionsMatrix: () => request("/management/permissions", { auth: true }),

  addPhoto: (id, payload) => request(`/properties/${id}/photos`, { method: "POST", body: payload, auth: true }),
  uploadPhoto: (id, formData) => requestForm(`/properties/${id}/photos/upload`, formData),
  deletePhoto: (id, photoId) => request(`/properties/${id}/photos/${photoId}`, { method: "DELETE", auth: true }),

  requestVisit: (id, payload) => request(`/properties/${id}/visits`, { method: "POST", body: payload, auth: true }),
  listVisits: () => request("/visits", { auth: true }),
  updateVisit: (id, status) => request(`/visits/${id}`, { method: "PATCH", body: { status }, auth: true }),

  listMeetings: () => request("/meetings", { auth: true }),
  createMeeting: (payload) => request("/meetings", { method: "POST", body: payload, auth: true }),
  deleteMeeting: (id) => request(`/meetings/${id}`, { method: "DELETE", auth: true }),

  listAlerts: () => request("/alerts", { auth: true }),
  createAlert: (payload) => request("/alerts", { method: "POST", body: payload, auth: true }),
  deleteAlert: (id) => request(`/alerts/${id}`, { method: "DELETE", auth: true }),

  listSales: () => request("/sales", { auth: true }),
  getSale: (id) => request(`/sales/${id}`, { auth: true }),
  createSale: (payload) => request("/sales", { method: "POST", body: payload, auth: true }),
  updateSale: (id, payload) => request(`/sales/${id}`, { method: "PATCH", body: payload, auth: true }),

  listFavorites: () => request("/favorites", { auth: true }),
  addFavorite: (id) => request(`/properties/${id}/favorites`, { method: "POST", auth: true }),
  removeFavorite: (id) => request(`/properties/${id}/favorites`, { method: "DELETE", auth: true }),

  startConversation: (payload) => request("/conversations", { method: "POST", body: payload, auth: true }),
  listConversations: () => request("/conversations", { auth: true }),
  deleteConversation: (id) => request(`/conversations/${id}`, { method: "DELETE", auth: true }),
  getMessages: (id) => request(`/conversations/${id}/messages`, { auth: true }),
  sendMessage: (id, body) => request(`/conversations/${id}/messages`, { method: "POST", body: { body }, auth: true }),
  unreadCount: () => request("/messages/unread-count", { auth: true }),

  ping: () => request("/ping"),
};

// AI endpoints, proxied by the Go API (the browser never hits the Python service).
export const ai = {
  estimate: (payload) => request("/ai/estimate", { method: "POST", body: payload }),
  predictDelay: (payload) => request("/ai/predict-delay", { method: "POST", body: payload }),
  dashboardKpis: () => request("/ai/dashboard/kpis", { auth: true }),
  trends: (city) => request(`/ai/trends${city ? `?city=${encodeURIComponent(city)}` : ""}`, { auth: true }),
  zones: () => request("/ai/zones", { auth: true }),
  popular: (limit = 6) => request(`/ai/popular?limit=${limit}`, { auth: true }),
};
