# gooo-release-transport-conformer

`gooo-release-transport-conformer` turns a released `.gooo` declaration into a
standard, draft-first GitHub Actions release workflow and a fail-closed
transport receipt. The `.gooo` declaration owns the release state machine;
Go only parses, evaluates, and generates caller-owned output.

The fixed denominator has exactly fourteen scenarios: five closed transport
contracts, three explicit unknowns, and six preserved refutations. Scenario
resolution precedence is `REFUTED > UNKNOWN > CLOSED`. Every unknown carries
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.
The top-level receipt is `CLOSED` only when all twelve declared scenarios are
classified exactly as declared. That is a transport-contract result, not a
global safety claim.

The generator observes the source repository read-only and writes exactly five
files to an empty caller-owned output directory:

- `release-workflow.yml`
- `transport-manifest.json`
- `transport-events.ndjson`
- `conformance-receipt.json`
- `human-report.md`

The receipt also preserves an authoring audit. This checkout records two local
test invocations and one actual local test execution during initial authoring;
the operational state is `REFUTED` for that process fact. The semantic
14-scenario result is separate, and from the PR onward GitHub Actions is the
only verification authority. Product runtime authority remains zero for
repository writes, commits, pushes, merges, tags, releases, and local product
tests.

The generated workflow uses only the standard `github.token`. It creates an
annotated tag, creates or reconciles a draft release through the releases list
API, resumes the exact draft by list-derived ID, uploads only when its asset
names/digests are empty or identical, publishes only after upload, and verifies
the public release, tag object, and asset digests. It does not call an
administration endpoint or consume a user token secret. The operator
immutable-release setting is enabled
once by the explicit user-authenticated `scripts/enable-immutable-releases.sh`
operation. Its receipt is accepted by the generator only as an external,
immutable digest input.

## Run

```sh
GOTOOLCHAIN=auto go test ./...
go build -trimpath -o /tmp/gooo-release-transport-conformer ./cmd/gooo-release-transport-conformer
./scripts/conformance.sh "$PWD" /tmp/gooo-release-transport-conformer /tmp/gooo-release-transport-output
```

The output directory must be empty and outside the observed repository. The
conformance script checks that the source repository status is unchanged.

To bind an operator receipt after the user-level setting has been enabled:

```sh
./scripts/enable-immutable-releases.sh OWNER/REPOSITORY /tmp/operator-policy-receipt.json
```

The release workflow is dispatched on `main` with the exact merged SHA, a new
unused `0.x.y` version, and the digest from that external receipt. Failed or
non-immutable release attempts are retained; existing tags and releases are
never edited, deleted, or reused.
