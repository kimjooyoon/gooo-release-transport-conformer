# Release transport protocol v1

The `.gooo` declaration is the semantic source of truth. Its eight activities
are bound exactly once in order: protocol declaration, external operator
receipt binding, source-run binding, workflow generation, annotated-tag target
verification, public asset-digest verification, counterexample preservation,
and receipt emission.

The generator has no product-repository write, commit, push, merge, tag, or
release authority. Those effects exist only in the generated caller workflow
or in the explicit user-authorized operator/release operations. In particular,
the ordinary Actions token is not assumed to have administration capability.

The fixed scenarios deliberately include both healthy observations and
negative evidence. `UNKNOWN` is used for missing or stale evidence and retains
all six required fields. `REFUTED` is used for observed order, identity,
immutability, path, or capability contradictions. The conformance receipt
closes the denominator when these classifications match the `.gooo` contract.

Performance evidence is recorded by CI for compile, build, test, conformance,
and integration as integer `wall_ms` and `peak_rss_kib` pairs. A performance
improvement is not claimed without an exact matched pair; absent such a pair it
is `UNKNOWN`.
