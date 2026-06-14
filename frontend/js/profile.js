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
  visites: document.getElementById("panel-visites"),
  alertes: document.getElementById("panel-alertes"),
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

// ---------- My visits (request, then cancel here) ----------
const dateFmt = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium", timeStyle: "short" });
const VISIT_STATUS = {
  REQUESTED: ["Demandée", "bg-gold/20 text-gold"],
  CONFIRMED: ["Confirmée", "bg-green-100 text-green-700"],
  CANCELLED: ["Annulée", "bg-slate2/20 text-slate2"],
  COMPLETED: ["Terminée", "bg-hibiscus/15 text-hibiscus"],
};

function visitItem(v) {
  const [label, cls] = VISIT_STATUS[v.status] || [v.status, "bg-slate2/20 text-slate2"];
  // A client may cancel a visit that is still open.
  const cancellable = v.status === "REQUESTED" || v.status === "CONFIRMED";
  return `
    <li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex items-center justify-between gap-4">
      <div>
        <p class="font-semibold">Bien #${v.property_id}</p>
        <p class="text-sm text-slate2">${dateFmt.format(new Date(v.scheduled_at))}</p>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-xs font-semibold px-2.5 py-1 rounded-full ${cls}">${label}</span>
        ${cancellable ? `<button data-visit="${v.id}" class="visit-cancel text-sm font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-3 py-1.5 rounded-md transition-colors">Annuler</button>` : ""}
      </div>
    </li>`;
}

async function loadVisits() {
  const list = document.getElementById("visites-list");
  const empty = document.getElementById("visites-empty");
  try {
    const { data } = await api.listVisits();
    document.getElementById("count-visites").textContent = data.length;
    list.innerHTML = data.map(visitItem).join("");
    empty.classList.toggle("hidden", data.length > 0);

    list.querySelectorAll(".visit-cancel").forEach((btn) =>
      btn.addEventListener("click", async () => {
        btn.disabled = true;
        try {
          await api.updateVisit(btn.dataset.visit, "CANCELLED");
          loadVisits();
        } catch {
          btn.disabled = false;
          alert("Annulation impossible.");
        }
      })
    );
  } catch {
    list.innerHTML = `<li class="text-slate2">Impossible de charger vos visites.</li>`;
  }
}

// ---------- My alerts (created from the search page) ----------
const CATEGORY_LABELS = {
  1: "Maison", 2: "Appartement", 3: "Studio", 4: "Terrain", 5: "Bureau", 6: "Local commercial", 7: "Entrepôt",
};
const euro = new Intl.NumberFormat("fr-FR", { style: "currency", currency: "EUR", maximumFractionDigits: 0 });

function alertSummary(a) {
  const parts = [];
  if (a.city) parts.push(escapeHtml(a.city));
  if (a.category_id) parts.push(CATEGORY_LABELS[a.category_id] || `Catégorie ${a.category_id}`);
  if (a.min_price != null) parts.push(`≥ ${euro.format(a.min_price)}`);
  if (a.max_price != null) parts.push(`≤ ${euro.format(a.max_price)}`);
  if (a.min_area != null) parts.push(`≥ ${a.min_area} m²`);
  if (a.max_energy) parts.push(`DPE ≤ ${a.max_energy}`);
  return parts.length ? parts.join(" · ") : "Tous les biens";
}

function alertItem(a) {
  return `
    <li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex items-center justify-between gap-4">
      <span class="text-sm">${alertSummary(a)}</span>
      <button data-alert="${a.id}" class="alert-del text-sm font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-3 py-1.5 rounded-md transition-colors">Supprimer</button>
    </li>`;
}

