// Messaging page: list conversations, open one, read & send messages.
// (Real-time updates via WebSocket are planned for later — for now we load on
// open and append after sending.)
import { api } from "./api.js";
import { currentUser, isLoggedIn } from "./auth.js";

if (!isLoggedIn()) {
  window.location.href = "./auth.html";
}

const me = currentUser() || {};
let activeId = null;

const list = document.getElementById("conv-list");
const empty = document.getElementById("conv-empty");
const header = document.getElementById("conv-header");
const messagesEl = document.getElementById("messages");
const placeholder = document.getElementById("messages-placeholder");
const sendForm = document.getElementById("send-form");
const input = document.getElementById("message-input");

function escapeHtml(value) {
  return String(value ?? "").replace(/[&<>"']/g, (c) => (
    { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
  ));
}

// The other participant of a conversation, plus a role label.
function otherParty(conv) {
  const iAmClient = conv.client_id === me.id;
  const person = iAmClient ? conv.agent : conv.client;
  const role = iAmClient ? "agent" : "client";
  const name = person ? `${person.first_name} ${person.last_name}` : `Utilisateur #${iAmClient ? conv.agent_id : conv.client_id}`;
  return { name, role };
}

// The property a conversation is about (empty string if none).
function propertyLabel(conv) {
  return conv.property ? conv.property.title : "";
}

// ----- Conversation list -----
function renderList(conversations) {
  empty.classList.toggle("hidden", conversations.length > 0);
  list.innerHTML = conversations
    .map((conv) => {
      const { name, role } = otherParty(conv);
      const property = propertyLabel(conv);
      const active = conv.id === activeId;
      return `
        <li>
          <button data-id="${conv.id}"
            class="conv-item w-full text-left rounded-lg px-4 py-3 border transition-colors
                   ${active ? "bg-hibiscus text-white border-hibiscus" : "border-hibiscus text-hibiscus hover:bg-hibiscus/10"}">
            <span class="font-semibold">${escapeHtml(name)}</span> <span class="font-normal opacity-80">( ${role} )</span>
            ${property ? `<span class="block text-xs font-normal opacity-80 mt-0.5">${escapeHtml(property)}</span>` : ""}
          </button>
        </li>`;
    })
    .join("");
}

// ----- Messages -----
function bubble(message) {
  const mine = message.sender_id === me.id;
  const side = mine ? "ml-auto border-gold" : "mr-auto border-hibiscus";
  return `
    <div class="max-w-[75%] ${side} border rounded-lg px-4 py-3 bg-white">
      <p class="text-sm whitespace-pre-wrap break-words">${escapeHtml(message.body)}</p>
    </div>`;
}

async function openConversation(conv) {
  activeId = conv.id;

  const { name, role } = otherParty(conv);
  const property = propertyLabel(conv);
  header.classList.add("flex", "items-center", "justify-between", "gap-4");
  header.innerHTML = `
    <div>
      <span>${escapeHtml(name)} ( ${role} )</span>
      ${property ? `<span class="block text-sm font-normal text-slate2">${escapeHtml(property)}</span>` : ""}
    </div>
    <button id="conv-delete" class="shrink-0 text-sm font-medium border border-hibiscus text-hibiscus hover:bg-hibiscus hover:text-white px-3 py-1.5 rounded-md transition-colors">
      Supprimer
    </button>`;
  header.classList.remove("hidden");
  sendForm.classList.remove("hidden");
  placeholder.classList.add("hidden");

  // Delete this conversation (only a participant may; the API enforces it).
  document.getElementById("conv-delete").addEventListener("click", () => deleteConversation(conv));

  // Re-render the list to move the highlight.
  renderList(conversations);

  messagesEl.innerHTML = `<p class="text-slate2 text-center mt-10">Chargement…</p>`;
  try {
    const { data } = await api.getMessages(conv.id);
    messagesEl.innerHTML = data.length
      ? data.map(bubble).join("")
      : `<p class="text-slate2 text-center mt-10">Aucun message. Démarrez la discussion !</p>`;
    messagesEl.scrollTop = messagesEl.scrollHeight;
  } catch {
    messagesEl.innerHTML = `<p class="text-slate2 text-center mt-10">Impossible de charger les messages.</p>`;
  }
}

// ----- Delete a conversation -----
async function deleteConversation(conv) {
  if (!confirm("Supprimer cette conversation ?")) return;
  try {
    await api.deleteConversation(conv.id);
    conversations = conversations.filter((c) => c.id !== conv.id);
    activeId = null;
    renderList(conversations);
    if (conversations.length) {
      openConversation(conversations[0]);
    } else {
      header.classList.add("hidden");
      sendForm.classList.add("hidden");
      messagesEl.innerHTML = "";
      placeholder.classList.remove("hidden");
    }
  } catch {
    alert("Suppression impossible pour le moment.");
  }
}

// ----- Unread badge -----
async function refreshUnread() {
  const badge = document.getElementById("unread-badge");
  try {
    const { unread } = await api.unreadCount();
    if (unread > 0) {
      badge.textContent = `${unread} message${unread > 1 ? "s" : ""} non lu${unread > 1 ? "s" : ""}`;
      badge.classList.remove("hidden");
    } else {
      badge.classList.add("hidden");
    }
  } catch {
    /* ignore */
  }
}

// ----- Send -----
sendForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const body = input.value.trim();
  if (!body || !activeId) return;

  input.value = "";
  try {
    const message = await api.sendMessage(activeId, body);
    if (messagesEl.querySelector("p")) messagesEl.innerHTML = "";
    messagesEl.insertAdjacentHTML("beforeend", bubble(message));
    messagesEl.scrollTop = messagesEl.scrollHeight;
  } catch {
    input.value = body; // restore on failure
  }
});

// ----- Boot -----
let conversations = [];

list.addEventListener("click", (e) => {
  const btn = e.target.closest(".conv-item");
  if (!btn) return;
  const conv = conversations.find((c) => c.id === Number(btn.dataset.id));
  if (conv) openConversation(conv);
});

async function loadConversations() {
  refreshUnread();
  try {
    const { data } = await api.listConversations();
    conversations = data;
    renderList(conversations);
    if (conversations.length) openConversation(conversations[0]); // open the first by default
  } catch {
    list.innerHTML = `<li class="text-slate2">Impossible de charger les conversations.</li>`;
  }
}

loadConversations();
