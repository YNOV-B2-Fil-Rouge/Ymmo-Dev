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

  // --- Favorites ---
  listFavorites: () => request("/favorites", { auth: true }),
  addFavorite: (id) => request(`/properties/${id}/favorites`, { method: "POST", auth: true }),
  removeFavorite: (id) => request(`/properties/${id}/favorites`, { method: "DELETE", auth: true }),

  // --- Messaging ---
  startConversation: (payload) => request("/conversations", { method: "POST", body: payload, auth: true }),
};

// ----- Python Data/AI service (separate base URL) -----
async function aiRequest(path, { method = "GET", body } = {}) {
  const res = await fetch(`${CONFIG.AI_BASE}${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  const isJson = res.headers.get("content-type")?.includes("application/json");
  const data = isJson ? await res.json() : null;
  if (!res.ok) throw { status: res.status, data };
  return data;
}

export const ai = {
  estimate: (payload) => aiRequest("/estimate", { method: "POST", body: payload }),
};
