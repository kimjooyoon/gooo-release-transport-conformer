# Release transport protocol v6

The `.gooo` file is the semantic authority. The generator parses it into a
semantic IR, the evaluator classifies the declared cases, and the projector
emits the machine and human artifacts. JSON fixtures are observations; they
cannot widen the contract.

## Contract boundary

The declaration owns the current patch version, the previous immutable release
identity, merged PR lineage, exact target commit binding, annotated tag object
and peeled target, create-response draft identity, expected asset entries
(name, size, digest), observed API receipt identities, the append-only burned
version ledger, and the state machine. The v6 denominator is append-only: all
twenty v5 cases remain and twelve new transport cases are added, for exactly
thirty-two cases. There are twelve indicator vectors.

The only valid state path is:

`PRECHECK -> TAGGED -> DRAFT_CREATED -> ASSETS_UPLOADED -> ASSETS_AUDITED -> PUBLISHED_IMMUTABLE`

Transitions are forward-only and each state is recorded once. Publication is
valid only at the explicit `FIXED_POINT` terminal. The evaluator resolves
`REFUTED > UNKNOWN > CLOSED`; an unknown record always contains the six fields
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.

## Mutation protocol

The generated operator workflow has a separate authority from the generator
runtime. Before mutation it conforms the tag payload, draft payload, lineage,
asset manifest shape, state transitions, and API receipt fixtures. It then
reads the merged PR API record directly; it does not infer lineage from the
compare API. The release target must be the exact merged commit, never `main`
or another symbolic ref. The previous immutable release is checked by release
ID, tag, annotated tag object, peeled commit, and `immutable=true`.

The workflow creates one annotated tag, creates one draft release before any
asset, continues with the create-response draft ID, obtains the draft detail
`upload_url`, uploads binary assets only through the resulting
`uploads.github.com` endpoint, audits exact asset count/name/size/digest, and
publishes once. It never deletes, edits, force-updates, retags, or reuses a
public object.

If any step fails after mutation starts, the always-run caller-owned artifact
contains an `OPERATIONAL_REFUTED` receipt with the exact `mutation_started`
boolean and the created tag ref, annotated tag object, draft release, published
release, and asset IDs observed so far. The attempted version is burned and
the only next operation is a fresh patch version. Pre-mutation refusals retain
the refusal receipt without claiming public object creation.

The direct-main target, ambiguous compare result, wrong target commitish,
tag/release ordering error, asset count/name/size/digest mismatch, duplicate
or burned version reuse, and mutable published release cases are `REFUTED`.
Missing immutable-setting evidence is `UNKNOWN` with the six-field causal
record. A disabled or contradictory setting receipt is `REFUTED`.

## Authority and evidence

The Go runtime has exact zero authority for repository writes, commits, merges,
pushes, tags, releases, local product tests, and cross-project required gates.
It writes only the caller-owned output directory. The operator release workflow
is declared separately and uses the standard `github.token`; PR CI retains
read-only contents permission and performs zero release mutations.

The generated receipt exposes exact case and indicator vectors, with the
`FOUNDATION`/`COHERENCE`/`REGRESSION` and `DRIVER`/`OUTCOME`/`GUARDRAIL`
partitions. It does not emit an aggregate score or percentage.
