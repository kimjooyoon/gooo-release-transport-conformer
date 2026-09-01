#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: conformance.sh REPOSITORY BINARY OUTPUT" >&2
  exit 64
fi

repository=$(cd "$1" && pwd)
binary=$2
output=$3
before=""
if git -C "$repository" rev-parse --show-toplevel >/dev/null 2>&1; then
  before=$(git -C "$repository" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
fi

"$binary" conformance --root "$repository" --source "$repository/.gooo/release-transport.gooo" --output "$output"

actual=$(find "$output" -maxdepth 1 -type f -exec basename {} \; | sort | paste -sd, -)
test "$actual" = "conformance-receipt.json,human-report.md,release-workflow.yml,transport-events.ndjson,transport-manifest.json"

jq -e '
  .schema == "gooo/release-transport-conformer/conformance-receipt/v2" and
  .decision == "CLOSED" and .conformance_closed == true and .denominator == 14 and
  .summary == {CLOSED:5, UNKNOWN:3, REFUTED:6} and
  (.scenarios | length) == 14 and
  (.scenarios | map(.id)) == [
    "draft-assets-publish-immutable", "deterministic-replay", "all-asset-digests-match",
    "exact-annotated-tag-target", "missing-operator-immutable-policy-receipt", "stale-source-run",
    "missing-git-identity", "tag-collision", "publish-before-assets", "published-immutable-false",
    "checksum-path-mismatch", "user-token-secret-or-admin-endpoint-in-actions",
    "resume-existing-exact-draft-by-list-id", "existing-draft-target-or-assets-mismatch"
  ] and
  ([.scenarios[] | select(.decision == "UNKNOWN") | .unknown | (.stage != "" and .step != "" and .reason != "" and .unknown_class != "" and .next_operation != "" and (.blocked_by|length) > 0)] | all) and
  (.activity_binding_counts | length) == 9 and
  ([.activity_binding_counts[]] | all(. == 1)) and
  .authority.repository_writes == 0 and .authority.commits == 0 and .authority.pushes == 0 and
  .authority.merges == 0 and .authority.tags == 0 and .authority.releases == 0 and
  .authority.local_go_tests == 0 and .authority.caller_owned_output == true and
  .authority.source_repository_read_only == true and
  .operational_audit.state == "REFUTED" and
  .operational_audit.authoring_local_test_invocations == 2 and
  .operational_audit.authoring_local_test_executions == 1 and
  .operational_audit.operator_stage_local_test_invocations == 0 and
  .operational_audit.operator_stage_local_test_executions == 0 and
  .operational_audit.local_test_invocations == 2 and
  .operational_audit.local_test_executions == 1 and
  .output_files == ["release-workflow.yml", "transport-manifest.json", "transport-events.ndjson", "conformance-receipt.json", "human-report.md"] and
  (.scope_note | contains("not a global safety claim"))
' "$output/conformance-receipt.json" >/dev/null

workflow=$output/release-workflow.yml
! grep -En 'secrets\.|GITHUB_TOKEN|GH_PAT|\bPAT\b|/immutable-releases|administration|admin:repo|gh release delete|--force|recreate.*tag|move.*tag' "$workflow"
grep -En 'GH_TOKEN: \$\{\{ github\.token \}\}' "$workflow" >/dev/null
draft_line=$(grep -n 'Create draft release before assets' "$workflow" | cut -d: -f1)
upload_line=$(grep -n 'Upload every release asset' "$workflow" | cut -d: -f1)
publish_line=$(grep -n 'Publish release after all uploads' "$workflow" | cut -d: -f1)
verify_line=$(grep -n 'Verify public immutable release' "$workflow" | cut -d: -f1)
test "$draft_line" -lt "$upload_line"
test "$upload_line" -lt "$publish_line"
test "$publish_line" -lt "$verify_line"

if [[ -n "$before" ]]; then
  after=$(git -C "$repository" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
  test "$before" = "$after"
fi
