// Shared favorite helpers, reused by the catalogue cards and the detail page.
import { api } from "./api.js";
import { isLoggedIn } from "./auth.js";

// Set of the current user's favorited property ids (empty if logged out).
export async function favoriteIdSet() {
  if (!isLoggedIn()) return new Set();
  try {
    const { data } = await api.listFavorites();
    return new Set(data.map((p) => p.id));
  } catch {
    return new Set();
  }
}

// Heart icon, filled when favorited.
export function heartIcon(filled) {
  const path = "M12 21s-7-4.4-9.5-8.6C.9 9.6 2 6 5.2 6c1.9 0 3.1 1 3.8 2 .7-1 1.9-2 3.8-2 3.2 0 4.3 3.6 2.7 6.4C19 16.6 12 21 12 21z";
  return filled
    ? `<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="${path}"/></svg>`
    : `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="${path}"/></svg>`;
}

// Toggle a favorite. Returns the new state (true = now favorited).
export async function toggleFavorite(id, currentlyFavorited) {
  if (currentlyFavorited) {
    await api.removeFavorite(id);
    return false;
  }
  await api.addFavorite(id);
  return true;
}
