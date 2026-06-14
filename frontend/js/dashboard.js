// Agent dashboard: my properties, KPIs (AI), and planning (visits + meetings).
import { api, ai } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";
import { enhanceTabsAria } from "./a11y.js";

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

const tabs = document.querySelectorAll(".dash-tab");
const panels = {
  overview: document.getElementById("panel-overview"),
  biens: document.getElementById("panel-biens"),
  valider: document.getElementById("panel-valider"),
  vendeurs: document.getElementById("panel-vendeurs"),
  planning: document.getElementById("panel-planning"),
  ventes: document.getElementById("panel-ventes"),
};
function showTab(name) {
  Object.entries(panels).forEach(([key, el]) => el.classList.toggle("hidden", key !== name));
  tabs.forEach((btn) => {
    const active = btn.dataset.tab === name;
    btn.classList.toggle("bg-hibiscus", active);
    btn.classList.toggle("text-white", active);
    btn.classList.toggle("text-hibiscus", !active);
    btn.setAttribute("aria-selected", String(active));
  });
}
enhanceTabsAria(tabs, panels);
tabs.forEach((btn) => btn.addEventListener("click", () => showTab(btn.dataset.tab)));
showTab("biens");

const STATUS = {
  DRAFT: ["Brouillon", "bg-slate2/20 text-slate2"],
  PENDING_REVIEW: ["À valider", "bg-amber-100 text-amber-800"],
  AVAILABLE: ["Disponible", "bg-green-100 text-green-700"],
  UNDER_OFFER: ["Sous offre", "bg-amber-100 text-amber-800"],
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

// Agent visit actions; REQUESTED is not a valid manual transition.
function visitActions(v) {
  const actions = [];
  if (v.status === "REQUESTED") actions.push(["CONFIRMED", "Confirmer"]);
  if (v.status === "CONFIRMED") actions.push(["COMPLETED", "Terminer"]);
  if (v.status === "REQUESTED" || v.status === "CONFIRMED") actions.push(["CANCELLED", "Annuler"]);
  return actions
    .map(
      ([status, label]) =>
        `<button data-visit="${v.id}" data-status="${status}" class="visit-act text-xs font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-2.5 py-1 rounded-md transition-colors">${label}</button>`
    )
    .join(" ");
}
function visitItem(v) {
  return `<li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex flex-wrap items-center justify-between gap-3">
      <span>Visite — bien #${v.property_id} <span class="text-slate2 text-sm">· ${dateFmt.format(new Date(v.scheduled_at))} · ${escapeHtml(v.status)}</span></span>
      <span class="flex gap-2">${visitActions(v)}</span>
    </li>`;
}

function meetingItem(m) {
  return `<li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex items-center justify-between gap-4">
      <span>${escapeHtml(m.title)} <span class="text-slate2 text-sm">· ${dateFmt.format(new Date(m.start_at))}</span></span>
      <button data-meeting="${m.id}" class="meeting-del text-xs font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-2.5 py-1 rounded-md transition-colors">Supprimer</button>
    </li>`;
}

async function loadPlanning() {
  const visitsList = document.getElementById("visits-list");
  const meetingsList = document.getElementById("meetings-list");
  try {
    const [{ data: visits }, { data: meetings }] = await Promise.all([api.listVisits(), api.listMeetings()]);

    visitsList.innerHTML = visits.map(visitItem).join("");
    document.getElementById("visits-empty").classList.toggle("hidden", visits.length > 0);
    visitsList.querySelectorAll(".visit-act").forEach((btn) =>
      btn.addEventListener("click", async () => {
        btn.disabled = true;
        try {
          await api.updateVisit(btn.dataset.visit, btn.dataset.status);
          loadPlanning();
        } catch {
          btn.disabled = false;
          alert("Action sur la visite impossible.");
        }
      })
    );

    meetingsList.innerHTML = meetings.map(meetingItem).join("");
    document.getElementById("meetings-empty").classList.toggle("hidden", meetings.length > 0);
    meetingsList.querySelectorAll(".meeting-del").forEach((btn) =>
      btn.addEventListener("click", async () => {
        if (!confirm("Supprimer cette réunion ?")) return;
        try {
          await api.deleteMeeting(btn.dataset.meeting);
          loadPlanning();
        } catch {
          alert("Suppression impossible (seul l'organisateur peut supprimer).");
        }
      })
    );
  } catch {
    visitsList.innerHTML = `<li class="text-slate2">Impossible de charger le planning.</li>`;
  }
}

const meetingForm = document.getElementById("meeting-form");
document.getElementById("meeting-toggle").addEventListener("click", () => meetingForm.classList.toggle("hidden"));
meetingForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const err = document.getElementById("meeting-error");
  err.classList.add("hidden");
  const payload = {
    title: document.getElementById("m-title").value.trim(),
    agency_id: Number(document.getElementById("m-agency").value),
    start_at: new Date(document.getElementById("m-start").value).toISOString(),
    end_at: new Date(document.getElementById("m-end").value).toISOString(),
  };
  const loc = document.getElementById("m-location").value.trim();
  if (loc) payload.location = loc;

  try {
    await api.createMeeting(payload);
    meetingForm.reset();
    meetingForm.classList.add("hidden");
    loadPlanning();
  } catch (e2) {
    err.textContent = e2.status === 409 ? "Vous avez déjà une réunion à cette date." : "Création impossible (vérifiez les champs).";
    err.classList.remove("hidden");
  }
});

