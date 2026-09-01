#!/usr/bin/env bash
set -Eeuo pipefail

root=${1:-.}
root=$(cd "$root" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

inspect=$(go run ./cmd/gooo-release-transport-conformer inspect --source "$root/.gooo/release-transport.gooo")
jq -e '
  .contract_id == "gooo-release-transport-conformer/v2" and
  .denominator == 14 and
  (.activities | length) == 9 and
  (.scenarios | length) == 14 and
  (.activities | unique | length) == 9
' <<<"$inspect" >/dev/null

gofmt_files=$(git -C "$root" ls-files '*.go' 2>/dev/null || true)
if [[ -n "$gofmt_files" ]]; then
  test -z "$(gofmt -l $gofmt_files)"
fi

go run ./cmd/gooo-release-transport-conformer generate --root "$root" --output "$work/generated" >/dev/null
cmp "$work/generated/release-workflow.yml" "$root/.github/workflows/release.yml"
test "$(find "$work/generated" -maxdepth 1 -type f | wc -l | tr -d ' ')" = 5
