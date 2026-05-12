#!/usr/bin/env bash
set -euo pipefail
DIR="pkg/database/migrations"
regex='^[0-9]{6}_[a-z0-9_]+\.(up|down)\.sql$'
fail=0
for f in "$DIR"/*.sql; do
  name="$(basename "$f")"
  if ! [[ "$name" =~ $regex ]]; then echo "::error::naming: $name"; fail=1; fi
done
for f in "$DIR"/*.up.sql; do
  pair="${f/.up./.down.}"
  if [[ ! -f "$pair" ]]; then echo "::error::missing pair for $(basename "$f")"; fail=1; fi
done
if [[ -d migrations ]]; then echo "::error::root migrations/ folder must not exist"; fail=1; fi
exit "$fail"