const SALE_STATUS = {
  OFFER: ["Offre", "bg-amber-100 text-amber-800"],
  PRELIMINARY_CONTRACT: ["Compromis", "bg-amber-100 text-amber-800"],
  DEED: ["Acte", "bg-green-100 text-green-700"],
  COMPLETED: ["Finalisée", "bg-hibiscus/15 text-hibiscus"],
  CANCELLED: ["Annulée", "bg-slate2/20 text-slate2"],
};
// Sequential lifecycle: each status offers the next step (+ cancel).
const SALE_NEXT = {
  OFFER: "PRELIMINARY_CONTRACT",
  PRELIMINARY_CONTRACT: "DEED",
  DEED: "COMPLETED",
};
const SALE_NEXT_LABEL = {
  PRELIMINARY_CONTRACT: "Passer au compromis",
  DEED: "Passer à l'acte",
  COMPLETED: "Finaliser la vente",
};

function saleActions(s) {
  if (s.status === "COMPLETED" || s.status === "CANCELLED") return "";
  const next = SALE_NEXT[s.status];
  const btns = [];
  if (next) btns.push(`<button data-sale="${s.id}" data-status="${next}" class="sale-act text-xs font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-2.5 py-1 rounded-md transition-colors">${SALE_NEXT_LABEL[next]}</button>`);
  btns.push(`<button data-sale="${s.id}" data-status="CANCELLED" class="sale-act text-xs font-medium border border-slate2 text-slate2 hover:bg-slate2 hover:text-white px-2.5 py-1 rounded-md transition-colors">Annuler</button>`);
  return btns.join(" ");
}
function saleItem(s) {
  const [label, cls] = SALE_STATUS[s.status] || [s.status, "bg-slate2/20 text-slate2"];
  const price = s.negotiated_price != null ? euro.format(s.negotiated_price) : "—";
  return `<li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p class="font-semibold">Dossier #${s.id} · Bien #${s.property_id} → Acheteur #${s.buyer_id}</p>
          <p class="text-sm text-slate2">Prix négocié : ${price}</p>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold px-2.5 py-1 rounded-full ${cls}">${label}</span>
          <button data-sale-detail="${s.id}" class="sale-detail text-xs font-medium border border-slate2 text-slate2 hover:bg-slate2 hover:text-white px-2.5 py-1 rounded-md transition-colors">Détails</button>
          ${saleActions(s)}
        </div>
      </div>
      <p id="sale-detail-${s.id}" class="hidden text-sm text-slate2 mt-2 border-t border-hibiscus/20 pt-2"></p>
    </li>`;
}

const dateOnly = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium" });
function saleDetailText(s) {
  const fmt = (d) => (d ? dateOnly.format(new Date(d)) : "—");
  return `Offre : ${fmt(s.offer_date)} · Compromis : ${fmt(s.contract_date)} · Acte : ${fmt(s.deed_date)}`;
}

async function loadSales() {
  const list = document.getElementById("sales-list");
  try {
    const { data } = await api.listSales();
    list.innerHTML = data.map(saleItem).join("");
    document.getElementById("sales-empty").classList.toggle("hidden", data.length > 0);
    list.querySelectorAll(".sale-act").forEach((btn) =>
      btn.addEventListener("click", async () => {
        btn.disabled = true;
        try {
          await api.updateSale(btn.dataset.sale, { status: btn.dataset.status });
          loadSales();
        } catch {
          btn.disabled = false;
          alert("Transition impossible (les étapes doivent être séquentielles).");
        }
      })
    );

    list.querySelectorAll(".sale-detail").forEach((btn) =>
      btn.addEventListener("click", async () => {
        const id = btn.dataset.saleDetail;
        const row = document.getElementById(`sale-detail-${id}`);
        if (!row.classList.contains("hidden")) {
          row.classList.add("hidden");
          return;
        }
        row.textContent = "Chargement…";
        row.classList.remove("hidden");
        try {
          const s = await api.getSale(id);
          row.textContent = saleDetailText(s);
        } catch {
          row.textContent = "Détails indisponibles.";
        }
      })
    );
  } catch {
    list.innerHTML = `<li class="text-slate2">Impossible de charger les dossiers de vente.</li>`;
  }
}

