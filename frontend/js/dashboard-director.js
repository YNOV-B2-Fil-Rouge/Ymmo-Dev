// Director / HQ dashboard: agency performance, property management,
// AI strategic analysis, and the collaborator directory.
import { api, ai } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

const ALLOWED = ["DIRECTOR", "HQ"];
const me = currentUser() || {};
if (!isLoggedIn() || !ALLOWED.includes(me.role)) {
  window.location.href = "./index.html";
}

const euro = new Intl.NumberFormat("fr-FR", { style: "currency", currency: "EUR", maximumFractionDigits: 0 });
const escapeHtml = (v) => String(v ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

// ----- Tabs -----
const tabs = document.querySelectorAll(".dash-tab");
const panels = {
  perf: document.getElementById("panel-perf"),
  biens: document.getElementById("panel-biens"),
  ia: document.getElementById("panel-ia"),
  annuaire: document.getElementById("panel-annuaire"),
};
function showTab(name) {
  Object.entries(panels).forEach(([k, el]) => el.classList.toggle("hidden", k !== name));
  tabs.forEach((b) => {
    const active = b.dataset.tab === name;
    b.classList.toggle("bg-hibiscus", active);
    b.classList.toggle("text-white", active);
    b.classList.toggle("text-hibiscus", !active);
  });
}
tabs.forEach((b) => b.addEventListener("click", () => showTab(b.dataset.tab)));
showTab("biens"); // default (matches wireframe)

// ----- Property management -----
const STATUS = {
  DRAFT: ["Brouillon", "bg-slate2/20 text-slate2"], PENDING_REVIEW: ["À valider", "bg-gold/20 text-gold"],
  AVAILABLE: ["Disponible", "bg-green-100 text-green-700"], UNDER_OFFER: ["Sous offre", "bg-gold/20 text-gold"],
  SOLD: ["Vendu", "bg-hibiscus/15 text-hibiscus"], WITHDRAWN: ["Retiré", "bg-slate2/20 text-slate2"],
};
function primaryPhoto(p) {
  const ph = (p.photos || []).find((x) => x.is_primary) || (p.photos || [])[0];
  return ph ? ph.url : "https://placehold.co/600x400?text=Ymmo";
}
function propertyCard(p) {
  const [label, classes] = STATUS[p.status] || [p.status, "bg-slate2/20 text-slate2"];
  return `
    <a href="./property-form.html?id=${p.id}" class="group block rounded-xl border border-hibiscus/40 bg-white overflow-hidden hover:shadow-lg transition-all">
      <div class="relative overflow-hidden">
        <img src="${primaryPhoto(p)}" alt="${escapeHtml(p.title)}" loading="lazy" class="w-full h-40 object-cover group-hover:scale-105 transition-transform duration-300" />
        <span class="absolute top-2 left-2 text-xs font-semibold px-2.5 py-1 rounded-full ${classes}">${label}</span>
      </div>
      <div class="p-4">
        <h3 class="font-semibold">${escapeHtml(p.title)}${p.area != null ? ` - ${p.area} m²` : ""}</h3>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(p.city)} · ${escapeHtml(p.postal_code)}</p>
        <p class="text-sm font-semibold text-hibiscus mt-1">${euro.format(p.price)}</p>
      </div>
    </a>`;
}
async function loadProperties() {
  const grid = document.getElementById("biens-grid");
  try {
    const { data } = await api.allProperties();
    grid.innerHTML = data.length ? data.map(propertyCard).join("") : `<p class="text-slate2">Aucun bien.</p>`;
  } catch {
    grid.innerHTML = `<p class="text-slate2">Impossible de charger les biens.</p>`;
  }
}

// ----- Performance KPIs -----
function kpiCard(label, value) {
  return `<div class="rounded-xl border border-hibiscus/30 bg-white p-4"><p class="text-3xl font-bold text-hibiscus">${value}</p><p class="text-sm text-slate2 mt-1">${label}</p></div>`;
}
async function loadKpis() {
  const grid = document.getElementById("kpi-grid");
  try {
    const k = await ai.dashboardKpis();
    grid.innerHTML = [
      kpiCard("Biens au total", k.total_properties), kpiCard("Disponibles", k.available),
      kpiCard("Vendus", k.sold), kpiCard("Prix moyen", euro.format(k.avg_available_price || 0)),
      kpiCard("Vues cumulées", k.total_views),
    ].join("");
  } catch {
    grid.innerHTML = `<p class="col-span-full text-slate2">KPIs indisponibles (service IA injoignable).</p>`;
  }
}

// ----- AI strategic analysis -----
function table(headers, rows) {
  return `<table class="w-full text-sm border border-hibiscus/20 rounded-lg overflow-hidden">
    <thead class="bg-hibiscus/10 text-left"><tr>${headers.map((h) => `<th class="px-3 py-2 font-semibold">${h}</th>`).join("")}</tr></thead>
    <tbody>${rows.map((r) => `<tr class="border-t border-hibiscus/10">${r.map((c) => `<td class="px-3 py-2">${c}</td>`).join("")}</tr>`).join("")}</tbody></table>`;
}
async function loadAi() {
  try {
    const z = await ai.zones();
    document.getElementById("zones-table").innerHTML = table(
      ["Ville", "Annonces", "Vues/annonce", "€/m² moyen", "Score"],
      z.data.map((r) => [escapeHtml(r.city), r.listings, r.views_per_listing, euro.format(r.avg_price_per_m2), r.opportunity_score])
    );
  } catch { document.getElementById("zones-table").innerHTML = `<p class="text-slate2">Indisponible.</p>`; }

  try {
    const t = await ai.trends();
    document.getElementById("trends-table").innerHTML = table(
      ["Ville", "Catégorie", "€/m² moyen", "Annonces"],
      t.data.map((r) => [escapeHtml(r.city), escapeHtml(r.category), euro.format(r.avg_price_per_m2), r.listings])
    );
  } catch { document.getElementById("trends-table").innerHTML = `<p class="text-slate2">Indisponible.</p>`; }

  try {
    const p = await ai.popular(6);
    document.getElementById("popular-list").innerHTML = p.data
      .map((x) => `<li class="border border-hibiscus/30 rounded-lg px-4 py-2 bg-white flex justify-between"><span>${escapeHtml(x.title)} — ${escapeHtml(x.city)}</span><span class="text-slate2">${x.view_count} vues</span></li>`)
      .join("");
  } catch { document.getElementById("popular-list").innerHTML = `<li class="text-slate2">Indisponible.</li>`; }
}

// ----- Collaborator directory -----
const ROLE_LABELS = { AGENT: "Agent", DIRECTOR: "Directeur", HQ: "Siège", IT: "IT & Support" };
async function loadCollaborators() {
  try {
    const { data } = await api.collaborators();
    document.getElementById("collab-table").innerHTML = table(
      ["Nom", "Email", "Rôle"],
      data.map((u) => [
        `${escapeHtml(u.last_name)} ${escapeHtml(u.first_name)}`,
        escapeHtml(u.email),
        ROLE_LABELS[u.role?.code] || u.role?.code || "",
      ])
    );
  } catch {
    document.getElementById("collab-table").innerHTML = `<p class="text-slate2">Impossible de charger l'annuaire.</p>`;
  }
}

// ----- Boot -----
loadProperties();
loadKpis();
loadAi();
loadCollaborators();
