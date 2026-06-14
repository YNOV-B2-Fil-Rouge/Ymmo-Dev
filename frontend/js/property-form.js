// Create or edit a property. ?id= present -> edit mode (PUT), otherwise create.
import { api } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

// Access: staff manage listings; a SELLER may submit a new one (create only).
const STAFF = ["AGENT", "DIRECTOR", "HQ"];
const me = currentUser() || {};
const isStaff = STAFF.includes(me.role);
const isSeller = me.role === "SELLER";
const editId = new URLSearchParams(window.location.search).get("id");

if (!isLoggedIn() || (!isStaff && !isSeller)) {
  window.location.href = "./index.html";
}
if (isSeller && editId) {
  window.location.href = "./profile.html"; // sellers can only create
}
const afterSubmitHref = isSeller ? "./profile.html" : "./dashboard.html";

// Reference data (fixed seed values).
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

$("category_id").innerHTML = CATEGORIES.map((c) => `<option value="${c.id}">${c.label}</option>`).join("");
$("agency_id").innerHTML = AGENCIES.map((a) => `<option value="${a.id}">${a.name}</option>`).join("");
$("energy_rating").innerHTML = `<option value="">—</option>` + ENERGY.map((e) => `<option value="${e}">${e}</option>`).join("");
$("ghg_rating").innerHTML = `<option value="">—</option>` + ENERGY.map((e) => `<option value="${e}">${e}</option>`).join("");
$("status").innerHTML = STATUSES.map((s) => `<option value="${s.v}">${s.l}</option>`).join("");

const form = $("property-form");
const errorBox = $("form-error");

// For a seller, reframe as a submission and hide the staff-only photo picker.
if (isSeller) {
  $("form-title").textContent = "Proposer un bien à la vente";
  $("submit-btn").textContent = "Soumettre pour validation";
  $("create-photos").classList.add("hidden");
}

// Edit mode: prefill the form.
if (editId) {
  $("form-title").textContent = "Modifier le bien";
  $("submit-btn").textContent = "Enregistrer";
  $("status-wrap").classList.remove("hidden");
  $("delete-btn").classList.remove("hidden");
  $("agency-wrap").classList.add("hidden"); // agency isn't editable via the API
  $("create-photos").classList.add("hidden"); // edit mode uses the photos manager below

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
    renderPhotos(p.photos || []);
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

  $("photos-section").classList.remove("hidden");

  $("photo-add").addEventListener("click", async () => {
    const url = $("photo-url").value.trim();
    const err = $("photo-error");
    err.classList.add("hidden");
    if (!url) {
      err.textContent = "Renseignez une URL d'image.";
      err.classList.remove("hidden");
      return;
    }
    try {
      await api.addPhoto(editId, { url, is_primary: $("photo-primary").checked });
      $("photo-url").value = "";
      $("photo-primary").checked = false;
      await reloadPhotos();
    } catch (e) {
      err.textContent = e.status === 400 ? "URL invalide (max 255 caractères)." : "Ajout impossible.";
      err.classList.remove("hidden");
    }
  });

  $("photo-upload").addEventListener("click", async () => {
    const input = $("photo-file");
    const err = $("photo-error");
    err.classList.add("hidden");
    const file = input.files && input.files[0];
    if (!file) {
      err.textContent = "Choisissez un fichier image.";
      err.classList.remove("hidden");
      return;
    }

    const fd = new FormData();
    fd.append("file", file);
    if ($("photo-file-primary").checked) fd.append("is_primary", "true");

    const btn = $("photo-upload");
    btn.disabled = true;
    try {
      await api.uploadPhoto(editId, fd);
      input.value = "";
      $("photo-file-primary").checked = false;
      await reloadPhotos();
    } catch (e) {
      err.textContent =
        e.status === 400 ? "Fichier invalide (image jpg/png/webp/gif, max 5 Mo)."
        : "Téléversement impossible.";
      err.classList.remove("hidden");
    }
    btn.disabled = false;
  });
}

async function reloadPhotos() {
  try {
    const p = await api.getProperty(editId);
    renderPhotos(p.photos || []);
  } catch {}
}

function renderPhotos(photos) {
  const grid = $("photos-grid");
  const empty = $("photos-empty");
  const sorted = photos.slice().sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
  empty.classList.toggle("hidden", sorted.length > 0);
  grid.innerHTML = sorted
    .map(
      (ph) => `
      <figure class="relative rounded-lg overflow-hidden border border-hibiscus/30 group">
        <img src="${ph.url}" alt="Photo du bien" class="w-full h-28 object-cover" />
        ${ph.is_primary ? `<figcaption class="absolute top-1 left-1 text-xs bg-gold text-midnight px-2 py-0.5 rounded-full">Principale</figcaption>` : ""}
        <button type="button" data-photo="${ph.id}"
          class="photo-del absolute top-1 right-1 w-7 h-7 rounded-full bg-white/90 text-hibiscus flex items-center justify-center shadow hover:bg-hibiscus hover:text-white transition-colors"
          aria-label="Supprimer la photo">×</button>
      </figure>`
    )
    .join("");

  grid.querySelectorAll(".photo-del").forEach((btn) =>
    btn.addEventListener("click", async () => {
      if (!confirm("Supprimer cette photo ?")) return;
      try {
        await api.deletePhoto(editId, btn.dataset.photo);
        await reloadPhotos();
      } catch {
        showError("Suppression de la photo impossible.");
      }
    })
  );
}

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
  // No negative values anywhere (price/area must be > 0, counts/floor/year >= 0).
  if (payload.price <= 0 || payload.area <= 0) {
    showError("Le prix et la surface doivent être strictement positifs.");
    return;
  }
  // Optional fields (only sent when filled).
  for (const k of ["rooms", "bedrooms", "bathrooms", "floor", "build_year"]) {
    const v = num(k);
    if (v !== undefined) {
      if (v < 0) {
        showError("Les valeurs numériques ne peuvent pas être négatives.");
        return;
      }
      payload[k] = v;
    }
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
      window.location.href = afterSubmitHref;
    } else {
      payload.agency_id = Number($("agency_id").value);
      const created = await api.createProperty(payload); // seller -> PENDING_REVIEW, staff -> DRAFT

      // Upload the staged photos to the new property (first one = primary).
      const files = $("create-photo-files").files;
      for (let i = 0; i < files.length; i++) {
        const fd = new FormData();
        fd.append("file", files[i]);
        if (i === 0) fd.append("is_primary", "true");
        try {
          await api.uploadPhoto(created.id, fd);
        } catch {}
      }

      // Staff land on the new property's edit page; sellers go back to profile.
      window.location.href = isSeller ? "./profile.html" : `./property-form.html?id=${created.id}`;
    }
  } catch (err) {
    showError(err.status === 400 ? "Vérifiez les champs (prix, surface, ville…)." : "Une erreur est survenue.");
    btn.disabled = false;
  }
});
