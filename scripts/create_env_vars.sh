#! usr/bin/bash

DB_URL="postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cat <<EOF > "$PROJECT_ROOT/.env"
DB_URL=$DB_URL
EOF