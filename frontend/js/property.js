// Property detail page: fetch one property, render it, wire the gallery, the
// AI insight box, and the "contact an agent" action.
import { api, ai } from "./api.js";
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
  account.innerHTML = `<a href="./auth.html" class="text-sm font-medium text-midnight hover:text-hibiscus transition-colors">Connexion</a>`;
}

// ----- Helpers -----
const euro = new Intl.NumberFormat("fr-FR", { style: "currency", currency: "EUR", maximumFractionDigits: 0 });

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (c) => (
    { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
  ));
}

function photoUrls(property) {
  const photos = (property.photos || []).slice().sort((a, b) => a.sort_order - b.sort_order);
  const urls = photos.map((p) => p.url);
  return urls.length ? urls : ["https://placehold.co/800x600?text=Ymmo"];
}

function detailLine(property) {
  const parts = [property.city];
  if (property.rooms) parts.push(`${property.rooms} pièces`);
  if (property.area) parts.push(`${property.area} m²`);
  return parts.join(" - ");
}

// ----- Render -----
const detail = document.getElementById("detail");

function render(property) {
  const images = photoUrls(property);
  const exclusivity = property.is_exclusive
    ? `<p class="text-gold font-semibold tracking-wide">Exclusivité YMMO</p>`
    : "";

  const thumbs = images
    .slice(0, 4)
    .map(
      (url, i) => `
      <button class="thumb border ${i === 0 ? "border-hibiscus" : "border-hibiscus/30"} rounded-lg overflow-hidden h-24 focus:outline-none focus:ring-2 focus:ring-hibiscus/50" data-url="${escapeHtml(url)}" aria-label="Voir la photo ${i + 1}">
        <img src="${escapeHtml(url)}" alt="Photo ${i + 1} du bien" class="w-full h-full object-cover" />
      </button>`
    )
    .join("");

  detail.innerHTML = `
    <!-- Gallery -->
    <div>
      <img id="gallery-main" src="${escapeHtml(images[0])}" alt="${escapeHtml(property.title)}"
           class="w-full h-96 object-cover rounded-xl border border-hibiscus/30" />
      <div class="grid grid-cols-4 gap-3 mt-3">${thumbs}</div>
    </div>

    <!-- Info -->
    <div>
      ${exclusivity}
      <h1 class="text-3xl font-bold mt-1">${escapeHtml(property.title)}</h1>
      <p class="text-2xl font-bold text-hibiscus mt-3">${euro.format(property.price)}</p>

      <hr class="my-5 border-slate2/30" />

      <p class="font-semibold">${escapeHtml(detailLine(property))}</p>
      <p class="text-slate2 mt-2 leading-relaxed">${escapeHtml(property.description || "")}</p>

      <!-- AI strategic insight -->
      <div id="ai-box" class="mt-6 border border-hibiscus/40 rounded-lg p-4 bg-white">
        <p class="flex items-center gap-2 text-hibiscus font-semibold">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 18h6M10 22h4M12 2a7 7 0 0 0-4 12c.7.7 1 1.5 1 2h6c0-.5.3-1.3 1-2A7 7 0 0 0 12 2z"/></svg>
          Information stratégique
        </p>
        <p id="ai-content" class="text-sm text-slate2 mt-2">Analyse en cours…</p>
      </div>

      <!-- Contact -->
      <button id="contact-btn" class="mt-6 w-full bg-hibiscus hover:bg-hibiscus/90 text-white font-medium py-3 rounded-lg transition-colors">
        Contacter un agent
      </button>
      <p id="contact-msg" class="hidden mt-3 text-sm text-center" role="status"></p>

      <!-- Favorite -->
      <button id="fav-btn" aria-pressed="false"
        class="mt-3 w-full border border-hibiscus text-hibiscus hover:bg-hibiscus/10 font-medium py-2.5 rounded-lg transition-colors flex items-center justify-center gap-2">
        ${heartIcon(false)}<span class="fav-label">Ajouter aux favoris</span>
      </button>
    </div>`;

  wireGallery();
  loadAiInsight(property);
  wireContact(property);
  wireFavorite(property);
}

