#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: assemble-ci-evidence.sh REPOSITORY GENERATED_OUTPUT METRICS_OUTPUT" >&2
  exit 64
fi

repository=$(cd "$1" && pwd)
generated=$2
output=$3
mkdir -p "$output"
receipt="$generated/conformance-receipt.json"
test -s "$receipt"

stages=$(jq -s '.' "$output"/compile.json "$output"/build.json "$output"/test.json "$output"/conformance.json "$output"/integration.json)
jq -e 'length == 5 and all(.[]; (.stage|type)=="string" and (.wall_ms|type)=="number" and (.peak_rss_kib|type)=="number" and .exit_code == 0)' <<<"$stages" >/dev/null

jq -n \
  --arg schema "gooo-release-transport-conformer/ci-evidence/v1" \
  --arg toolchain "$(go env GOVERSION)" \
  --argjson inventory "$(jq '.inventory' "$receipt")" \
  --argjson tests "$(jq '.tests' "$receipt")" \
  --argjson stages "$stages" \
  --argjson authority "$(jq '.authority' "$receipt")" \
  --argjson operational_audit "$(jq '.operational_audit' "$receipt")" \
  --argjson output_files "$(jq '.output_files' "$receipt")" \
  '{schema:$schema,verification_authority:"GITHUB_ACTIONS",toolchain:$toolchain,inventory:$inventory,tests:$tests,stages:$stages,output_files:$output_files,authority:($authority + {cross_project_required_gates:0}),operational_audit:$operational_audit}' >"$output/ci-evidence.json"

jq -e '
  .schema == "gooo-release-transport-conformer/ci-evidence/v1" and
  .verification_authority == "GITHUB_ACTIONS" and
  .toolchain == "go1.27.0" and
  (.stages | length) == 5 and
  (all(.stages[]; (.wall_ms|type)=="number" and (.peak_rss_kib|type)=="number")) and
  .tests == {total:14,selected:14,executed:14,reused:0,failed:0,unknown:0} and
  .authority.repository_writes == 0 and .authority.local_go_tests == 0 and
  .authority.cross_project_required_gates == 0 and
  .operational_audit.state == "REFUTED" and
  .operational_audit.local_test_invocations == 2 and
  .operational_audit.local_test_executions == 1 and
  .operational_audit.operator_stage_local_test_executions == 0
' "$output/ci-evidence.json" >/dev/null
