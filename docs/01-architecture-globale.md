# Ymmo — Architecture globale (Module 0)

> Document technique — Plateforme web immobilière Ymmo
> Version 0.1 · Auteur : équipe DEV · Dernière mise à jour : voir Git

## 1. Objectif du document

Ce document pose l'architecture cible de la plateforme **avant** d'écrire la moindre ligne de code métier. Il sert deux buts :

1. **Cadrer le développement** : tout le monde dans l'équipe code contre la même cible.
2. **Préparer l'oral** : chaque choix technique est justifié ici, pour pouvoir le défendre devant le jury (critère « Concevoir une solution logicielle », coef. 5).

## 2. Vue d'ensemble

On retient une architecture **orientée services**, comme l'exige explicitement le sujet (« architecture backend orientée services »). Concrètement, le système se décompose en quatre blocs indépendants qui communiquent par HTTP/JSON :

```
┌─────────────────────────────────────────────────────────────────────┐
│                          NAVIGATEUR (Client)                          │
│   HTML + Tailwind CSS + JavaScript  ·  Responsive  ·  WCAG/ARIA        │
└───────────────────────────────┬───────────────────────────────────────┘
                                 │  HTTPS (JSON)
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      API REST — Go (Golang)                           │
│   Routeur HTTP + GORM (ORM)  ·  Auth JWT  ·  RBAC (matrice des droits) │
│   Validation des entrées  ·  Documentation Swagger / OpenAPI          │
└──────────────┬───────────────────────────────────┬────────────────────┘
               │ SQL                                │ HTTP (interne)
               ▼                                    ▼
┌──────────────────────────────┐     ┌──────────────────────────────────┐
│     BASE DE DONNÉES           │     │   MICROSERVICE DATA / IA — Python │
│         MariaDB               │◄────│   FastAPI + pandas/scikit-learn   │
│   Schéma relationnel 3NF      │     │   Tendances, biens populaires,    │
│                               │     │   prévisions de vente, zones      │
└──────────────────────────────┘     └──────────────────────────────────┘
```

### Pourquoi ce découpage en 4 blocs ?

