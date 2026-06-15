# Ymmo Front-end

## 📖 Table of Contents
1. [🤔 Module Overview](#-module-overview)
2. [🛠️ Prerequisites](#-prerequisites)
3. [🚀 Installation and Startup](#-installation-and-startup)
   1. [🌍 With Docker (recommended)](#-with-docker-recommended)
   2. [🧰 Dev Server](#-dev-server)
4. [🔐 Configuration](#-configuration)
5. [🐳 Docker Architecture](#-docker-architecture)
6. [🎨 Design & Accessibility](#-design--accessibility)
7. [🗂️ Module Tree](#-module-tree)
8. [🌐 Usage Notes](#-usage-notes)
9. [👥 Authors](#-authors)
10. [🪢 Appendix](#-appendix)

## 🤔 Module Overview
This is the **web interface** of the Ymmo platform: a static front-end in **HTML**, **Tailwind CSS** and **vanilla JavaScript (ES modules)**, served by **nginx** (`nginx:1.27-alpine`).

It is intentionally independent from the back-end:
- it only serves static files;
- it talks to the Go API through `fetch` (`js/api.js` centralizes all calls);
- the API base URL is configured in `js/config.js` (`http://localhost:8080/api/v1`);
- AI features are called through the API proxy, never the Python service directly.

Pages: home/catalogue, search, property detail, auth, profile, messaging, and role-based dashboards (agent, director/HQ, IT).

## 🛠️ Prerequisites
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff)](https://www.docker.com/products/)

[![Python](https://img.shields.io/badge/Python-3776AB?logo=python&logoColor=fff)](https://www.python.org/downloads/) *(dev server only)*

## 🚀 Installation and Startup

### 🌍 With Docker (recommended)
Part of the project's `docker-compose.yml`. From the **project root**:
```bash
docker compose up -d --build frontend
```
The interface is served on **http://localhost**.

> ⚠️ nginx serves the files frozen at build time. After changing a page, rebuild with `docker compose up -d --build frontend` and hard-refresh (Ctrl+Shift+R).

### 🧰 Dev Server
For live editing, serve the folder over HTTP (not `file://`, which the API's CORS rejects). The CORS allows `localhost:5173`:
```bash
cd frontend
python -m http.server 5173       # http://localhost:5173
```
Make sure the API is running.

## 🔐 Configuration
The front-end uses **no Docker `.env`** (it is a static site). Its only configuration is `js/config.js`:
```js
export const CONFIG = {
  API_BASE: "http://localhost:8080/api/v1",
  AI_BASE: "http://localhost:8000",
};
```
The browser hits the Go API directly on the published port `8080`.

## 🐳 Docker Architecture
- Base image **`nginx:1.27-alpine`** (lightweight, reduced attack surface).
- Build copies only the `*.html` pages and the `js/` folder into the web root; the custom `nginx.conf` handles multi-page routing.
- Published on host port **80** (`80:80`).

## 🎨 Design & Accessibility
- **Charte graphique** tokens (shared Tailwind config in `js/tailwind-config.js`): `midnight #2c3e50`, `alyssum #efebe7`, `hibiscus #b63753`, `slate2`, `gold`.
- **Accessibility (WCAG / ARIA)**: `lang="fr"`, skip links, semantic landmarks, image `alt`, ARIA tab interfaces on dashboards, `aria-pressed` toggles, live region on search results, AA color contrast.

## 🗂️ Module Tree
```text
frontend/
|-- Dockerfile
|-- nginx.conf
|-- *.html              # index, auth, property, profile, messages, recherche, dashboards...
`-- js/
    |-- config.js  api.js  auth.js  a11y.js  favorites.js  tailwind-config.js
    `-- home.js  search.js  property.js  profile.js  messages.js  property-form.js  dashboard*.js
```

## 🌐 Usage Notes
- Designed to work with the Ymmo back-end; in isolation the pages render but API-dependent features won't work.
- All network calls go through `js/api.js` (base URL `/api/v1`).

## 👥 Authors
- **LEFEBVRE Nino** — DEV
- **AMIARD Renaud** — INFRA
- **LASBENNES Lucas** — INFRA

## 🪢 Appendix
- 🌍 [Global README](../README.md)
- 📦 [Back-end (Go API)](../backend/README.md)
- 🧠 [Data/AI service](../ai-service/README.md)
