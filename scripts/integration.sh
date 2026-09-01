#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: integration.sh REPOSITORY BINARY OUTPUT" >&2
  exit 64
fi

repository=$(cd "$1" && pwd)
binary=$2
output=$3
mkdir -p "$output"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

"$binary" generate --root "$repository" --source "$repository/.gooo/release-transport.gooo" --output "$work/generated" >/dev/null
cmp "$work/generated/release-workflow.yml" "$repository/.github/workflows/release.yml"
jq -e '.output_files == ["release-workflow.yml", "transport-manifest.json", "transport-events.ndjson", "conformance-receipt.json", "human-report.md"]' "$work/generated/conformance-receipt.json" >/dev/null
jq -e '.activities | length == 12 and (unique | length) == 12' "$work/generated/transport-manifest.json" >/dev/null
jq -s -e 'length == 12 and (map(.activity) | unique | length == 12) and (map(.ordinal) == [1,2,3,4,5,6,7,8,9,10,11,12])' < <(sed '/^[[:space:]]*$/d' "$work/generated/transport-events.ndjson") >/dev/null
jq -e 'all(.scenarios[]; .decision == "CLOSED" or .decision == "UNKNOWN" or .decision == "REFUTED")' "$work/generated/conformance-receipt.json" >/dev/null
jq -n --arg schema "gooo-release-transport-conformer/integration/v1" --arg workflow_digest "$(sha256sum "$work/generated/release-workflow.yml" | awk '{print "sha256:" $1}')" '{schema:$schema,decision:"CLOSED",workflow_digest:$workflow_digest,generated_files:5}' >"$output/integration.json"
