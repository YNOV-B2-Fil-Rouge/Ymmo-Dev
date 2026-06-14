// Thin API client around fetch. Centralizes the base URL, JSON handling,
// the JWT, and error shape so pages never call fetch directly (DRY).
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

// request performs an HTTP call and returns parsed JSON (or null for 204).
// On a non-2xx response it throws { status, data } so callers can branch.
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

export const api = {
  // --- Auth ---
  register: (payload) => request("/auth/register", { method: "POST", body: payload }),
  login: (payload) => request("/auth/login", { method: "POST", body: payload }),
  me: () => request("/auth/me", { auth: true }),

  // --- Properties ---
  listProperties: (queryString = "") => request(`/properties${queryString}`),
  getProperty: (id) => request(`/properties/${id}`),
  myProperties: () => request("/me/properties", { auth: true }),
  createProperty: (payload) => request("/properties", { method: "POST", body: payload, auth: true }),
  updateProperty: (id, payload) => request(`/properties/${id}`, { method: "PUT", body: payload, auth: true }),
  deleteProperty: (id) => request(`/properties/${id}`, { method: "DELETE", auth: true }),

  // --- Management (director / HQ) ---
  allProperties: () => request("/management/properties", { auth: true }),
  collaborators: () => request("/management/collaborators", { auth: true }),

  // --- Seller submissions awaiting validation (staff) ---
  pendingProperties: () => request("/management/pending-properties", { auth: true }),
  validateProperty: (id) => request(`/properties/${id}/validate`, { method: "POST", auth: true }),

  // --- IT administration ---
  allUsers: () => request("/management/users", { auth: true }),
  permissionsMatrix: () => request("/management/permissions", { auth: true }),

  // --- Photos (staff) ---
  addPhoto: (id, payload) => request(`/properties/${id}/photos`, { method: "POST", body: payload, auth: true }),
  deletePhoto: (id, photoId) => request(`/properties/${id}/photos/${photoId}`, { method: "DELETE", auth: true }),

  // --- Visits / Planning ---
  requestVisit: (id, payload) => request(`/properties/${id}/visits`, { method: "POST", body: payload, auth: true }),
  listVisits: () => request("/visits", { auth: true }),
  updateVisit: (id, status) => request(`/visits/${id}`, { method: "PATCH", body: { status }, auth: true }),

  // --- Meetings ---
  listMeetings: () => request("/meetings", { auth: true }),
  createMeeting: (payload) => request("/meetings", { method: "POST", body: payload, auth: true }),
  deleteMeeting: (id) => request(`/meetings/${id}`, { method: "DELETE", auth: true }),

  // --- Alerts (saved searches) ---
  listAlerts: () => request("/alerts", { auth: true }),
  createAlert: (payload) => request("/alerts", { method: "POST", body: payload, auth: true }),
  deleteAlert: (id) => request(`/alerts/${id}`, { method: "DELETE", auth: true }),

  // --- Sale files ---
  listSales: () => request("/sales", { auth: true }),
  getSale: (id) => request(`/sales/${id}`, { auth: true }),
  createSale: (payload) => request("/sales", { method: "POST", body: payload, auth: true }),
  updateSale: (id, payload) => request(`/sales/${id}`, { method: "PATCH", body: payload, auth: true }),

  // --- Favorites ---
  listFavorites: () => request("/favorites", { auth: true }),
  addFavorite: (id) => request(`/properties/${id}/favorites`, { method: "POST", auth: true }),
  removeFavorite: (id) => request(`/properties/${id}/favorites`, { method: "DELETE", auth: true }),

  // --- Messaging ---
  startConversation: (payload) => request("/conversations", { method: "POST", body: payload, auth: true }),
  listConversations: () => request("/conversations", { auth: true }),
  deleteConversation: (id) => request(`/conversations/${id}`, { method: "DELETE", auth: true }),
  getMessages: (id) => request(`/conversations/${id}/messages`, { auth: true }),
  sendMessage: (id, body) => request(`/conversations/${id}/messages`, { method: "POST", body: { body }, auth: true }),
  unreadCount: () => request("/messages/unread-count", { auth: true }),

  // --- Infra / monitoring ---
  ping: () => request("/ping"),
};

// ----- Data/AI: proxied through the Go API (/ai/...) -----
// The browser never talks to the Python service directly; the Go API relays to
// it on the internal Docker network.
export const ai = {
  estimate: (payload) => request("/ai/estimate", { method: "POST", body: payload }),
  dashboardKpis: () => request("/ai/dashboard/kpis", { auth: true }),
  trends: (city) => request(`/ai/trends${city ? `?city=${encodeURIComponent(city)}` : ""}`, { auth: true }),
  zones: () => request("/ai/zones", { auth: true }),
  popular: (limit = 6) => request(`/ai/popular?limit=${limit}`, { auth: true }),
};
