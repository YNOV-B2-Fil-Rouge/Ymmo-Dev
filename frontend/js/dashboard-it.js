// IT dashboard: user management, the access-rights matrix, monitoring and the
// collaborator directory.
import { api, ai } from "./api.js";
import { CONFIG } from "./config.js";
import { currentUser, isLoggedIn } from "./auth.js";

const me = currentUser() || {};
if (!isLoggedIn() || !["IT", "HQ"].includes(me.role)) {
  window.location.href = "./index.html";
}

const escapeHtml = (v) => String(v ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

// ----- Tabs -----
const tabs = document.querySelectorAll(".dash-tab");
const panels = {
  users: document.getElementById("panel-users"),
  matrix: document.getElementById("panel-matrix"),
  monitoring: document.getElementById("panel-monitoring"),
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
showTab("matrix"); // default (matches the wireframe)

function table(headers, rows) {
  return `<table class="w-full text-sm border border-hibiscus/20 rounded-lg overflow-hidden">
    <thead class="bg-hibiscus/10 text-left"><tr>${headers.map((h) => `<th class="px-3 py-2 font-semibold">${h}</th>`).join("")}</tr></thead>
    <tbody>${rows.map((r) => `<tr class="border-t border-hibiscus/10">${r.map((c) => `<td class="px-3 py-2">${c}</td>`).join("")}</tr>`).join("")}</tbody></table>`;
}

const ROLE_LABELS = { VISITOR: "Visiteur", BUYER: "Acheteur", SELLER: "Vendeur", AGENT: "Agent", DIRECTOR: "Directeur", HQ: "Siège", IT: "IT & Support" };

// ----- Users -----
async function loadUsers() {
  try {
    const { data } = await api.allUsers();
    document.getElementById("users-table").innerHTML = table(
      ["Nom", "Email", "Rôle", "Agence"],
      data.map((u) => [
        `${escapeHtml(u.last_name)} ${escapeHtml(u.first_name)}`,
        escapeHtml(u.email),
        ROLE_LABELS[u.role?.code] || u.role?.code || "",
        u.agency_id ?? "—",
      ])
    );
  } catch {
    document.getElementById("users-table").innerHTML = `<p class="text-slate2">Impossible de charger les utilisateurs.</p>`;
  }
}

// ----- Access matrix -----
const DEPT_LABELS = { Management: "Direction", Sales: "Commercial", "Comm. & Mktg": "Comm. & Mktg", "Admin/HR": "Admin/RH", "IT & Support": "IT & Support" };
function levelCell(level) {
  if (level === "READ_WRITE") return `<span class="text-green-700 font-semibold">✔ R/W</span>`;
  if (level === "READ") return `<span class="text-green-700 font-semibold">✔ R</span>`;
  return `<span class="text-hibiscus font-semibold">🔒 Interdit</span>`;
}
async function loadMatrix() {
  const box = document.getElementById("matrix-table");
  try {
    const { data } = await api.permissionsMatrix();
    // Pivot rows {requester, target, level} into a grid.
    const requesters = [...new Set(data.map((d) => d.requester))];
    const targets = [...new Set(data.map((d) => d.target))];
    const lookup = {};
    data.forEach((d) => { lookup[`${d.requester}|${d.target}`] = d.level; });

    const headers = ["Pôle métier", ...targets.map((t) => `Dossier ${DEPT_LABELS[t] || t}`)];
    const rows = requesters.map((r) => [
      `<span class="font-semibold">${DEPT_LABELS[r] || r}</span>`,
      ...targets.map((t) => levelCell(lookup[`${r}|${t}`])),
    ]);
    box.innerHTML = table(headers, rows);
  } catch {
    box.innerHTML = `<p class="text-slate2">Impossible de charger la matrice des droits.</p>`;
  }
}

// ----- Monitoring (client-side service pings) -----
function statusCard(label, ok, detail = "") {
  const color = ok ? "text-green-700" : "text-hibiscus";
  const dot = ok ? "bg-green-500" : "bg-hibiscus";
  return `<div class="rounded-xl border border-hibiscus/30 bg-white p-4">
      <div class="flex items-center gap-2"><span class="w-2.5 h-2.5 rounded-full ${dot}"></span><p class="font-semibold">${label}</p></div>
      <p class="text-sm ${color} mt-1">${ok ? "En ligne" : "Hors ligne"}${detail ? ` · ${detail}` : ""}</p>
    </div>`;
}
async function loadMonitoring() {
  const grid = document.getElementById("monitoring-grid");
  // API Go (also reports the DB status).
  let apiOk = false, dbOk = false;
  try {
    const r = await fetch(`${CONFIG.API_BASE.replace("/api/v1", "")}/health`);
    const j = await r.json();
    apiOk = r.ok;
    dbOk = j.database === "up";
  } catch { /* down */ }
  // Python AI service.
  // The AI service is internal-only (not exposed on localhost), so we probe it
  // THROUGH the Go proxy: a successful KPI call means the AI is reachable.
  let aiOk = false;
  try {
    await ai.dashboardKpis();
    aiOk = true;
  } catch (e) {
    // A 502 from the proxy means the AI is down; any other answer means it
    // replied (so it's up). Only a network failure leaves aiOk false.
    aiOk = Boolean(e && e.status && e.status !== 502);
  }

  grid.innerHTML = [
    statusCard("API Go", apiOk),
    statusCard("Base de données", dbOk),
    statusCard("Service IA (Python, interne)", aiOk),
  ].join("");
}

// ----- Technical links: Swagger + /ping -----
function wireTechLinks() {
  const apiRoot = CONFIG.API_BASE.replace("/api/v1", "");
  document.getElementById("swagger-link").href = `${apiRoot}/swagger/index.html`;

  const btn = document.getElementById("ping-btn");
  const out = document.getElementById("ping-result");
  btn.addEventListener("click", async () => {
    out.textContent = "…";
    out.className = "text-sm self-center text-slate2";
    try {
      const r = await api.ping(); // GET /api/v1/ping -> { message: "pong" }
      out.textContent = `✔ ${r.message}`;
      out.className = "text-sm self-center text-green-700";
    } catch {
      out.textContent = "✖ API injoignable";
      out.className = "text-sm self-center text-hibiscus";
    }
  });
}

// ----- Collaborators -----
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
loadUsers();
loadMatrix();
loadMonitoring();
loadCollaborators();
wireTechLinks();
