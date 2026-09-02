# gooo-release-transport-conformer

`gooo-release-transport-conformer` turns a released `.gooo` declaration into a
semantic IR, a standard draft-first GitHub Actions release workflow, and a
fail-closed transport receipt. The `.gooo` declaration owns the release
version, previous immutable identity, merged-PR lineage, exact target commit,
annotated tag, draft identity, expected asset manifest, observed API receipts,
state machine, and burned-version ledger. Go only parses, projects, evaluates,
and generates caller-owned output.

The v6 contract appends twelve transport cases to the released v5 denominator:
exactly thirty-two cases and twelve indicator vectors. It preserves the
previous twenty cases and adds the bounded state machine, pre-mutation fixture
conformance, operational preservation, lineage, target, immutable-setting,
asset-manifest, and version-reuse boundaries. Scenario resolution precedence is `REFUTED > UNKNOWN > CLOSED`. Every unknown carries
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.
The top-level receipt is `CLOSED` only when all thirty-two declared cases are
classified exactly as declared. That is a transport-contract result, not a
global safety claim.

The generator observes the source repository read-only and writes exactly six
files to an empty caller-owned output directory:

- `release-workflow.yml`
- `semantic-ir.json`
- `transport-manifest.json`
- `transport-events.ndjson`
- `conformance-receipt.json`
- `human-report.md`

The receipt also preserves an authoring audit. This checkout records two local
test invocations and one actual local test execution during initial authoring;
the operational state is `REFUTED` for that process fact. The semantic
32-case result is separate, and from the PR onward GitHub Actions is the
only verification authority. Product runtime authority remains zero for
repository writes, commits, pushes, merges, tags, releases, and local product
tests.

The generated workflow uses only the standard `github.token`. It performs a
read-only PRECHECK against fixture API responses, exact merged-PR lineage,
the previous immutable release, and an unused patch version before any public
mutation. It then advances only
`PRECHECK -> TAGGED -> DRAFT_CREATED -> ASSETS_UPLOADED -> ASSETS_AUDITED -> PUBLISHED_IMMUTABLE`, using the create-response draft ID and the draft detail `upload_url` for binary assets. A failed attempt emits an `OPERATIONAL_REFUTED` receipt containing `mutation_started` and every created public object ID, preserves those objects, and burns the version. No tag, draft, release, or asset is deleted, edited, or reused. It does not call an administration endpoint or consume a user token secret. The operator immutable-setting receipt is external and digest-bound.

The twelve activities are partitioned exactly into `FOUNDATION`, `COHERENCE`,
and `REGRESSION` families and `DRIVER`, `OUTCOME`, and `GUARDRAIL` roles. The
receipt exposes per-case and per-indicator vectors; it deliberately contains
no aggregate score or percentage.

## Verification

GitHub Actions is the authoritative verification environment for compile,
build, test, conformance, integration, and artifact evidence. The generator
only writes an empty caller-owned output directory outside the observed
repository; the conformance script also checks that the source repository
status is unchanged.

To bind an operator receipt after the user-level setting has been enabled:

```sh
./scripts/enable-immutable-releases.sh OWNER/REPOSITORY /tmp/operator-policy-receipt.json
```

The release workflow is dispatched on `main` with the exact merged SHA, the
merged PR number, and the digest from that external receipt. Failed or
non-immutable release attempts are retained; existing tags and releases are
never edited, deleted, or reused. PR CI has zero release mutation authority.
