# Release transport protocol v4

The `.gooo` declaration is the semantic source of truth. Its twelve activities
are bound exactly once in order: protocol declaration, external operator
receipt binding, source-run binding, workflow generation, create-response draft
ID use, existing-draft reconciliation, symbolic release-target resolution through the peeled annotated
tag target, release upload URL binding, annotated-tag target verification,
public asset-digest verification, counterexample preservation, and receipt
emission.

The generator has no product-repository write, commit, push, merge, tag, or
release authority. Those effects exist only in the generated caller workflow
or in the explicit user-authorized operator/release operations. In particular,
the ordinary Actions token is not assumed to have administration capability.

The fixed twenty scenarios deliberately include both healthy observations and
negative evidence. `UNKNOWN` is used for missing or stale evidence and retains
all six required fields. `REFUTED` is used for observed order, identity,
immutability, path, or capability contradictions. The conformance receipt
closes the denominator when these classifications match the `.gooo` contract.
Draft lookup uses the releases list endpoint because a draft can be hidden by
the releases-by-tag endpoint; an existing draft is resumed only when its tag,
source target, and assets match exactly or its assets are empty. A newly created
draft continues with the ID in the create response; it does not require
immediate list visibility. A symbolic
`target_commitish` is never treated as an exact commit; it is reconciled with
the peeled annotated tag target. Binary upload
uses only the draft detail `upload_url` after removing its template; the
api.github.com release-assets endpoint is a refuted transport path. Drafts and
tags are never deleted, moved, recreated, or force-updated.

Performance evidence is recorded by CI for compile, build, test, conformance,
and integration as integer `wall_ms` and `peak_rss_kib` pairs. A performance
improvement is not claimed without an exact matched pair; absent such a pair it
is `UNKNOWN`.
