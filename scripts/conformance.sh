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
test "$actual" = "conformance-receipt.json,human-report.md,release-workflow.yml,semantic-ir.json,transport-events.ndjson,transport-manifest.json"

jq -e '
  .schema == "gooo/release-transport-conformer/conformance-receipt/v6" and
  .decision == "CLOSED" and .terminal == "FIXED_POINT" and .state == "PUBLISHED_IMMUTABLE" and
  .conformance_closed == true and .denominator == 32 and .case_denominator == 32 and .indicator_denominator == 12 and
  .semantic_ir.previous_denominator == 20 and .semantic_ir.append_only == true and
  (.semantic_ir.previous_release.assets | length) == 3 and
  ([.semantic_ir.previous_release.assets[] | select(.id != "" and (.digest|startswith("sha256:")))] | length) == 3 and
  .summary == {CLOSED:12, UNKNOWN:4, REFUTED:16} and
  (.scenarios | length) == 32 and (.case_vectors | length) == 32 and (.indicator_vectors | length) == 12 and
  (.scenarios | map(.id)) == [
    "draft-assets-publish-immutable", "deterministic-replay", "all-asset-digests-match",
    "exact-annotated-tag-target", "missing-operator-immutable-policy-receipt", "stale-source-run",
    "missing-git-identity", "tag-collision", "publish-before-assets", "published-immutable-false",
    "checksum-path-mismatch", "user-token-secret-or-admin-endpoint-in-actions",
    "resume-existing-exact-draft-by-list-id", "existing-draft-target-or-assets-mismatch",
    "upload-assets-via-release-upload-url", "upload-assets-via-api-endpoint",
    "reconcile-symbolic-target-with-peeled-tag-target", "treat-symbolic-target-commitish-as-exact-commit",
    "continue-with-create-response-draft-id", "require-immediate-draft-list-visibility-after-create",
    "linear-forward-state-machine", "pre-mutation-fixture-conformance",
    "failed-attempt-preserves-objects-and-burns-version", "direct-main-release-target",
    "ambiguous-compare-lineage", "wrong-target-commitish", "tag-release-ordering-error",
    "missing-immutable-setting-evidence", "asset-manifest-count-name-size-digest-mismatch",
    "duplicate-or-burned-version-reuse", "mutable-published-release", "fixed-point-only-publication"
  ] and
  ([.scenarios[] | select(.decision == "UNKNOWN") | .unknown | (.stage != "" and .step != "" and .reason != "" and .unknown_class != "" and .next_operation != "" and (.blocked_by|length) > 0)] | all) and
  (.activity_binding_counts | length) == 12 and
  ([.activity_binding_counts[]] | all(. == 1)) and
  (.semantic_ir.schema == "gooo/release-transport-conformer/semantic-ir/v6") and
  (.semantic_ir.states == ["PRECHECK","TAGGED","DRAFT_CREATED","ASSETS_UPLOADED","ASSETS_AUDITED","PUBLISHED_IMMUTABLE"]) and
  ((.semantic_ir.transitions | length) == 5) and
  (.indicator_vectors | all(.[]; .denominator == 1 and (.numerator == 0 or .numerator == 1))) and
  (.case_vectors | all(.[]; .denominator == 1 and (.numerator == 1))) and
  .authority.repository_writes == 0 and .authority.commits == 0 and .authority.pushes == 0 and
  .authority.merges == 0 and .authority.tags == 0 and .authority.releases == 0 and
  .authority.local_go_tests == 0 and .authority.cross_project_required_gates == 0 and .authority.caller_owned_output == true and
  .authority.source_repository_read_only == true and
  .operational_audit.state == "REFUTED" and
  .operational_audit.authoring_local_test_invocations == 2 and
  .operational_audit.authoring_local_test_executions == 1 and
  .operational_audit.authoring_static_validation_invocations == 2 and
  .operational_audit.authoring_static_validation_executions == 2 and
  .operational_audit.operator_stage_local_test_invocations == 0 and
  .operational_audit.operator_stage_local_test_executions == 0 and
  .operational_audit.local_test_invocations == 2 and
  .operational_audit.local_test_executions == 1 and
  .output_files == ["release-workflow.yml", "semantic-ir.json", "transport-manifest.json", "transport-events.ndjson", "conformance-receipt.json", "human-report.md"] and
  .attempt.mutation_started == false and .attempt.preserve_never_delete == true and .attempt.burned_version == false and
  (.scope_note | contains("not a global safety claim"))
' "$output/conformance-receipt.json" >/dev/null

workflow=$output/release-workflow.yml
! grep -En 'secrets\.|GITHUB_TOKEN|GH_PAT|\bPAT\b|/immutable-releases|administration|admin:repo|compare/|release_delete|tag_delete|asset_delete|same_version_recreate|gh release delete|--force|recreate.*tag|move.*tag' "$workflow"
grep -En 'GH_TOKEN: \$\{\{ github\.token \}\}' "$workflow" >/dev/null
grep -En 'upload_url|uploads\.github\.com|upload_base' "$workflow" >/dev/null
! grep -En -- '--method POST.*releases/\$DRAFT_ID/assets' "$workflow"
fixture_line=$(grep -n 'Conform payloads and fixture API responses before mutation' "$workflow" | cut -d: -f1)
precheck_line=$(grep -n 'PRECHECK exact lineage policy and unused version' "$workflow" | cut -d: -f1)
tag_line=$(grep -n 'TAGGED create annotated tag exactly once' "$workflow" | cut -d: -f1)
draft_line=$(grep -n 'DRAFT_CREATED create draft before any asset' "$workflow" | cut -d: -f1)
upload_line=$(grep -n 'ASSETS_UPLOADED assemble exact manifest from exact target' "$workflow" | cut -d: -f1)
audit_line=$(grep -n 'ASSETS_AUDITED verify count name size and digest exactly' "$workflow" | cut -d: -f1)
publish_line=$(grep -n 'PUBLISHED_IMMUTABLE publish only audited draft at FIXED_POINT' "$workflow" | cut -d: -f1)
incident_line=$(grep -n 'Emit exact attempt receipt and preserve failures' "$workflow" | cut -d: -f1)
test "$fixture_line" -lt "$precheck_line"
test "$precheck_line" -lt "$tag_line"
test "$tag_line" -lt "$draft_line"
test "$draft_line" -lt "$upload_line"
test "$upload_line" -lt "$audit_line"
test "$audit_line" -lt "$publish_line"
test "$publish_line" -lt "$incident_line"

if [[ -n "$before" ]]; then
  after=$(git -C "$repository" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
  test "$before" = "$after"
fi
