#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
BACKEND_PID_FILE="$RUN_DIR/backend.pid"
FRONTEND_PID_FILE="$RUN_DIR/frontend.pid"
BACKEND_LOG="$RUN_DIR/backend.log"
FRONTEND_LOG="$RUN_DIR/frontend.log"

mkdir -p "$RUN_DIR"

is_running() {
  local pid="$1"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

read_pid() {
  local file="$1"
  if [[ -f "$file" ]]; then
    tr -d '[:space:]' < "$file"
  else
    printf ''
  fi
}

ensure_not_running() {
  local name="$1"
  local file="$2"
  local pid
  pid="$(read_pid "$file")"
  if is_running "$pid"; then
    printf '%s is already running (pid=%s)\n' "$name" "$pid"
    exit 1
  fi
  rm -f "$file"
}

ensure_not_running "backend" "$BACKEND_PID_FILE"
ensure_not_running "frontend" "$FRONTEND_PID_FILE"

(
  cd "$ROOT_DIR"
  nohup go run . > "$BACKEND_LOG" 2>&1 &
  echo $! > "$BACKEND_PID_FILE"
)

(
  cd "$ROOT_DIR/web"
  nohup npm run dev -- --host 127.0.0.1 > "$FRONTEND_LOG" 2>&1 &
  echo $! > "$FRONTEND_PID_FILE"
)

sleep 2

BACKEND_PID="$(read_pid "$BACKEND_PID_FILE")"
FRONTEND_PID="$(read_pid "$FRONTEND_PID_FILE")"

if ! is_running "$BACKEND_PID"; then
  echo "backend failed to start; see $BACKEND_LOG"
  exit 1
fi

if ! is_running "$FRONTEND_PID"; then
  echo "frontend failed to start; see $FRONTEND_LOG"
  exit 1
fi

cat <<EOF
Started CloudAlbum local dev services.
- Backend:  http://localhost:8080  (pid: $BACKEND_PID)
- Frontend: http://127.0.0.1:3000 (pid: $FRONTEND_PID)
Logs:
- $BACKEND_LOG
- $FRONTEND_LOG
EOF
