# Front-end is 100% static (HTML + Tailwind via CDN + vanilla JS modules),
# so there is no build step: we just serve the files with nginx.
FROM nginx:1.27-alpine

# Custom server config (multi-page routing, light caching).
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static assets only (explicit copies so config files never reach the web root).
COPY *.html /usr/share/nginx/html/
COPY js /usr/share/nginx/html/js

EXPOSE 80
