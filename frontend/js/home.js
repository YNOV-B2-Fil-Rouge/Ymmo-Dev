// Home page: nav state, catalogue loading from the API, and search.
import { api } from "./api.js";
import { currentUser, logout, isLoggedIn } from "./auth.js";
import { favoriteIdSet, heartIcon, toggleFavorite } from "./favorites.js";

// ---------- Footer year ----------
document.getElementById("year").textContent = new Date().getFullYear();

// ---------- Auth state in the nav ----------
const account = document.getElementById("nav-account");
const user = currentUser();
if (user) {
  // Roles allowed to publish a listing.
  const canPublish = ["AGENT", "DIRECTOR", "HQ", "SELLER"].includes(user.role);
  const isStaff = ["AGENT", "DIRECTOR", "HQ"].includes(user.role);
  account.innerHTML = `
    ${isStaff ? `<a href="./dashboard.html" class="text-sm font-medium text-midnight hover:text-hibiscus transition-colors">Dashboard</a>` : ""}
    ${canPublish ? `<a href="./property-new.html" class="text-sm font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-4 py-1.5 rounded-md transition-colors">Publier une annonce</a>` : ""}
    <a href="./profile.html" title="Mon profil" aria-label="Mon profil (${escapeHtml(user.first_name)})"
       class="w-9 h-9 rounded-full border border-hibiscus text-hibiscus flex items-center justify-center hover:bg-hibiscus hover:text-white transition-colors">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-3.5 3.6-6 8-6s8 2.5 8 6"/></svg>
    </a>`;
}

// ---------- Catalogue ----------
const grid = document.getElementById("catalogue");
const empty = document.getElementById("empty");

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (c) => (
    { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
  ));
}

function primaryPhoto(property) {
  const photos = property.photos || [];
  const main = photos.find((p) => p.is_primary) || photos[0];
  return main ? main.url : "https://placehold.co/600x400?text=Ymmo";
}

function card(property, isFav = false) {
  const title = escapeHtml(property.title);
  const area = property.area != null ? `${property.area} m²` : "";
  const badge = property.is_exclusive
    ? `<span class="absolute top-2 right-2 inline-flex items-center gap-1 bg-gold text-white text-xs font-semibold px-2.5 py-1 rounded-full shadow">★ Exclusivité</span>`
    : "";

  return `
    <a href="./property.html?id=${property.id}"
       class="group block rounded-xl border border-hibiscus/40 bg-white overflow-hidden transition-all duration-200 hover:shadow-xl hover:border-hibiscus hover:scale-[1.03]">
      <div class="relative overflow-hidden">
        <img src="${primaryPhoto(property)}" alt="${title}" loading="lazy"
             class="w-full h-44 object-cover group-hover:scale-105 transition-transform duration-300" />
        ${badge}
        <button class="fav-btn absolute top-2 left-2 w-9 h-9 rounded-full bg-white/90 text-hibiscus flex items-center justify-center shadow hover:bg-white transition-colors"
                data-id="${property.id}" aria-pressed="${isFav}" aria-label="${isFav ? "Retirer des favoris" : "Ajouter aux favoris"}">
          ${heartIcon(isFav)}
        </button>
      </div>
      <div class="p-4">
        <h2 class="font-semibold text-midnight">${title}${area ? ` - ${area}` : ""}</h2>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(property.city)} · ${escapeHtml(property.postal_code)}</p>
      </div>
    </a>`;
}

function skeleton() {
  return Array.from({ length: 6 })
    .map(() => `<div class="rounded-xl border border-hibiscus/20 bg-white overflow-hidden animate-pulse">
        <div class="h-44 bg-slate2/20"></div>
        <div class="p-4 space-y-2"><div class="h-4 bg-slate2/20 rounded w-3/4"></div><div class="h-3 bg-slate2/20 rounded w-1/2"></div></div>
      </div>`)
    .join("");
}

async function loadCatalogue(params = {}) {
  const qs = new URLSearchParams(params).toString();
  grid.innerHTML = skeleton();
  empty.classList.add("hidden");

  try {
    const [favSet, res] = await Promise.all([
      favoriteIdSet(),
      api.listProperties(qs ? `?${qs}` : ""),
    ]);
    grid.innerHTML = res.data.map((p) => card(p, favSet.has(p.id))).join("");
    empty.classList.toggle("hidden", res.data.length > 0);
  } catch (err) {
    grid.innerHTML = `<p class="col-span-full text-center text-slate2 py-10">Impossible de charger les biens. L'API est-elle démarrée ?</p>`;
  }
}

// Favorite toggle via event delegation (the grid element persists across loads).
grid.addEventListener("click", async (e) => {
  const btn = e.target.closest(".fav-btn");
  if (!btn) return;
  e.preventDefault();
  if (!isLoggedIn()) {
    window.location.href = "./auth.html";
    return;
  }
  const id = Number(btn.dataset.id);
  const wasFav = btn.getAttribute("aria-pressed") === "true";
  btn.disabled = true;
  try {
    const now = await toggleFavorite(id, wasFav);
    btn.setAttribute("aria-pressed", String(now));
    btn.setAttribute("aria-label", now ? "Retirer des favoris" : "Ajouter aux favoris");
    btn.innerHTML = heartIcon(now);
  } catch {
    /* keep previous state on error */
  }
  btn.disabled = false;
});

// ---------- Search ----------
// The API filters on an exact city for now; a fuzzy text search (city/postal/
// region) is a small backend enhancement we can add later.
const searchForm = document.getElementById("search-form");
searchForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const q = searchForm.q.value.trim();
  loadCatalogue(q ? { city: q } : {});
});

loadCatalogue();