- **Séparation logique métier / présentation** : le front ne connaît que l'API, jamais la base. C'est ce que demande le sujet (« séparation entre la logique métier et l'interface utilisateur »).
- **Le module Data/IA est isolé en Python** : voir §4. C'est le point d'architecture le plus important à comprendre pour l'oral.

## 3. Stack technique retenue et justification

| Couche | Technologie | Justification (à dire à l'oral) |
|---|---|---|
| Front-end | HTML + **Tailwind CSS** + JavaScript | Tailwind = design system cohérent et rapide, classes utilitaires → facile de respecter la charte graphique et le responsive (desktop/tablette/mobile exigé). |
| Back-end API | **Go (Golang)** | Langage compilé, très performant, fortement typé → robustesse et sécurité. Excellent pour une API REST. |
| ORM | **GORM** | ORM Go le plus mature et le plus utilisé. Migrations automatiques, relations, hooks, protection native contre l'injection SQL (requêtes paramétrées). |
| Base de données | **MariaDB** | SGBD relationnel exigé (« SQL avancé »). MariaDB = compatible MySQL, open-source, supporte les contraintes, index, vues et procédures stockées dont on aura besoin. |
| Module Data/IA | **Python** (FastAPI + pandas + scikit-learn) | **Le sujet impose Python pour l'analyse de données.** On l'expose en microservice que l'API Go interroge. |
| Documentation API | **Swagger / OpenAPI** | Génère une doc interactive de l'API → critère « documentation technique » + démonstration à l'oral. |
| Versioning | **Git + GitHub** | Branches par fonctionnalité, commits documentés, Pull Requests (critère transverse coef. 3). |

### ⚠️ Point de vigilance à arbitrer avec votre intervenant

La grille de notation liste les langages attendus côté DEV : **Python, Java, C#, PHP** (et « Javascript » est cité dans la description du critère). **Go n'y figure pas explicitement.** Deux options :

- **Option A (sûre)** : confirmer auprès de l'intervenant que Go est accepté. S'il valide, l'architecture ci-dessus est idéale.
- **Option B (zéro risque)** : remplacer le back-end Go par **PHP (Laravel)** ou **Python (FastAPI/Django)**, qui sont dans la liste. Le reste de l'architecture (MariaDB, microservice data Python, front Tailwind, Swagger) ne bouge pas.

Mon conseil de mentor : posez la question **avant** d'investir dans le back-end. Le critère « Développer une application fonctionnelle » pèse coef. 10 — c'est le plus lourd de la partie DEV, il ne faut pas qu'un choix de langage non validé vous coûte des points. Pour l'instant, je continue sur Go comme demandé.

## 4. Le point clé : pourquoi un microservice Python pour la Data/IA ?

C'est le choix d'architecture qui rapporte le plus de points et qu'il faut savoir défendre.

Le sujet impose **deux choses contradictoires en apparence** :
- un back-end applicatif (où vous avez choisi Go) ;
- de l'**analyse de données en Python** (« Analyse et manipulation de données en Python », critère explicite).

La solution propre : on ne mélange pas les deux dans le même service. L'API Go gère le métier transactionnel (utilisateurs, biens, messagerie, visites…). Un **microservice Python séparé** gère le calcul analytique (tendances de prix, biens populaires, prévisions de vente, zones stratégiques). L'API Go l'appelle en HTTP quand le front demande un tableau de bord IA.

Avantages à mettre en avant :
- On respecte **les deux** exigences du sujet sans compromis.
- On applique **SOLID** (responsabilité unique) et **séparation des préoccupations** à l'échelle de l'architecture, pas seulement du code.
- Python apporte `pandas`, `scikit-learn`, `matplotlib` → l'écosystème data que Go n'a pas.

## 5. Principes de qualité appliqués (SOLID / DRY / KISS)

Ces principes sont notés (coef. 3+3). On les matérialise par l'organisation du code Go :

- **SOLID** : architecture en couches `handlers → services → repositories → models`. Chaque couche a une responsabilité unique (S), on dépend d'interfaces et non d'implémentations (D).
- **DRY** : la validation, la gestion d'erreurs et la réponse JSON sont centralisées dans des middlewares/helpers réutilisables.
- **KISS** : pas de sur-ingénierie. Un endpoint = une action claire.

Arborescence cible du back-end Go :

```
/backend
  /cmd/api          → point d'entrée (main.go)
  /internal
    /config         → configuration, connexion BDD
    /models         → structs GORM (entités métier)
    /repositories   → accès aux données (requêtes SQL/GORM)
    /services       → logique métier
    /handlers       → contrôleurs HTTP (routes)
    /middleware     → auth JWT, RBAC, logs, CORS, validation
    /dto            → objets de transfert (entrées/sorties validées)
  /docs             → fichiers Swagger générés
  /migrations       → migrations SQL
```

## 6. Sécurité (transverse à toutes les couches)

| Exigence du sujet | Mise en œuvre côté application |
|---|---|
| Authentification | JWT signé, mot de passe hashé (bcrypt). |
| Matrice des droits / pôles | **RBAC** : chaque utilisateur a un rôle + un pôle. Un middleware vérifie le droit (Interdit / Lecture / Lecture-Écriture) avant chaque action sensible. La matrice du sujet est traduite en table `permissions_pole` (voir schéma BDD). |
| Validation des entrées | DTO validés à l'entrée de chaque endpoint (taille, type, format email, bornes prix/surface…). |
| Injection SQL | GORM utilise des requêtes paramétrées → protection native. |
| Flux chiffrés siège↔agences | Géré par la partie INFRA (VPN/IPSec). Côté app : HTTPS de bout en bout. |

> Note : la matrice des droits par pôle (Direction, Commercial, Comm. & Mktg, Admin/RH, IT & Support) concerne d'abord le partage de fichiers (partie INFRA). On la reflète aussi dans l'application via le RBAC pour rester cohérent — c'est un bon point à souligner à l'oral (lien DEV ↔ INFRA).

## 7. Rôles applicatifs (matrice d'accès fonctionnelle)

| Rôle | Accès principal |
|---|---|
| Visiteur (non authentifié) | Catalogue + filtres de recherche de base. |
| Client acheteur | Compte, favoris, alertes, messagerie avec agents, demandes de visite. |
| Client vendeur | Publication d'annonces (soumises à validation d'un agent). |
| Agent commercial | Gestion des biens assignés, dossiers de vente, messagerie, planning, estimations. |
| Directeur d'agence | Supervision de son agence + statistiques locales + réunions. |
| Personnel du siège | Vue nationale, statistiques toutes agences, tableaux de bord IA. |
| IT / Support | Administration des comptes et des droits. |

## 8. Prochaines étapes proposées

1. **Module BDD** (en cours) → voir `db/schema.sql`.
2. Initialisation du projet Go (squelette, connexion MariaDB, premier `/health`).
3. Module Authentification (inscription/login JWT) + middleware RBAC.
4. Module Biens (CRUD + recherche/filtres + upload photos).
5. Modules Favoris / Alertes / Messagerie / Planning.
6. Microservice Python Data/IA.
7. Front-end Tailwind (à l'arrivée des wireframes).
8. Documentation Swagger finalisée.
