// Agent dashboard: my properties, KPIs (AI), and planning (visits + meetings).
import { api, ai } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

// Staff only.
const STAFF = ["AGENT", "DIRECTOR", "HQ"];
const me = currentUser() || {};
if (!isLoggedIn() || !STAFF.includes(me.role)) {
  window.location.href = "./index.html";
}

const euro = new Intl.NumberFormat("fr-FR", { style: "currency", currency: "EUR", maximumFractionDigits: 0 });
const dateFmt = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium", timeStyle: "short" });

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (c) => (
    { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
  ));
}

// ----- Tabs -----
const tabs = document.querySelectorAll(".dash-tab");
const panels = {
  overview: document.getElementById("panel-overview"),
  biens: document.getElementById("panel-biens"),
  planning: document.getElementById("panel-planning"),
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
showTab("biens"); // default (matches the wireframe)

// ----- My properties -----
const STATUS = {
  DRAFT: ["Brouillon", "bg-slate2/20 text-slate2"],
  PENDING_REVIEW: ["À valider", "bg-gold/20 text-gold"],
  AVAILABLE: ["Disponible", "bg-green-100 text-green-700"],
  UNDER_OFFER: ["Sous offre", "bg-gold/20 text-gold"],
  SOLD: ["Vendu", "bg-hibiscus/15 text-hibiscus"],
  WITHDRAWN: ["Retiré", "bg-slate2/20 text-slate2"],
};

function primaryPhoto(p) {
  const photos = p.photos || [];
  const main = photos.find((x) => x.is_primary) || photos[0];
  return main ? main.url : "https://placehold.co/600x400?text=Ymmo";
}

function propertyCard(p) {
  const [label, classes] = STATUS[p.status] || [p.status, "bg-slate2/20 text-slate2"];
  const area = p.area != null ? ` - ${p.area} m²` : "";
  // On the dashboard, a card leads to the edit form (management view).
  return `
    <a href="./property-form.html?id=${p.id}" class="group block rounded-xl border border-hibiscus/40 bg-white overflow-hidden hover:shadow-lg transition-all">
      <div class="relative overflow-hidden">
        <img src="${primaryPhoto(p)}" alt="${escapeHtml(p.title)}" loading="lazy" class="w-full h-40 object-cover group-hover:scale-105 transition-transform duration-300" />
        <span class="absolute top-2 left-2 text-xs font-semibold px-2.5 py-1 rounded-full ${classes}">${label}</span>
      </div>
      <div class="p-4">
        <h3 class="font-semibold">${escapeHtml(p.title)}${area}</h3>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(p.city)} · ${escapeHtml(p.postal_code)}</p>
        <p class="text-sm font-semibold text-hibiscus mt-1">${euro.format(p.price)}</p>
      </div>
    </a>`;
}

async function loadProperties() {
  const grid = document.getElementById("biens-grid");
  const empty = document.getElementById("biens-empty");
  try {
    const { data } = await api.myProperties();
    grid.innerHTML = data.map(propertyCard).join("");
    empty.classList.toggle("hidden", data.length > 0);
  } catch {
    grid.innerHTML = `<p class="text-slate2">Impossible de charger vos biens.</p>`;
  }
}

// ----- KPIs -----
function kpiCard(label, value) {
  return `
    <div class="rounded-xl border border-hibiscus/30 bg-white p-4">
      <p class="text-3xl font-bold text-hibiscus">${value}</p>
      <p class="text-sm text-slate2 mt-1">${label}</p>
    </div>`;
}
async function loadKpis() {
  const grid = document.getElementById("kpi-grid");
  try {
    const k = await ai.dashboardKpis();
    grid.innerHTML = [
      kpiCard("Biens au total", k.total_properties),
      kpiCard("Disponibles", k.available),
      kpiCard("Vendus", k.sold),
      kpiCard("Prix moyen", euro.format(k.avg_available_price || 0)),
      kpiCard("Vues cumulées", k.total_views),
    ].join("");
  } catch {
    grid.innerHTML = `<p class="col-span-full text-slate2">KPIs indisponibles (service IA injoignable).</p>`;
  }
}

// ----- Planning -----
function visitItem(v) {
  return `<li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex justify-between gap-4">
      <span>Visite — bien #${v.property_id}</span>
      <span class="text-slate2 text-sm">${dateFmt.format(new Date(v.scheduled_at))} · ${escapeHtml(v.status)}</span>
    </li>`;
}
function meetingItem(m) {
  return `<li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex justify-between gap-4">
      <span>${escapeHtml(m.title)}</span>
      <span class="text-slate2 text-sm">${dateFmt.format(new Date(m.start_at))}</span>
    </li>`;
}
async function loadPlanning() {
  try {
    const [{ data: visits }, { data: meetings }] = await Promise.all([api.listVisits(), api.listMeetings()]);
    document.getElementById("visits-list").innerHTML = visits.map(visitItem).join("");
    document.getElementById("visits-empty").classList.toggle("hidden", visits.length > 0);
    document.getElementById("meetings-list").innerHTML = meetings.map(meetingItem).join("");
    document.getElementById("meetings-empty").classList.toggle("hidden", meetings.length > 0);
  } catch {
    document.getElementById("visits-list").innerHTML = `<li class="text-slate2">Impossible de charger le planning.</li>`;
  }
}

// ----- Boot -----
loadProperties();
loadKpis();
loadPlanning();
