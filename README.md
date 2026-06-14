# Ymmo — Plateforme immobilière

Plateforme web centralisée d'achat/vente de biens immobiliers pour le réseau
d'agences **Ymmo**, avec un module d'analyse de données et d'IA pour les
décisions stratégiques (tendances de prix, biens populaires, zones à cibler,
estimation de prix et prédiction du délai de vente).

Projet Fil Rouge B2 — Ynov Informatique (partie **DEV**).

## Stack

| Couche | Technologies |
|---|---|
| Front-end | HTML · Tailwind CSS · JavaScript (vanilla, modules ES) |
| API métier | Go (Gin) · GORM · JWT · Swagger/OpenAPI |
| Analyse / IA | Python · FastAPI · pandas · scikit-learn (lecture seule) |
| Base de données | MariaDB (schéma relationnel normalisé) |
| Conteneurisation | Docker · Docker Compose |

## Architecture

```
Navigateur ──► Front (nginx, :3000)
       │
       └──────► API Go (Gin, :8080) ──► MariaDB (:3306)
                     │
                     └─ proxy /api/v1/ai/* ──► Service IA Python (interne)
```

- L'**API Go** porte toute la logique métier et les écritures en base.
- Le **service IA Python** lit la *même* base en **lecture seule** (utilisateur
  SQL dédié) et fait l'analyse avec pandas / scikit-learn.
- Le service IA n'est **jamais exposé publiquement** : le navigateur passe
  toujours par l'API Go (`/api/v1/ai/*`), qui relaie en interne sur le réseau
  Docker.

Architecture back-end en couches (SOLID) : `handlers → services → repositories
→ models`, avec des DTO validés en entrée.

## Lancer tout le projet

Tout est conteneurisé. Depuis la racine :

```bash
cp .env.example .env      # renseigner les identifiants + JWT_SECRET
docker compose up --build
```

Cela démarre les 4 services : `ymmo-db`, `ymmo-api`, `ymmo-ai`, `ymmo-front`.

| Service | URL |
|---|---|
| Front-end | http://localhost:3000 |
| API Go | http://localhost:8080/api/v1 |
| Documentation API (Swagger) | http://localhost:8080/swagger/index.html |
| Health check API | http://localhost:8080/health |

> Le service IA n'a pas de port public : il est joignable uniquement via le
> proxy de l'API (`/api/v1/ai/*`).

Arrêter / réinitialiser :

```bash
docker compose down        # arrêt
docker compose down -v     # arrêt + suppression du volume base de données
```

## Jeu de données (seed)

Le `db/schema.sql` (tables + données de référence) est joué **automatiquement
au premier démarrage** sur un volume vierge. Les données de démonstration
(biens, comptes…) se chargent ensuite avec le script de seed :

```bash
docker cp db/seed.sql ymmo-db:/tmp/seed.sql
docker exec ymmo-db sh -c "mariadb -u root -p\"$MARIADB_ROOT_PASSWORD\" ymmo < /tmp/seed.sql"
```

### Comptes de test

| Email | Mot de passe | Rôle |
|---|---|---|
| `buyer@ymmo.fr` | `buyer1234` | Acheteur |
| `agent@ymmo.fr` | `agent1234` | Agent |
| `director@ymmo.fr` | `director1234` | Directeur d'agence |
| `hq@ymmo.fr` | `hq1234` | Siège |
| `it@ymmo.fr` | `it1234` | IT & Support |

Un acheteur peut demander à **devenir vendeur** (Profil → Paramètres) ; un agent
valide la demande (Dashboard → Demandes vendeur), ce qui le promeut vendeur.

## Développement (sans Docker)

### Base de données

```bash
mariadb -u root -p < db/schema.sql
```

### API Go

Prérequis : Go 1.22+.

```bash
cd backend
cp .env.example .env        # identifiants BDD + JWT_SECRET
go mod tidy
go run ./cmd/api            # http://localhost:8080
```

La doc Swagger est générée à partir des annotations des handlers. En Docker
c'est automatique ; en local :

```bash
go install github.com/swaggo/swag/cmd/swag@latest
cd backend && swag init -g cmd/api/main.go -o docs
```

### Service IA Python

Prérequis : Python 3.11–3.13.

```bash
cd ai-service
python -m venv .venv
source .venv/bin/activate         # Windows : .venv\Scripts\activate
pip install -r requirements.txt
cp .env.example .env
uvicorn app.main:app --reload --port 8000
```

### Front-end

Les pages appellent l'API : elles doivent être servies en HTTP (pas en
`file://`, rejeté par le CORS). Le CORS autorise `localhost:3000` et `:5173`.

```bash
cd frontend
python -m http.server 5173        # puis http://localhost:5173
```

## Charte graphique

| Token | Couleur | Usage |
|---|---|---|
| `midnight` | #2c3e50 | Texte / titres |
| `alyssum` | #efebe7 | Fond |
| `hibiscus` | #b63753 | Boutons principaux, navigation |
| `slate2` | #7f8c8d | Bordures, texte secondaire |
| `gold` | #d4af37 | Badges prestige, éléments IA |

## Accessibilité

Repères sémantiques (`header`/`main`/`footer`/`nav`), lien d'évitement,
`lang="fr"`, états de focus visibles et libellés descriptifs (WCAG / ARIA).

## Conventions

- Une branche par fonctionnalité, commits documentés, Pull Requests
  (pas de push direct sur `main`).
- `.env` n'est jamais commité (voir `.gitignore`).
