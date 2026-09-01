#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "$#" -lt 4 ]]; then
  echo "usage: measure-command.sh OUTPUT_JSON STAGE LOG COMMAND [ARGS...]" >&2
  exit 64
fi

output=$1
stage=$2
log=$3
shift 3
time_output=$(mktemp)
trap 'rm -f "$time_output"' EXIT
status=0

if /usr/bin/time -f 'wall_seconds=%e\npeak_rss_kib=%M' -o "$time_output" "$@" >"$log" 2>&1; then
  status=0
else
  status=$?
fi

wall_seconds=$(awk -F= '$1 == "wall_seconds" {print $2}' "$time_output")
peak_rss_kib=$(awk -F= '$1 == "peak_rss_kib" {print $2}' "$time_output")
wall_ms=$(awk -v seconds="${wall_seconds:-0}" 'BEGIN { printf "%d", (seconds * 1000) + 0.5 }')
peak_rss_kib=${peak_rss_kib:-0}

jq -n \
  --arg stage "$stage" \
  --argjson exit_code "$status" \
  --argjson wall_ms "$wall_ms" \
  --argjson peak_rss_kib "$peak_rss_kib" \
  '{stage:$stage,exit_code:$exit_code,wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib}' >"$output"

exit "$status"
