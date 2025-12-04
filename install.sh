#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
DEFAULT_TARGET="${PWD}/mediastack-install"
TARGET_DIR=${1:-$DEFAULT_TARGET}

PS3="Choose a compose flavor (1-3): "
OPTIONS=("full-download-vpn" "mini-download-vpn" "no-download-vpn")

if ! command -v docker >/dev/null 2>&1; then
  echo "❌ docker not found; install Docker before continuing." >&2
  exit 1
fi

if ! command -v docker compose >/dev/null 2>&1; then
  echo "❌ docker compose plugin not found; install Docker Compose v2." >&2
  exit 1
fi

printf "\nSelect a MediaStack configuration to stage:\n"
select CHOICE in "${OPTIONS[@]}"; do
  if [[ -n "$CHOICE" ]]; then
    break
  fi
  echo "Please pick 1, 2, or 3." >&2
done

mkdir -p "$TARGET_DIR"
if command -v rsync >/dev/null 2>&1; then
  rsync -a "$SCRIPT_DIR/base-working-files/" "$TARGET_DIR/"
else
  cp -a "$SCRIPT_DIR/base-working-files/." "$TARGET_DIR/"
fi
cp "$SCRIPT_DIR/.env.example" "$TARGET_DIR/.env"
cp "$SCRIPT_DIR/$CHOICE/docker-compose.yaml" "$TARGET_DIR/docker-compose.yaml"

cat <<INFO

✅ Staged $CHOICE into: $TARGET_DIR
   - Edit $TARGET_DIR/.env with your secrets and port choices.
   - Then run: (cd $TARGET_DIR && docker compose up -d)
INFO
