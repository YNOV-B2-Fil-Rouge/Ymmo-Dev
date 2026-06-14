// Dedicated search page: sidebar filters drive the catalogue query live.
import { api } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";
import { favoriteIdSet, heartIcon, toggleFavorite } from "./favorites.js";

document.getElementById("year").textContent = new Date().getFullYear();

// ----- Header account state -----
const account = document.getElementById("nav-account");
const user = currentUser();
if (user) {
  account.innerHTML = `
    <a href="./profile.html" title="Mon profil" aria-label="Mon profil"
       class="w-9 h-9 rounded-full border border-hibiscus text-hibiscus flex items-center justify-center hover:bg-hibiscus hover:text-white transition-colors">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-3.5 3.6-6 8-6s8 2.5 8 6"/></svg>
    </a>`;
} else {
  account.innerHTML = `
    <a href="./auth.html" class="text-sm font-medium text-midnight hover:text-hibiscus transition-colors">Connexion</a>
    <a href="./auth.html#register" class="text-sm font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-4 py-1.5 rounded-md transition-colors">Inscription</a>`;
}

// ----- Helpers -----
const escapeHtml = (v) => String(v ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
function primaryPhoto(p) {
  const ph = (p.photos || []).find((x) => x.is_primary) || (p.photos || [])[0];
  return ph ? ph.url : "https://placehold.co/600x400?text=Ymmo";
}
function card(p, isFav) {
  const badge = p.is_exclusive ? `<span class="absolute top-2 right-2 inline-flex items-center gap-1 bg-gold text-white text-xs font-semibold px-2.5 py-1 rounded-full shadow">★ Exclusivité</span>` : "";
  return `
    <a href="./property.html?id=${p.id}" class="group block rounded-xl border border-hibiscus/40 bg-white overflow-hidden transition-all duration-200 hover:shadow-xl hover:border-hibiscus hover:scale-[1.02]">
      <div class="relative overflow-hidden">
        <img src="${primaryPhoto(p)}" alt="${escapeHtml(p.title)}" loading="lazy" class="w-full h-44 object-cover group-hover:scale-105 transition-transform duration-300" />
        ${badge}
        <button class="fav-btn absolute top-2 left-2 w-9 h-9 rounded-full bg-white/90 text-hibiscus flex items-center justify-center shadow hover:bg-white transition-colors"
                data-id="${p.id}" aria-pressed="${isFav}" aria-label="${isFav ? "Retirer des favoris" : "Ajouter aux favoris"}">${heartIcon(isFav)}</button>
      </div>
      <div class="p-4">
        <h2 class="font-semibold">${escapeHtml(p.title)}${p.area != null ? ` - ${p.area} m²` : ""}</h2>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(p.city)} · ${escapeHtml(p.postal_code)}</p>
      </div>
    </a>`;
}

// ----- Filter state -----
const filters = {}; // sector, category_id, min_price, max_price, city, max_energy
const grid = document.getElementById("results-grid");
const empty = document.getElementById("results-empty");

async function runSearch() {
  const params = {};
  for (const [k, v] of Object.entries(filters)) {
    if (v !== undefined && v !== "" && v !== null) params[k] = v;
  }
  const qs = new URLSearchParams(params).toString();
  try {
    const [favSet, res] = await Promise.all([favoriteIdSet(), api.listProperties(qs ? `?${qs}` : "")]);
    grid.innerHTML = res.data.map((p) => card(p, favSet.has(p.id))).join("");
    empty.classList.toggle("hidden", res.data.length > 0);
  } catch {
    grid.innerHTML = `<p class="col-span-full text-center text-slate2 py-10">Impossible de charger les biens.</p>`;
  }
}

// Debounce for free-text / number inputs.
let timer;
const debouncedSearch = () => { clearTimeout(timer); timer = setTimeout(runSearch, 350); };

// ----- Type buttons (mutually exclusive) -----
const TYPE_FILTER = {
  residential: { sector: "RESIDENTIAL", category_id: undefined },
  office: { sector: undefined, category_id: 5 },
  land: { sector: undefined, category_id: 4 },
};
let activeType = null;
document.querySelectorAll(".type-btn").forEach((btn) => {
  btn.addEventListener("click", () => {
    const type = btn.dataset.type;
    if (activeType === type) {
      activeType = null;
      filters.sector = undefined;
      filters.category_id = undefined;
    } else {
      activeType = type;
      Object.assign(filters, TYPE_FILTER[type]);
    }
    document.querySelectorAll(".type-btn").forEach((b) => {
      const on = b.dataset.type === activeType;
      b.classList.toggle("bg-hibiscus", on);
      b.classList.toggle("text-white", on);
    });
    runSearch();
  });
});

// ----- Energy buttons (single select) -----
let activeEnergy = null;
document.querySelectorAll(".energy-btn").forEach((btn) => {
  btn.addEventListener("click", () => {
    const e = btn.dataset.energy;
    activeEnergy = activeEnergy === e ? null : e;
    filters.max_energy = activeEnergy || undefined;
    document.querySelectorAll(".energy-btn").forEach((b) => {
      const on = b.dataset.energy === activeEnergy;
      b.classList.toggle("bg-hibiscus", on);
      b.classList.toggle("text-white", on);
    });
    runSearch();
  });
});

// ----- Budget + city -----
document.getElementById("min-price").addEventListener("input", (e) => { filters.min_price = e.target.value; debouncedSearch(); });
document.getElementById("max-price").addEventListener("input", (e) => { filters.max_price = e.target.value; debouncedSearch(); });
document.getElementById("min-area").addEventListener("input", (e) => { filters.min_area = e.target.value; debouncedSearch(); });
document.getElementById("max-area").addEventListener("input", (e) => { filters.max_area = e.target.value; debouncedSearch(); });
const cityInput = document.getElementById("city");
cityInput.addEventListener("input", (e) => { filters.city = e.target.value.trim(); debouncedSearch(); });

// Prefill the city from the URL (?city=) when coming from the home search bar.
const initialCity = new URLSearchParams(window.location.search).get("city");
if (initialCity) { cityInput.value = initialCity; filters.city = initialCity; }

// ----- Favorite toggle (delegation) -----
grid.addEventListener("click", async (e) => {
  const btn = e.target.closest(".fav-btn");
  if (!btn) return;
  e.preventDefault();
  if (!isLoggedIn()) { window.location.href = "./auth.html"; return; }
  const id = Number(btn.dataset.id);
  const wasFav = btn.getAttribute("aria-pressed") === "true";
  btn.disabled = true;
  try {
    const now = await toggleFavorite(id, wasFav);
    btn.setAttribute("aria-pressed", String(now));
    btn.innerHTML = heartIcon(now);
  } catch { /* ignore */ }
  btn.disabled = false;
});

// ----- Create a saved-search alert from the active filters -----
const alertBtn = document.getElementById("create-alert");
const alertMsg = document.getElementById("alert-msg");
const showAlertMsg = (text, cls = "text-hibiscus") => {
  alertMsg.textContent = text;
  alertMsg.className = `mt-2 text-sm text-center ${cls}`;
};
alertBtn.addEventListener("click", async () => {
  if (!isLoggedIn()) { window.location.href = "./auth.html"; return; }

  // Map the search filters to the alert payload (no "sector" on alerts).
  const payload = {};
  if (filters.city) payload.city = filters.city;
  if (filters.category_id) payload.category_id = Number(filters.category_id);
  if (filters.min_price) payload.min_price = Number(filters.min_price);
  if (filters.max_price) payload.max_price = Number(filters.max_price);
  if (filters.min_area) payload.min_area = Number(filters.min_area);
  if (filters.max_energy) payload.max_energy = filters.max_energy;

  if (Object.keys(payload).length === 0) {
    showAlertMsg("Ajoutez au moins un critère avant de créer une alerte.");
    return;
  }

  alertBtn.disabled = true;
  try {
    await api.createAlert(payload);
    showAlertMsg("Alerte créée ! Retrouvez-la dans votre profil.", "text-green-700");
  } catch (err) {
    showAlertMsg(
      err.status === 409 ? "Vous avez déjà une alerte identique."
      : err.status === 400 ? "Critères invalides."
      : "Création impossible pour le moment."
    );
  }
  alertBtn.disabled = false;
});

runSearch();
