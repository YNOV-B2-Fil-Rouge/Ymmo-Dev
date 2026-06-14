// Shared Tailwind config: brand colors from the charte graphique.
tailwind.config = {
  theme: {
    extend: {
      colors: {
        midnight: "#2c3e50", // Text — Midnight Blue
        alyssum: "#efebe7", // Background — White Alyssum
        hibiscus: "#b63753", // Primary — buttons & key nav
        slate2: "#5a6a6b", // Secondary text/borders (darkened for WCAG AA contrast)
        gold: "#d4af37", // Accent — prestige badges (use with dark text)
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "Segoe UI", "sans-serif"],
      },
    },
  },
};