const saleForm = document.getElementById("sale-form");
document.getElementById("sale-toggle").addEventListener("click", () => saleForm.classList.toggle("hidden"));
saleForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const err = document.getElementById("sale-error");
  err.classList.add("hidden");
  const payload = {
    property_id: Number(document.getElementById("s-property").value),
    buyer_id: Number(document.getElementById("s-buyer").value),
  };
  const price = document.getElementById("s-price").value;
  if (price) payload.negotiated_price = Number(price);
  const offer = document.getElementById("s-offer").value;
  if (offer) payload.offer_date = new Date(offer).toISOString();

  try {
    await api.createSale(payload);
    saleForm.reset();
    saleForm.classList.add("hidden");
    loadSales();
  } catch (e2) {
    err.textContent =
      e2.status === 404 ? "Bien ou acheteur introuvable."
      : e2.status === 409 ? "Un dossier actif existe déjà pour ce bien."
      : e2.status === 400 ? "Vérifiez l'acheteur (rôle BUYER) et la date."
      : "Création impossible.";
    err.classList.remove("hidden");
  }
});

function pendingCard(p) {
  const area = p.area != null ? ` - ${p.area} m²` : "";
  return `
    <div class="rounded-xl border border-gold/60 bg-white overflow-hidden">
      <img src="${primaryPhoto(p)}" alt="${escapeHtml(p.title)}" loading="lazy" class="w-full h-40 object-cover" />
      <div class="p-4">
        <h3 class="font-semibold">${escapeHtml(p.title)}${area}</h3>
        <p class="text-sm text-slate2 mt-1">${escapeHtml(p.city)} · ${escapeHtml(p.postal_code)}</p>
        <p class="text-sm font-semibold text-hibiscus mt-1">${euro.format(p.price)}</p>
        <button data-validate="${p.id}" class="validate-btn mt-3 w-full bg-hibiscus hover:bg-hibiscus/90 text-white text-sm font-medium py-2 rounded-md transition-colors">
          Valider et publier
        </button>
      </div>
    </div>`;
}

async function loadPending() {
  const grid = document.getElementById("pending-grid");
  const empty = document.getElementById("pending-empty");
  try {
    const { data } = await api.pendingProperties();
    grid.innerHTML = data.map(pendingCard).join("");
    empty.classList.toggle("hidden", data.length > 0);
    grid.querySelectorAll(".validate-btn").forEach((btn) =>
      btn.addEventListener("click", async () => {
        btn.disabled = true;
        try {
          await api.validateProperty(btn.dataset.validate);
          loadPending();
          loadProperties(); // it's now assigned to me
        } catch {
          btn.disabled = false;
          alert("Validation impossible.");
        }
      })
    );
  } catch {
    grid.innerHTML = `<p class="text-slate2">Impossible de charger les biens à valider.</p>`;
  }
}

function sellerAppItem(a) {
  const u = a.user || {};
  const name = `${escapeHtml(u.first_name || "")} ${escapeHtml(u.last_name || "")}`.trim() || `Utilisateur #${a.user_id}`;
  const motivation = a.motivation ? `<p class="text-sm text-slate2 mt-1 italic">« ${escapeHtml(a.motivation)} »</p>` : "";
  return `
    <li class="border border-hibiscus/30 rounded-lg px-4 py-3 bg-white flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="font-semibold">${name}</p>
        <p class="text-sm text-slate2">${escapeHtml(u.email || "")}</p>
        ${motivation}
      </div>
      <div class="flex items-center gap-2">
        <button data-approve="${a.id}" class="seller-approve text-sm font-medium bg-hibiscus hover:bg-hibiscus/90 text-white px-3 py-1.5 rounded-md transition-colors">Approuver</button>
        <button data-reject="${a.id}" class="seller-reject text-sm font-medium border border-slate2 text-slate2 hover:bg-slate2 hover:text-white px-3 py-1.5 rounded-md transition-colors">Refuser</button>
      </div>
    </li>`;
}

async function loadSellerApps() {
  const list = document.getElementById("sellers-list");
  try {
    const { data } = await api.pendingSellerApplications();
    list.innerHTML = data.map(sellerAppItem).join("");
    document.getElementById("sellers-empty").classList.toggle("hidden", data.length > 0);

    const review = (id, approve) => async () => {
      try {
        await (approve ? api.approveSellerApplication(id) : api.rejectSellerApplication(id));
        loadSellerApps();
      } catch {
        alert("Action impossible.");
      }
    };
    list.querySelectorAll(".seller-approve").forEach((b) => b.addEventListener("click", review(b.dataset.approve, true)));
    list.querySelectorAll(".seller-reject").forEach((b) => b.addEventListener("click", review(b.dataset.reject, false)));
  } catch {
    list.innerHTML = `<li class="text-slate2">Impossible de charger les demandes.</li>`;
  }
}

loadProperties();
loadKpis();
loadPlanning();
loadSales();
loadPending();
loadSellerApps();
