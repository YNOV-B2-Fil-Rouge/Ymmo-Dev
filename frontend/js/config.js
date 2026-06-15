// Front-end configuration.
// The API is reached on the SAME origin: nginx reverse-proxies /api, /uploads,
// /health and /swagger to the Go container. This avoids hard-coding a host and
// removes the need for CORS (works identically in dev and on a server).
export const CONFIG = {
  API_BASE: "/api/v1",
};
