#!/bin/bash
# Runs once at database initialization (after 01-schema.sql).
# Creates a dedicated, READ-ONLY user for the Python AI service: if that process
# is ever compromised, it can only SELECT — never write/update/delete.
set -e

mariadb -u root -p"${MARIADB_ROOT_PASSWORD}" <<-EOSQL
  CREATE USER IF NOT EXISTS '${AI_DB_USER}'@'%' IDENTIFIED BY '${AI_DB_PASSWORD}';
  GRANT SELECT ON \`${MARIADB_DATABASE}\`.* TO '${AI_DB_USER}'@'%';
  FLUSH PRIVILEGES;
EOSQL
