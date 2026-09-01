package transport

// ReleaseWorkflow is deliberately a plain, standard GitHub Actions workflow.
// The generator owns this source so callers do not have to hand-maintain a
// collection of subtly different release YAML files.
func ReleaseWorkflow() string {
	return `name: gooo release transport

"on":
  workflow_dispatch:
    inputs:
      expected_sha:
        description: Exact merged main commit to release
        required: true
        type: string
      release_version:
        description: New unused 0.x.y version; failed releases are never reused
        required: true
        default: 0.1.0
        type: string
      operator_policy_receipt_digest:
        description: External immutable operator-policy receipt digest
        required: true
        type: string

permissions:
  contents: write

jobs:
  release:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-24.04
    env:
      EXPECTED_SHA: ${{ inputs.expected_sha }}
      RELEASE_VERSION: ${{ inputs.release_version }}
      OPERATOR_POLICY_RECEIPT_DIGEST: ${{ inputs.operator_policy_receipt_digest }}
      GH_TOKEN: ${{ github.token }}
    steps:
      - name: Checkout exact merged commit
        uses: actions/checkout@v5
        with:
          ref: main
          fetch-depth: 0

      - name: Install Go 1.27
        uses: actions/setup-go@v6
        with:
          go-version: 1.27.0
          cache: false

      - name: Verify exact source and external operator receipt binding
        shell: bash
        run: |
          set -Eeuo pipefail
          test "$(git rev-parse HEAD)" = "$EXPECTED_SHA"
          test "$(go env GOVERSION)" = "go1.27.0"
          case "$OPERATOR_POLICY_RECEIPT_DIGEST" in
            sha256:*) ;;
            *) echo 'operator policy receipt must be supplied as an external immutable digest' >&2; exit 1 ;;
          esac

      - name: Refuse tag and release reuse
        shell: bash
        run: |
          set -Eeuo pipefail
          case "$RELEASE_VERSION" in
            0.*.*) ;;
            *) echo 'release_version must be 0.x.y' >&2; exit 1 ;;
          esac
          RELEASE_TAG="v$RELEASE_VERSION"
          echo "RELEASE_TAG=$RELEASE_TAG" >> "$GITHUB_ENV"
          if git ls-remote --exit-code --tags origin "refs/tags/$RELEASE_TAG" >/dev/null 2>&1; then
            echo "refusing to reuse existing tag $RELEASE_TAG" >&2
            exit 1
          fi
          if gh api "repos/$GITHUB_REPOSITORY/releases/tags/$RELEASE_TAG" >/dev/null 2>&1; then
            echo "refusing to reuse existing release $RELEASE_TAG" >&2
            exit 1
          fi

      - name: Create annotated tag exactly once
        shell: bash
        run: |
          set -Eeuo pipefail
          git config user.name 'github-actions[bot]'
          git config user.email '41898282+github-actions[bot]@users.noreply.github.com'
          git tag -a "$RELEASE_TAG" "$EXPECTED_SHA" -m "gooo release transport $RELEASE_TAG"
          git push origin "refs/tags/$RELEASE_TAG"
          tag_ref=$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/$RELEASE_TAG")
          test "$(jq -r '.object.type' <<<"$tag_ref")" = tag
          TAG_OBJECT_SHA=$(jq -r '.object.sha' <<<"$tag_ref")
          tag_body=$(gh api "repos/$GITHUB_REPOSITORY/git/tags/$TAG_OBJECT_SHA")
          jq -e --arg expected "$EXPECTED_SHA" '.object.type == "commit" and .object.sha == $expected' <<<"$tag_body" >/dev/null
          echo "TAG_OBJECT_SHA=$TAG_OBJECT_SHA" >> "$GITHUB_ENV"

      - name: Create draft release before assets
        shell: bash
        run: |
          set -Eeuo pipefail
          gh release create "$RELEASE_TAG" --verify-tag --draft --title "gooo release transport $RELEASE_TAG" \
            --notes "Draft release; assets must upload before publication."

      - name: Assemble release assets and checksums
        shell: bash
        run: |
          set -Eeuo pipefail
          RELEASE_DIR="$RUNNER_TEMP/gooo-release-transport-$RELEASE_TAG"
          mkdir -p "$RELEASE_DIR"
          git archive --format=tar.gz --prefix="gooo-release-transport-conformer-$RELEASE_TAG/" "$EXPECTED_SHA" > "$RELEASE_DIR/source-$RELEASE_TAG.tar.gz"
          go test ./... > "$RELEASE_DIR/go-test.txt"
          go build -trimpath -o "$RELEASE_DIR/gooo-release-transport-conformer" ./cmd/gooo-release-transport-conformer
          ./scripts/conformance.sh "$GITHUB_WORKSPACE" "$RELEASE_DIR/gooo-release-transport-conformer" "$RELEASE_DIR/conformance"
          tar -czf "$RELEASE_DIR/evidence-$RELEASE_TAG.tar.gz" -C "$RELEASE_DIR" conformance
          (cd "$RELEASE_DIR" && sha256sum "source-$RELEASE_TAG.tar.gz" "evidence-$RELEASE_TAG.tar.gz" > SHA256SUMS)
          echo "RELEASE_DIR=$RELEASE_DIR" >> "$GITHUB_ENV"

      - name: Upload every release asset
        shell: bash
        run: |
          set -Eeuo pipefail
          test "$(grep -c '^[0-9a-f]\{64\}  ' "$RELEASE_DIR/SHA256SUMS")" = 2
          gh release upload "$RELEASE_TAG" "$RELEASE_DIR/source-$RELEASE_TAG.tar.gz" "$RELEASE_DIR/evidence-$RELEASE_TAG.tar.gz" "$RELEASE_DIR/SHA256SUMS"

      - name: Publish release after all uploads
        shell: bash
        run: |
          set -Eeuo pipefail
          assets=$(gh release view "$RELEASE_TAG" --json assets --jq '.assets | length')
          test "$assets" = 3
          gh release edit "$RELEASE_TAG" --draft=false

      - name: Verify public immutable release, tag target, and asset digests
        shell: bash
        run: |
          set -Eeuo pipefail
          release_json=$(gh api "repos/$GITHUB_REPOSITORY/releases/tags/$RELEASE_TAG")
          jq -e --arg tag "$RELEASE_TAG" '.tag_name == $tag and .draft == false and .prerelease == false and .immutable == true and (.assets | length) == 3' <<<"$release_json" >/dev/null
          test "$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/$RELEASE_TAG" --jq '.object.type')" = tag
          test "$(gh api "repos/$GITHUB_REPOSITORY/git/ref/tags/$RELEASE_TAG" --jq '.object.sha')" = "$TAG_OBJECT_SHA"
          jq -e --arg expected "$EXPECTED_SHA" '.object.type == "commit" and .object.sha == $expected' <(gh api "repos/$GITHUB_REPOSITORY/git/tags/$TAG_OBJECT_SHA") >/dev/null
          while read -r digest path; do
            expected="sha256:$digest"
            actual=$(jq -r --arg path "$(basename "$path")" '.assets[] | select(.name == $path) | .digest' <<<"$release_json")
            test "$actual" = "$expected"
          done < "$RELEASE_DIR/SHA256SUMS"
`
}