// Favorite toggle on the detail page.
async function wireFavorite(property) {
  const btn = document.getElementById("fav-btn");
  const setBtn = (fav) => {
    btn.setAttribute("aria-pressed", String(fav));
    btn.querySelector(".fav-label").textContent = fav ? "Retirer des favoris" : "Ajouter aux favoris";
    btn.querySelector("svg")?.remove();
    btn.insertAdjacentHTML("afterbegin", heartIcon(fav));
  };

  let fav = false;
  if (isLoggedIn()) {
    fav = (await favoriteIdSet()).has(property.id);
    setBtn(fav);
  }

  btn.addEventListener("click", async () => {
    if (!isLoggedIn()) {
      window.location.href = "./auth.html";
      return;
    }
    btn.disabled = true;
    try {
      fav = await toggleFavorite(property.id, fav);
      setBtn(fav);
    } catch {
      /* keep previous state */
    }
    btn.disabled = false;
  });
}

// Thumbnail click swaps the main image.
function wireGallery() {
  const main = document.getElementById("gallery-main");
  document.querySelectorAll(".thumb").forEach((btn) => {
    btn.addEventListener("click", () => {
      main.src = btn.dataset.url;
      document.querySelectorAll(".thumb").forEach((b) => {
        b.classList.toggle("border-hibiscus", b === btn);
        b.classList.toggle("border-hibiscus/30", b !== btn);
      });
    });
  });
}

// AI insight: compare the listing price with our model's estimate.
async function loadAiInsight(property) {
  const box = document.getElementById("ai-content");
  try {
    const est = await ai.estimate({
      city: property.city,
      category_id: property.category_id,
      area: Number(property.area),
    });
    const diff = Math.round(((property.price - est.estimated_price) / est.estimated_price) * 100);
    const sign = diff > 0 ? "+" : "";
    box.innerHTML = `Estimation de notre IA : <strong>${euro.format(est.estimated_price)}</strong>.
      Ce bien est affiché <strong>${sign}${diff}%</strong> par rapport à l'estimation
      (modèle : ${est.method === "linear_regression" ? "régression" : "prix/m² moyen"}, ${est.sample_size} biens comparés).`;
  } catch (err) {
    box.textContent = "Analyse IA indisponible pour le moment.";
  }
}

// Contact an agent: start a conversation about this property.
function wireContact(property) {
  const btn = document.getElementById("contact-btn");
  const msg = document.getElementById("contact-msg");

  if (!property.agent_id) {
    btn.disabled = true;
    btn.classList.add("opacity-60", "cursor-not-allowed");
    btn.textContent = "Aucun agent assigné";
    return;
  }

  btn.addEventListener("click", async () => {
    if (!isLoggedIn()) {
      window.location.href = "./auth.html";
      return;
    }
    btn.disabled = true;
    try {
      await api.startConversation({ agent_id: property.agent_id, property_id: property.id });
      msg.textContent = "Conversation démarrée — retrouvez-la dans votre messagerie.";
      msg.classList.remove("hidden");
      msg.classList.add("text-hibiscus");
    } catch (err) {
      msg.textContent = "Impossible de contacter l'agent pour le moment.";
      msg.classList.remove("hidden");
      msg.classList.add("text-hibiscus");
      btn.disabled = false;
    }
  });
}

// ----- Boot -----
const id = new URLSearchParams(window.location.search).get("id");
if (!id) {
  detail.innerHTML = `<p class="text-slate2">Bien introuvable.</p>`;
} else {
  api
    .getProperty(id)
    .then(render)
    .catch((err) => {
      detail.innerHTML =
        err.status === 404
          ? `<p class="text-slate2">Ce bien n'existe pas ou n'est plus disponible.</p>`
          : `<p class="text-slate2">Impossible de charger le bien. L'API est-elle démarrée ?</p>`;
    });
}
