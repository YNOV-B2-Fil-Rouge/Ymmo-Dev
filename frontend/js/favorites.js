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

// Heart icon (symmetric, Feather-style), filled when favorited.
export function heartIcon(filled) {
  const path =
    "M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z";
  const common = `width="20" height="20" viewBox="0 0 24 24" aria-hidden="true" class="transition-transform duration-150"`;
  return filled
    ? `<svg ${common} fill="currentColor" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="${path}"/></svg>`
    : `<svg ${common} fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"><path d="${path}"/></svg>`;
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
