#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
BACKEND_PID_FILE="$RUN_DIR/backend.pid"
FRONTEND_PID_FILE="$RUN_DIR/frontend.pid"
BACKEND_LOG="$RUN_DIR/backend.log"
FRONTEND_LOG="$RUN_DIR/frontend.log"

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

print_status() {
  local name="$1"
  local pid_file="$2"
  local url="$3"
  local log_file="$4"
  local pid
  pid="$(read_pid "$pid_file")"

  if is_running "$pid"; then
    printf '%s: running (pid=%s)\n' "$name" "$pid"
    printf '  url: %s\n' "$url"
    printf '  log: %s\n' "$log_file"
  elif [[ -n "$pid" ]]; then
    printf '%s: stopped (stale pid file: %s)\n' "$name" "$pid"
    printf '  log: %s\n' "$log_file"
  else
    printf '%s: stopped\n' "$name"
    printf '  log: %s\n' "$log_file"
  fi
}

print_status "backend" "$BACKEND_PID_FILE" "http://localhost:8080" "$BACKEND_LOG"
printf '\n'
print_status "frontend" "$FRONTEND_PID_FILE" "http://127.0.0.1:3000" "$FRONTEND_LOG"
