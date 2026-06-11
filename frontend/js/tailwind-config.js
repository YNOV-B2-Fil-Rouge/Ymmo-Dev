// Shared Tailwind configuration (brand colors from the charte graphique).
// Loaded on every page right after the Tailwind script, so the design tokens
// are defined in ONE place (DRY).
tailwind.config = {
  theme: {
    extend: {
      colors: {
        midnight: "#2c3e50", // Text — Midnight Blue
        alyssum: "#efebe7", // Background — White Alyssum
        hibiscus: "#b63753", // Primary — buttons & key nav
        slate2: "#7f8c8d", // Secondary — borders, icons, placeholders
        gold: "#d4af37", // Accent — prestige badges, AI highlights
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "Segoe UI", "sans-serif"],
      },
    },
  },
};
