#!/usr/bin/env bash
set -Eeuo pipefail

root=${1:-.}
root=$(cd "$root" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

inspect=$(go run ./cmd/gooo-release-transport-conformer inspect --source "$root/.gooo/release-transport.gooo")
jq -e '
  .contract_id == "gooo-release-transport-conformer/v3" and
  .denominator == 16 and
  (.activities | length) == 10 and
  (.scenarios | length) == 16 and
  (.activities | unique | length) == 10
' <<<"$inspect" >/dev/null

go run ./cmd/gooo-release-transport-conformer generate --root "$root" --output "$work/generated" >/dev/null
cmp "$work/generated/release-workflow.yml" "$root/.github/workflows/release.yml"
test "$(find "$work/generated" -maxdepth 1 -type f | wc -l | tr -d ' ')" = 5
