# Ymmo — Front-end

Web interface for the Ymmo platform.
Stack: **HTML · Tailwind CSS · JavaScript (vanilla, ES modules)**.

It talks to the Go API (`:8080`) and, later, the Python AI service (`:8000`).

## Run (dev)

The pages call the API, so they must be served over HTTP (not opened as
`file://`, which the API's CORS rejects). The API allows `localhost:3000`, so
serve on port 3000:

```bash
cd frontend
python -m http.server 3000
# open http://localhost:3000
```

Make sure the API is running (`docker compose up -d`) and the database seeded.

## Structure

```
frontend/
  index.html            # home + design-system reference
  js/
    config.js           # API / AI base URLs
    api.js              # fetch client + JWT handling
    auth.js             # login / logout / current user
    tailwind-config.js  # brand colors (charte graphique), shared by all pages
    home.js             # home page script
```

## Design tokens (charte graphique)

| Token | Color | Usage |
|---|---|---|
| `midnight` | #2c3e50 | Text / titles |
| `alyssum` | #efebe7 | Background |
| `hibiscus` | #b63753 | Primary buttons, key nav |
| `slate2` | #7f8c8d | Borders, secondary text |
| `gold` | #d4af37 | Prestige badges, AI highlights |

Use them as Tailwind classes: `bg-hibiscus`, `text-midnight`, `border-slate2`…

## Accessibility

Pages use semantic landmarks (`header`/`main`/`footer`/`nav`), a skip link,
`lang="fr"`, visible focus states and descriptive labels (WCAG / ARIA).

## Production build (next step)

Dev uses the Tailwind CDN for zero setup. For the final delivery (performance
criterion), switch to a compiled stylesheet with the Tailwind CLI:

```bash
npx tailwindcss -i ./src/input.css -o ./css/styles.css --minify
```

…then replace the CDN `<script>` with `<link rel="stylesheet" href="./css/styles.css">`.
The `tailwind.config.js` reuses the same brand colors.
