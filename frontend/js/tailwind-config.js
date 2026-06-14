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
        // Secondary text/borders. Darkened from the original #7f8c8d so that
        // small secondary text reaches WCAG AA contrast (>= 4.5:1 on white).
        slate2: "#5a6a6b",
        gold: "#d4af37", // Accent — prestige badges (use with dark text, not white)
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "Segoe UI", "sans-serif"],
      },
    },
  },
};
