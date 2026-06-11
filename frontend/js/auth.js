// Authentication helpers built on top of the API client. Pages use these to
// log in/out and read the current user without touching localStorage directly.
import { api, setToken, clearToken } from "./api.js";

const USER_KEY = "ymmo_user";

export async function login(email, password) {
  const data = await api.login({ email, password });
  setToken(data.token);
  localStorage.setItem(USER_KEY, JSON.stringify(data.user));
  return data.user;
}

export function logout() {
  clearToken();
  localStorage.removeItem(USER_KEY);
}

export function currentUser() {
  const raw = localStorage.getItem(USER_KEY);
  return raw ? JSON.parse(raw) : null;
}

export function isLoggedIn() {
  return Boolean(getTokenSafe());
}

function getTokenSafe() {
  return localStorage.getItem("ymmo_token");
}
