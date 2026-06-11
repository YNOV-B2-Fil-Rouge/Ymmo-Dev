// Small page script: sets the year and reflects the auth state in the nav.
import { currentUser, logout } from "./auth.js";

// Footer year.
document.getElementById("year").textContent = new Date().getFullYear();

// Swap the "Connexion" button for the user's name + logout when logged in.
const account = document.getElementById("nav-account");
const user = currentUser();

if (user) {
  account.innerHTML = `
    <span class="text-sm text-white/80">Bonjour, ${user.first_name}</span>
    <button id="logout-btn"
      class="border border-white/30 hover:border-white text-white text-sm px-3 py-1.5 rounded-md transition-colors">
      Déconnexion
    </button>`;
  document.getElementById("logout-btn").addEventListener("click", () => {
    logout();
    window.location.reload();
  });
}