async function loadAlerts() {
  const list = document.getElementById("alertes-list");
  const empty = document.getElementById("alertes-empty");
  try {
    const { data } = await api.listAlerts();
    document.getElementById("count-alertes").textContent = data.length;
    list.innerHTML = data.map(alertItem).join("");
    empty.classList.toggle("hidden", data.length > 0);

    list.querySelectorAll(".alert-del").forEach((btn) =>
      btn.addEventListener("click", async () => {
        btn.disabled = true;
        try {
          await api.deleteAlert(btn.dataset.alert);
          loadAlerts();
        } catch {
          btn.disabled = false;
          alert("Suppression impossible.");
        }
      })
    );
  } catch {
    list.innerHTML = `<li class="text-slate2">Impossible de charger vos alertes.</li>`;
  }
}

loadVisits();
loadAlerts();
renderSellerCta();

// ---------- Become a seller (buyers only) ----------
async function renderSellerCta() {
  if (user.role !== "BUYER") return; // sellers/staff don't see this
  const box = document.getElementById("seller-cta");
  box.classList.remove("hidden");

  let app = null;
  try {
    const res = await api.mySellerApplication();
    app = res.data; // may be null
  } catch {
    /* treat as no application */
  }

  // Pending or approved: just show the status.
  if (app && app.status === "PENDING") {
    box.innerHTML = `
      <p class="font-semibold text-gold">Demande pour devenir vendeur</p>
      <p class="text-sm text-slate2 mt-1">Votre demande est en cours d'examen par un agent.</p>`;
    return;
  }
  if (app && app.status === "APPROVED") {
    box.innerHTML = `
      <p class="font-semibold text-green-700">Demande approuvée 🎉</p>
      <p class="text-sm text-slate2 mt-1">Déconnectez-vous puis reconnectez-vous pour activer votre compte vendeur et publier des annonces.</p>`;
    return;
  }

  // No application yet, or a previous one was rejected: show the form.
  const rejected = app && app.status === "REJECTED";
  box.innerHTML = `
    <p class="font-semibold">Devenir vendeur</p>
    <p class="text-sm text-slate2 mt-1">Proposez vos biens à la vente. Votre demande sera validée par un agent.</p>
    ${rejected ? `<p class="text-sm text-hibiscus mt-1">Votre précédente demande a été refusée. Vous pouvez en soumettre une nouvelle.</p>` : ""}
    <textarea id="seller-motivation" rows="2" maxlength="500" placeholder="Quelques mots sur votre projet (optionnel)"
      class="mt-3 w-full rounded-md border border-hibiscus/60 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-hibiscus/30"></textarea>
    <button id="seller-apply" class="mt-2 bg-hibiscus hover:bg-hibiscus/90 text-white text-sm font-medium px-5 py-2 rounded-md transition-colors">
      Demander à devenir vendeur
    </button>
    <p id="seller-apply-msg" class="hidden mt-2 text-sm"></p>`;

  document.getElementById("seller-apply").addEventListener("click", async () => {
    const btn = document.getElementById("seller-apply");
    const msg = document.getElementById("seller-apply-msg");
    const motivation = document.getElementById("seller-motivation").value.trim();
    btn.disabled = true;
    try {
      await api.applyAsSeller(motivation ? { motivation } : {});
      renderSellerCta(); // re-render into the "pending" state
    } catch (err) {
      msg.textContent =
        err.status === 409 ? "Vous avez déjà une demande en cours."
        : err.status === 403 ? "Seuls les acheteurs peuvent faire cette demande."
        : "Envoi impossible pour le moment.";
      msg.className = "mt-2 text-sm text-hibiscus";
      btn.disabled = false;
    }
  });
}

// Refresh the account from the server (also validates the token: a 401 means
// the session expired, so we send the user back to the login page).
api.me()
  .then((fresh) => {
    document.getElementById("set-email").textContent = fresh.email || user.email || "";
    document.getElementById("set-role").textContent = roleLabels[fresh.role] || fresh.role || "";
  })
  .catch((err) => {
    if (err.status === 401) {
      logout();
      window.location.href = "./auth.html";
    }
  });
