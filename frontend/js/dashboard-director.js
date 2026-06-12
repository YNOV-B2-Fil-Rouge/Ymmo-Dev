// Director / HQ dashboard.
// A director only sees their own agency (scoping is enforced by the API);
// HQ gets the national view. KPIs and the director's analysis are computed
// from the (already scoped) properties, so no other agency's data leaks.
import { api, ai } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

const ALLOWED = ["DIRECTOR", "HQ"];
const me = currentUser() || {};
if (!isLoggedIn() || !ALLOWED.includes(me.role)) {
  window.location.href = "./index.html";
}
const isHQ = me.role === "HQ";

// HQ sees a national view with agency-oriented labels.
const AGENCIES = { 1: "Ymmo Siège (Aix)", 2: "Ymmo Paris", 3: "Ymmo Lyon", 4: "Ymmo Marseille" };
if (isHQ) {
  document.querySelector('[data-tab="perf"]').textContent = "Performance globale";
  document.querySelector('[data-tab="biens"]').textContent = "Gestion des agences";
  document.getElementById("perf-title").textContent = "Performance globale";
  document.getElementById("biens-title").textContent = "Gestion des agences";
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
showTab("biens");

// ----- Shared helpers -----
const STATUS = {
  DRAFT: ["Brouillon", "bg-slate2/20 text-slate2"], PENDING_REVIEW: ["À valider", "bg-gold/20 text-gold"],
  AVAILABLE: ["Disponible", "bg-green-100 text-green-700"], UNDER_OFFER: ["Sous offre", "bg-gold/20 text-gold"],
  SOLD: ["Vendu", "bg-hibiscus/15 text-hibiscus"], WITHDRAWN: ["Retiré", "bg-slate2/20 text-slate2"],
};
function primaryPhoto(p) {
  const ph = (p.photos || []).find((x) => x.is_primary) || (p.photos || [])[0];
  return ph ? ph.url : "https://placehold.co/600x400?text=Ymmo";
}
function table(headers, rows) {
  return `<table class="w-full text-sm border border-hibiscus/20 rounded-lg overflow-hidden">
    <thead class="bg-hibiscus/10 text-left"><tr>${headers.map((h) => `<th class="px-3 py-2 font-semibold">${h}</th>`).join("")}</tr></thead>
    <tbody>${rows.map((r) => `<tr class="border-t border-hibiscus/10">${r.map((c) => `<td class="px-3 py-2">${c}</td>`).join("")}</tr>`).join("")}</tbody></table>`;
}
function kpiCard(label, value) {
  return `<div class="rounded-xl border border-hibiscus/30 bg-white p-4"><p class="text-3xl font-bold text-hibiscus">${value}</p><p class="text-sm text-slate2 mt-1">${label}</p></div>`;
}

// ----- Properties + KPIs + analysis (all from the scoped data) -----
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

function renderKpis(props) {
  const available = props.filter((p) => p.status === "AVAILABLE");
  const sold = props.filter((p) => p.status === "SOLD").length;
  const avg = available.length ? Math.round(available.reduce((s, p) => s + p.price, 0) / available.length) : 0;
  const views = props.reduce((s, p) => s + (p.view_count || 0), 0);
  document.getElementById("kpi-grid").innerHTML = [
    kpiCard("Biens", props.length), kpiCard("Disponibles", available.length),
    kpiCard("Vendus", sold), kpiCard("Prix moyen", euro.format(avg)), kpiCard("Vues cumulées", views),
  ].join("");
}

// Director: analysis computed from the agency's own properties.
function renderAgencyAnalysis(props) {
  // Avg price per m² by category.
  const byCat = {};
  props.forEach((p) => {
    if (!p.area) return;
    const cat = p.category?.label || "—";
    (byCat[cat] ||= []).push(p.price / p.area);
  });
  const catRows = Object.entries(byCat).map(([cat, arr]) => [
    escapeHtml(cat), euro.format(Math.round(arr.reduce((a, b) => a + b, 0) / arr.length)), arr.length,
  ]);

  // Most viewed.
  const top = [...props].sort((a, b) => (b.view_count || 0) - (a.view_count || 0)).slice(0, 6);

  document.getElementById("ia-body").innerHTML = `
    <div>
      <h3 class="font-semibold text-hibiscus mb-3">Prix moyen au m² par catégorie (votre agence)</h3>
      ${catRows.length ? table(["Catégorie", "€/m² moyen", "Biens"], catRows) : '<p class="text-slate2">Pas assez de données.</p>'}
    </div>
    <div>
      <h3 class="font-semibold text-hibiscus mb-3">Vos biens les plus consultés</h3>
      <ul class="space-y-2">${top.map((p) => `<li class="border border-hibiscus/30 rounded-lg px-4 py-2 bg-white flex justify-between"><span>${escapeHtml(p.title)} — ${escapeHtml(p.city)}</span><span class="text-slate2">${p.view_count || 0} vues</span></li>`).join("")}</ul>
    </div>`;
}

// HQ: national market analysis from the Python AI service.
async function renderNationalAnalysis() {
  const body = document.getElementById("ia-body");
  body.innerHTML = `
    <div><h3 class="font-semibold text-hibiscus mb-3">Zones à fort potentiel</h3><div id="zones-table"></div></div>
    <div><h3 class="font-semibold text-hibiscus mb-3">Tendances de prix (€/m²)</h3><div id="trends-table"></div></div>
    <div><h3 class="font-semibold text-hibiscus mb-3">Biens les plus consultés</h3><ul id="popular-list" class="space-y-2"></ul></div>`;
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

// HQ only: a per-agency summary (properties grouped by agency).
function renderAgenciesOverview(props) {
  const box = document.getElementById("agencies-overview");
  const byAgency = {};
  props.forEach((p) => {
    const a = (byAgency[p.agency_id] ||= { total: 0, available: 0, sold: 0 });
    a.total++;
    if (p.status === "AVAILABLE") a.available++;
    if (p.status === "SOLD") a.sold++;
  });
  const rows = Object.entries(byAgency).map(([id, a]) => [
    escapeHtml(AGENCIES[id] || `Agence #${id}`), a.total, a.available, a.sold,
  ]);
  box.innerHTML = `<h3 class="font-semibold text-hibiscus mb-3">Synthèse par agence</h3>${table(["Agence", "Biens", "Disponibles", "Vendus"], rows)}`;
  box.classList.remove("hidden");
}

async function loadProperties() {
  const grid = document.getElementById("biens-grid");
  try {
    const { data } = await api.allProperties();
    grid.innerHTML = data.length ? data.map(propertyCard).join("") : `<p class="text-slate2">Aucun bien.</p>`;
    renderKpis(data);
    if (isHQ) {
      renderAgenciesOverview(data);
      renderNationalAnalysis();
    } else {
      renderAgencyAnalysis(data);
    }
  } catch {
    grid.innerHTML = `<p class="text-slate2">Impossible de charger les biens.</p>`;
  }
}

// ----- Collaborators -----
const ROLE_LABELS = { AGENT: "Agent", DIRECTOR: "Directeur", HQ: "Siège", IT: "IT & Support" };
async function loadCollaborators() {
  try {
    const { data } = await api.collaborators();
    document.getElementById("collab-table").innerHTML = table(
      ["Nom", "Email", "Rôle"],
      data.map((u) => [`${escapeHtml(u.last_name)} ${escapeHtml(u.first_name)}`, escapeHtml(u.email), ROLE_LABELS[u.role?.code] || u.role?.code || ""])
    );
  } catch {
    document.getElementById("collab-table").innerHTML = `<p class="text-slate2">Impossible de charger l'annuaire.</p>`;
  }
}

// ----- Boot -----
loadProperties();
loadCollaborators();
