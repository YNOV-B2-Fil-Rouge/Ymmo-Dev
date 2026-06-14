#!/usr/bin/env sh
# Regenerate db/seed.sql from the LIVE database (the data you created via the app).
#
# Usage (stack running):   sh db/dump-seed.sh      (or: bash db/dump-seed.sh)
#
# Notes:
#  - The reference tables (roles, departments, property_categories,
#    department_permissions) are populated by schema.sql, so they are
#    intentionally EXCLUDED here to avoid duplicate-key errors on re-init.
#  - The dump is assembled inside the container and copied out with `docker cp`,
#    so the file keeps clean UTF-8 (no Windows/PowerShell encoding issues).
#  - Uploaded photo FILES live in the `ymmo_uploads` volume, not in the DB.
#    This seed only captures their URLs; copy the volume separately if needed.
set -e

CONTAINER="${DB_CONTAINER:-ymmo-db}"

docker exec "$CONTAINER" sh -c '
  set -e
  # Demo tables, parents first (insertion order).
  TABLES="agencies users properties property_photos favorites alerts \
conversations messages visits meetings meeting_participants \
sale_files property_views price_history seller_applications"
  # Same list, children first (reset order).
  RESET="messages conversations visits meeting_participants meetings \
sale_files property_views price_history favorites alerts \
seller_applications property_photos properties users agencies"

  {
    echo "-- ====================================================================="
    echo "--  YMMO seed — generated from the live database by db/dump-seed.sh"
    echo "--  Loaded after schema.sql (reference tables already seeded there)."
    echo "-- ====================================================================="
    echo "USE ymmo;"
    echo "SET NAMES utf8mb4;"
    echo
    echo "SET FOREIGN_KEY_CHECKS = 0;"
    for t in $RESET; do echo "DELETE FROM $t;"; done
    echo "SET FOREIGN_KEY_CHECKS = 1;"
    echo
    mariadb-dump -u root -p"$MARIADB_ROOT_PASSWORD" \
      --no-create-info --complete-insert --skip-extended-insert --single-transaction \
      --no-tablespaces --skip-comments --skip-add-locks --skip-disable-keys \
      ymmo $TABLES
  } > /tmp/ymmo_seed.sql
'

docker cp "$CONTAINER:/tmp/ymmo_seed.sql" ./db/seed.sql
echo "OK -> db/seed.sql regenerated from the live database."
