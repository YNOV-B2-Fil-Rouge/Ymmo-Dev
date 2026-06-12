// Profile page: sidebar tabs + the user's favorites and account settings.
import { api } from "./api.js";
import { currentUser, logout, isLoggedIn } from "./auth.js";

// Must be logged in.
if (!isLoggedIn()) {
  window.location.href = "./auth.html";
}

const user = currentUser() || {};

// ---------- Identity (avatar + name) ----------
const initials = ((user.first_name?.[0] || "") + (user.last_name?.[0] || "")).toUpperCase() || "?";
document.getElementById("avatar").textContent = initials;
document.getElementById("full-name").textContent =
  `${user.first_name || ""} ${(user.last_name || "").toUpperCase()}`.trim();

// ---------- Settings panel ----------
const roleLabels = {
  BUYER: "Acheteur", SELLER: "Vendeur", AGENT: "Agent",
  DIRECTOR: "Directeur d'agence", HQ: "Siège", IT: "IT & Support",
};
document.getElementById("set-name").textContent =
  `${user.first_name || ""} ${user.last_name || ""}`.trim();
document.getElementById("set-email").textContent = user.email || "";
document.getElementById("set-role").textContent = roleLabels[user.role] || user.role || "";

document.getElementById("logout-btn").addEventListener("click", () => {
  logout();
  window.location.href = "./index.html";
});

// ---------- Tabs ----------
const tabs = document.querySelectorAll(".profile-tab");
const panels = {
  biens: document.getElementById("panel-biens"),
  favoris: document.getElementById("panel-favoris"),
  settings: document.getElementById("panel-settings"),
};

function showTab(name) {
  Object.entries(panels).forEach(([key, el]) => el.classList.toggle("hidden", key !== name));
  tabs.forEach((btn) => {
    const active = btn.dataset.tab === name;
    btn.classList.toggle("bg-hibiscus", active);
    btn.classList.toggle("text-white", active);
    btn.classList.toggle("text-hibiscus", !active);
  });
}

tabs.forEach((btn) => btn.addEventListener("click", () => showTab(btn.dataset.tab)));
showTab("favoris"); // default tab (matches the wireframe)

// ---------- Favorites ----------
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

function card(property) {
  const area = property.area != null ? ` - ${property.area} m²` : "";
  return `
    <a href="./property.html?id=${property.id}"
       class="group block rounded-xl border border-hibiscus/40 bg-white overflow-hidden transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5">
      <div class="overflow-hidden">
        <img src="${primaryPhoto(property)}" alt="${escapeHtml(property.title)}" loading="lazy"
             class="w-full h-40 object-cover group-hover:scale-105 transition-transform duration-300" />
      </div>
      <div class="p-4">
        <h2 class="font-semibold">${escapeHtml(property.title)}${area}</h2>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(property.city)} · ${escapeHtml(property.postal_code)}</p>
      </div>
    </a>`;
}

async function loadFavorites() {
  const grid = document.getElementById("favoris-grid");
  const empty = document.getElementById("favoris-empty");
  try {
    const { data } = await api.listFavorites();
    document.getElementById("count-favoris").textContent = data.length;
    grid.innerHTML = data.map(card).join("");
    empty.classList.toggle("hidden", data.length > 0);
  } catch (err) {
    grid.innerHTML = `<p class="text-slate2">Impossible de charger vos favoris.</p>`;
  }
}

loadFavorites();
