// Auth page behaviour: tab toggle, password visibility, and form submission.
import { api } from "./api.js";
import { login } from "./auth.js";

const tabLogin = document.getElementById("tab-login");
const tabRegister = document.getElementById("tab-register");
const loginForm = document.getElementById("login-form");
const registerForm = document.getElementById("register-form");

const ACTIVE = ["bg-hibiscus", "text-white"];
const INACTIVE = ["text-hibiscus"];

function showTab(tab) {
  const isLogin = tab === "login";
  loginForm.classList.toggle("hidden", !isLogin);
  registerForm.classList.toggle("hidden", isLogin);

  tabLogin.classList.toggle("bg-hibiscus", isLogin);
  tabLogin.classList.toggle("text-white", isLogin);
  tabLogin.classList.toggle("text-hibiscus", !isLogin);
  tabLogin.setAttribute("aria-selected", String(isLogin));

  tabRegister.classList.toggle("bg-hibiscus", !isLogin);
  tabRegister.classList.toggle("text-white", !isLogin);
  tabRegister.classList.toggle("text-hibiscus", isLogin);
  tabRegister.setAttribute("aria-selected", String(!isLogin));
}

tabLogin.addEventListener("click", () => showTab("login"));
tabRegister.addEventListener("click", () => showTab("register"));

if (window.location.hash === "#register") showTab("register");

document.querySelectorAll(".toggle-password").forEach((btn) => {
  btn.addEventListener("click", () => {
    const input = document.getElementById(btn.dataset.target);
    const show = input.type === "password";
    input.type = show ? "text" : "password";
    btn.setAttribute("aria-label", show ? "Masquer le mot de passe" : "Afficher le mot de passe");
  });
});

function showError(el, message) {
  el.textContent = message;
  el.classList.remove("hidden");
}
function hideError(el) {
  el.classList.add("hidden");
}
function busy(form, on, label) {
  const btn = form.querySelector("button[type=submit]");
  btn.disabled = on;
  btn.textContent = on ? "Veuillez patienter…" : label;
}

const loginError = loginForm.querySelector(".login-error");

loginForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  hideError(loginError);
  busy(loginForm, true);

  try {
    await login(loginForm.email.value.trim(), loginForm.password.value);
    window.location.href = "./index.html";
  } catch (err) {
    const msg =
      err.status === 401
        ? "Email ou mot de passe incorrect."
        : "Une erreur est survenue. Réessayez.";
    showError(loginError, msg);
    busy(loginForm, false, "Se connecter");
  }
});

const registerError = registerForm.querySelector(".register-error");

registerForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  hideError(registerError);

  // "NOM Prénom" -> last_name (first word) + first_name (the rest).
  const parts = registerForm.name.value.trim().split(/\s+/).filter(Boolean);
  const lastName = parts[0] || "";
  const firstName = parts.slice(1).join(" ");
  if (lastName.length < 2 || firstName.length < 2) {
    showError(registerError, "Indiquez votre nom ET votre prénom (ex : DOE John).");
    return;
  }

  const password = registerForm.password.value;
  if (password.length < 8) {
    showError(registerError, "Le mot de passe doit faire au moins 8 caractères.");
    return;
  }
  if (password !== registerForm.confirm.value) {
    showError(registerError, "Les deux mots de passe ne correspondent pas.");
    return;
  }

  const payload = {
    first_name: firstName,
    last_name: lastName,
    email: registerForm.email.value.trim(),
    password,
  };
  const phone = registerForm.phone.value.trim();
  if (phone) payload.phone = phone;

  busy(registerForm, true);
  try {
    await api.register(payload);
    // Auto-login right after a successful sign-up.
    await login(payload.email, password);
    window.location.href = "./index.html";
  } catch (err) {
    let msg = "Une erreur est survenue. Réessayez.";
    if (err.status === 409) msg = "Cet email est déjà utilisé.";
    else if (err.status === 400) msg = "Vérifiez les informations saisies (email, téléphone…).";
    showError(registerError, msg);
    busy(registerForm, false, "S'inscrire");
  }
});
