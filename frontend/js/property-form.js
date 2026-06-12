// Create or edit a property. ?id= present -> edit mode (PUT), otherwise create.
import { api } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

// Staff only.
const STAFF = ["AGENT", "DIRECTOR", "HQ"];
const me = currentUser() || {};
if (!isLoggedIn() || !STAFF.includes(me.role)) {
  window.location.href = "./index.html";
}

// Reference data (fixed seed values). Could come from API endpoints later.
const CATEGORIES = [
  { id: 1, label: "Maison" }, { id: 2, label: "Appartement" }, { id: 3, label: "Studio" },
  { id: 4, label: "Terrain" }, { id: 5, label: "Bureau" }, { id: 6, label: "Local commercial" }, { id: 7, label: "Entrepôt" },
];
const AGENCIES = [
  { id: 1, name: "Ymmo Siège (Aix-en-Provence)" }, { id: 2, name: "Ymmo Paris" },
  { id: 3, name: "Ymmo Lyon" }, { id: 4, name: "Ymmo Marseille" },
];
const ENERGY = ["A", "B", "C", "D", "E", "F", "G"];
const STATUSES = [
  { v: "DRAFT", l: "Brouillon" }, { v: "PENDING_REVIEW", l: "À valider" }, { v: "AVAILABLE", l: "Disponible" },
  { v: "UNDER_OFFER", l: "Sous offre" }, { v: "SOLD", l: "Vendu" }, { v: "WITHDRAWN", l: "Retiré" },
];

const $ = (id) => document.getElementById(id);

// Populate the selects.
$("category_id").innerHTML = CATEGORIES.map((c) => `<option value="${c.id}">${c.label}</option>`).join("");
$("agency_id").innerHTML = AGENCIES.map((a) => `<option value="${a.id}">${a.name}</option>`).join("");
$("energy_rating").innerHTML = `<option value="">—</option>` + ENERGY.map((e) => `<option value="${e}">${e}</option>`).join("");
$("ghg_rating").innerHTML = `<option value="">—</option>` + ENERGY.map((e) => `<option value="${e}">${e}</option>`).join("");
$("status").innerHTML = STATUSES.map((s) => `<option value="${s.v}">${s.l}</option>`).join("");

const editId = new URLSearchParams(window.location.search).get("id");
const form = $("property-form");
const errorBox = $("form-error");

// ---------- Edit mode: prefill ----------
if (editId) {
  $("form-title").textContent = "Modifier le bien";
  $("submit-btn").textContent = "Enregistrer";
  $("status-wrap").classList.remove("hidden");
  $("delete-btn").classList.remove("hidden");
  $("agency-wrap").classList.add("hidden"); // agency isn't editable via the API

  api.getProperty(editId).then((p) => {
    $("title").value = p.title || "";
    $("description").value = p.description || "";
    $("category_id").value = p.category_id;
    $("status").value = p.status;
    $("price").value = p.price;
    $("area").value = p.area;
    ["rooms", "bedrooms", "bathrooms", "floor", "build_year"].forEach((k) => { if (p[k] != null) $(k).value = p[k]; });
    $("energy_rating").value = p.energy_rating || "";
    $("ghg_rating").value = p.ghg_rating || "";
    $("address").value = p.address || "";
    $("city").value = p.city || "";
    $("postal_code").value = p.postal_code || "";
    $("is_exclusive").checked = !!p.is_exclusive;
  }).catch(() => showError("Impossible de charger ce bien."));

  $("delete-btn").addEventListener("click", async () => {
    if (!confirm("Supprimer définitivement ce bien ?")) return;
    try {
      await api.deleteProperty(editId);
      window.location.href = "./dashboard.html";
    } catch {
      showError("Suppression impossible.");
    }
  });
}

// ---------- Helpers ----------
function showError(msg) {
  errorBox.textContent = msg;
  errorBox.classList.remove("hidden");
}
function num(id) {
  const v = $(id).value;
  return v === "" ? undefined : Number(v);
}
function str(id) {
  const v = $(id).value.trim();
  return v === "" ? undefined : v;
}

// ---------- Submit ----------
form.addEventListener("submit", async (e) => {
  e.preventDefault();
  errorBox.classList.add("hidden");

  const payload = {
    title: $("title").value.trim(),
    category_id: Number($("category_id").value),
    price: Number($("price").value),
    area: Number($("area").value),
    city: $("city").value.trim(),
    postal_code: $("postal_code").value.trim(),
    is_exclusive: $("is_exclusive").checked,
  };
  // Optional fields (only sent when filled).
  for (const k of ["rooms", "bedrooms", "bathrooms", "floor", "build_year"]) {
    const v = num(k);
    if (v !== undefined) payload[k] = v;
  }
  for (const k of ["description", "address"]) {
    const v = str(k);
    if (v !== undefined) payload[k] = v;
  }
  if ($("energy_rating").value) payload.energy_rating = $("energy_rating").value;
  if ($("ghg_rating").value) payload.ghg_rating = $("ghg_rating").value;

  const btn = $("submit-btn");
  btn.disabled = true;

  try {
    if (editId) {
      payload.status = $("status").value;
      await api.updateProperty(editId, payload);
    } else {
      payload.agency_id = Number($("agency_id").value);
      await api.createProperty(payload);
    }
    window.location.href = "./dashboard.html";
  } catch (err) {
    showError(err.status === 400 ? "Vérifiez les champs (prix, surface, ville…)." : "Une erreur est survenue.");
    btn.disabled = false;
  }
});
