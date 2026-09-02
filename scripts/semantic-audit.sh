#!/usr/bin/env bash
set -Eeuo pipefail

root=${1:-.}
root=$(cd "$root" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

inspect=$(go run ./cmd/gooo-release-transport-conformer inspect --source "$root/.gooo/release-transport.gooo")
jq -e '
  .contract_id == "gooo-release-transport-conformer/v6" and
  .denominator == 32 and
  .semantic_ir_schema == "gooo/release-transport-conformer/semantic-ir/v6" and
  .version.version == "0.1.7" and .previous_release.release_id == "380375220" and
  .states == ["PRECHECK","TAGGED","DRAFT_CREATED","ASSETS_UPLOADED","ASSETS_AUDITED","PUBLISHED_IMMUTABLE"] and
  (.transitions | length) == 5 and
  (.activities | length) == 12 and
  (.scenarios | length) == 32 and
  (.activities | unique | length) == 12
' <<<"$inspect" >/dev/null

go run ./cmd/gooo-release-transport-conformer generate --root "$root" --output "$work/generated" >/dev/null
cmp "$work/generated/release-workflow.yml" "$root/.github/workflows/release.yml"
test "$(find "$work/generated" -maxdepth 1 -type f | wc -l | tr -d ' ')" = 6
