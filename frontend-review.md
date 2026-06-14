# Revue de code — PR Front-End (`feature/frontend`)

Périmètre : 33 fichiers, +3 715 lignes (100 % ajouts). Front statique HTML + Tailwind (CDN) + JavaScript vanilla (ES modules), servi par nginx. Revue du dossier `frontend/`.

## Vue d'ensemble
La PR introduit l'intégralité du front : 10 pages HTML (accueil, recherche, fiche bien, auth, profil, messagerie, formulaire d'annonce, 3 dashboards), un client API centralisé (`api.js`), des helpers partagés (favoris, accessibilité), la charte graphique en tokens Tailwind, et la conteneurisation nginx. L'architecture est propre et cohérente : un seul point d'entrée réseau (`api.js`), séparation claire par page, échappement HTML systématique.

## Points forts
- **Client API unique** (`api.js`) : tous les appels passent par `request()` / `requestForm()`, avec gestion JWT et erreurs homogène (`throw { status, data }`).
- **Sécurité XSS bien traitée** : `escapeHtml` appliqué partout où des données serveur/utilisateur sont injectées (cartes, messagerie y compris le corps des messages, tableaux dashboards).
- **Accessibilité soignée** : skip links, `aria-pressed` sur les favoris, régions `role="status"` live, interface à onglets ARIA (`a11y.js`), `alt` pertinents, `lang="fr"`, contrastes AA.
- **UX** : squelettes de chargement, délégation d'événements (favoris), messages d'erreur clairs quand l'API est down.
- **Aucun secret commité**, tokens de charte centralisés (`tailwind-config.js`), PR purement additive → risque de régression faible sur les autres modules.

## Problèmes à corriger

### Bloquant déploiement
- **`config.js` — `API_BASE: "http://localhost:8080/api/v1"` est figé dans l'image nginx.** Le front buildé pointera toujours vers `localhost:8080`, donc il ne fonctionne que sur la machine de dev. En prod (ou pour un correcteur sur une autre machine), tous les appels échouent. **Recommandation** : utiliser un chemin relatif (`/api/v1`) et faire un reverse-proxy nginx `location /api/ { proxy_pass http://api:8080; }`, ou injecter l'URL au build.

### Sécurité
- **JWT stocké en `localStorage`** : exposé à tout XSS (vol de token = prise de compte totale). L'échappement systématique limite fortement le risque, mais à mentionner. Une alternative plus robuste serait un cookie `HttpOnly` (changement d'archi plus lourd).
- **En-têtes de sécurité absents dans `nginx.conf`** : pas de `Content-Security-Policy`, `X-Content-Type-Options: nosniff`, `X-Frame-Options`, `Referrer-Policy`. À ajouter (surtout avec Tailwind chargé depuis un CDN).
- **Confiance au `role` de `localStorage`** : la navigation/affichage se base sur `currentUser().role`. C'est OK pour l'UI (l'accès réel est protégé côté serveur par le JWT), mais à garder en tête : ne jamais en faire une protection. ✅ Le serveur reste l'autorité.

### Correctness / incohérences
- **`property-form.js` — liste `AGENCIES` codée en dur à 4 agences** alors que la base en compte 5 (Bordeaux ajoutée). Conséquence : impossible de créer une annonce rattachée à l'agence Bordeaux. **Recommandation** : charger les agences depuis l'API plutôt que de les coder en dur (même remarque pour `CATEGORIES`/`STATUSES`, risque de dérive).
- **`config.js` — `AI_BASE: "http://localhost:8000"` mort et trompeur** : le service IA est interne (jamais exposé), tout passe par le proxy Go (`/api/v1/ai/*`). Cette constante n'est pas utilisée et contredit l'archi documentée → à supprimer.
- **Gestion du 401 incohérente** : seul `profile.js` déconnecte et redirige sur expiration de session. Les autres pages (dashboard, messagerie…) affichent une erreur générique. **Recommandation** : intercepter le 401 dans `request()` (clear token + redirection vers `auth.html`) pour un comportement global.

### Maintenabilité / nits
- **`escapeHtml` dupliqué** dans ~4 fichiers (home, search, dashboard-it, messages…). À extraire dans un `dom.js` partagé (comme `a11y.js`).
- **Logique d'en-tête « nav-account » dupliquée** dans `home.js`, `property.js`, `search.js`. À factoriser (`renderNav()`).
- **Onglets ARIA sans navigation clavier fléchée** : `a11y.js` pose bien les rôles `tablist/tab/tabpanel`, mais la spec WAI-ARIA attend la navigation Flèches gauche/droite entre onglets. À compléter pour une conformité complète.
- **Tailwind via CDN (compilation runtime)** : acceptable pour un projet école, mais en prod → pas de purge, FOUC possible, dépendance CDN. À noter.
- **`nginx.conf` met en cache le JS 1h sans fingerprint** → risque de JS périmé après déploiement. Prévoir un cache-busting (querystring de version) ou un cache plus court sur le HTML.
- **Aucun test automatisé JS** : acceptable pour le périmètre, mais à signaler.

## Verdict
PR solide et cohérente pour le périmètre du projet : bonne archi, accessibilité au-dessus de la moyenne, sécurité XSS gérée. Deux choses à traiter en priorité avant de parler de « prod » : **l'URL d'API figée sur localhost** (bloquant hors machine de dev) et **les en-têtes de sécurité nginx**. Le reste (agences en dur, 401 global, factorisation) est de la dette mineure, traitable au fil de l'eau.
