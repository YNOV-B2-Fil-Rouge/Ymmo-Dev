-- Convert absolute photo URLs to same-origin relative paths so images load
-- through the front nginx proxy on any host (run once on an existing database).
UPDATE property_photos
SET url = REPLACE(url, 'http://localhost:8080/uploads/', '/uploads/')
WHERE url LIKE 'http://localhost:8080/%';
