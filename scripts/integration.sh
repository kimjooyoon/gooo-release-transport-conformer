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
jq -e '.output_files == ["release-workflow.yml", "semantic-ir.json", "transport-manifest.json", "transport-events.ndjson", "conformance-receipt.json", "human-report.md"] and .attempt.mutation_started == false and .attempt.created_public_object_ids.asset_ids == []' "$work/generated/conformance-receipt.json" >/dev/null
jq -e '(.activities | length) == 12 and ((.activities | unique | length) == 12) and .previous_denominator == 20 and .append_only == true and .state_machine == ["PRECHECK","TAGGED","DRAFT_CREATED","ASSETS_UPLOADED","ASSETS_AUDITED","PUBLISHED_IMMUTABLE"]' "$work/generated/transport-manifest.json" >/dev/null
jq -e '.schema == "gooo/release-transport-conformer/semantic-ir/v6" and (.expected_asset_manifest | all(.[]; .name != "" and .size != "" and .digest != "")) and (.observed_api_receipts | length) == 5' "$work/generated/semantic-ir.json" >/dev/null
jq -s -e 'length == 12 and (map(.activity) | unique | length == 12) and (map(.ordinal) == [1,2,3,4,5,6,7,8,9,10,11,12])' < <(sed '/^[[:space:]]*$/d' "$work/generated/transport-events.ndjson") >/dev/null
jq -e 'all(.scenarios[]; (.decision == "CLOSED" or .decision == "UNKNOWN" or .decision == "REFUTED") and .resolution == "FIXED_POINT") and (.case_vectors | length) == 32 and (.indicator_vectors | length) == 12' "$work/generated/conformance-receipt.json" >/dev/null
jq -n --arg schema "gooo-release-transport-conformer/integration/v1" --arg workflow_digest "$(sha256sum "$work/generated/release-workflow.yml" | awk '{print "sha256:" $1}')" '{schema:$schema,decision:"CLOSED",workflow_digest:$workflow_digest,generated_files:6}' >"$output/integration.json"
