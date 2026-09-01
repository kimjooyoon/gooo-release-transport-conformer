package transport

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func EvaluateFixed(contract Contract, workflow string) ([]ScenarioResult, map[string]int, error) {
	if err := ValidateContract(contract); err != nil {
		return nil, nil, err
	}
	results := make([]ScenarioResult, 0, len(contract.Scenarios))
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	for _, spec := range contract.Scenarios {
		observation := fixedObservation(spec.ID, workflow)
		result := EvaluateScenario(spec, observation)
		if result.Decision != spec.Expected {
			return nil, nil, fmt.Errorf("scenario %s evaluated as %s, want %s", spec.ID, result.Decision, spec.Expected)
		}
		results = append(results, result)
		counts[string(result.Decision)]++
	}
	return results, counts, nil
}

func EvaluateScenario(spec ScenarioSpec, observation TransportObservation) ScenarioResult {
	result := ScenarioResult{Ordinal: spec.Ordinal, ID: spec.ID, Expected: spec.Expected}
	closed := func(reason string) ScenarioResult {
		result.Decision = Closed
		result.Reason = reason
		return result
	}
	unknown := func(evidence UnknownEvidence) ScenarioResult {
		result.Decision = Unknown
		result.Reason = evidence.Reason
		result.Unknown = &evidence
		return result
	}
	refuted := func(kind, reason string) ScenarioResult {
		result.Decision = Refuted
		result.Reason = reason
		result.Refutation = &Refutation{Kind: kind, Reason: reason}
		return result
	}

	switch spec.ID {
	case "draft-assets-publish-immutable":
		if workflowOrderValid(observation.Workflow) && observation.DraftCreated && observation.AssetsUploaded && observation.Published && observation.PublishedImmutable {
			return closed("draft created, every asset uploaded, then published and observed immutable")
		}
		return refuted("ORDER_OR_IMMUTABILITY_MISMATCH", "draft-first order, complete upload, or immutable publication was not observed")
	case "deterministic-replay":
		if observation.DeterministicReplay {
			return closed("two identical source/contract evaluations produced the same workflow and receipt bytes")
		}
		return refuted("NON_DETERMINISTIC_REPLAY", "replay bytes differed for the same source and contract")
	case "all-asset-digests-match":
		if allAssetDigestsMatch(observation.Assets) {
			return closed("all published asset digests match their checksum paths")
		}
		return refuted("ASSET_DIGEST_MISMATCH", "at least one public asset digest differs from its checksum binding")
	case "exact-annotated-tag-target":
		if observation.Tag.Annotated && observation.Tag.ObjectType == "tag" && observation.Tag.TargetCommit == observation.Tag.ExpectedCommit && observation.Tag.ObjectSHA != "" && !observation.Tag.Collision {
			return closed("annotated tag object resolves to the exact expected commit")
		}
		return refuted("TAG_TARGET_MISMATCH", "the public tag was not an annotated tag object targeting the expected commit")
	case "missing-operator-immutable-policy-receipt":
		if observation.OperatorPolicy == nil {
			return unknown(UnknownEvidence{Stage: "OPERATOR_POLICY", Step: "bind-immutable-policy-receipt", Reason: "external immutable policy receipt is absent", UnknownClass: "DIRECT_MISSING", NextOperation: "provide-external-immutable-policy-receipt", BlockedBy: []string{"operator-api-receipt"}})
		}
		if observation.OperatorPolicy.Immutable && observation.OperatorPolicy.Enabled && strings.HasPrefix(observation.OperatorPolicy.Digest, "sha256:") {
			return closed("external immutable policy receipt is present and digest-bound")
		}
		return refuted("INVALID_OPERATOR_POLICY_RECEIPT", "the supplied operator receipt is not an enabled immutable digest")
	case "stale-source-run":
		if observation.SourceRun.Stale || observation.SourceRun.SourceDigest != observation.SourceRun.CurrentSourceDigest {
			return unknown(UnknownEvidence{Stage: "SOURCE_RUN", Step: "compare-source-run-digest", Reason: "source run does not identify the current source bytes", UnknownClass: "STALE", NextOperation: "rerun-on-current-merged-commit", BlockedBy: []string{"source-run-digest", "current-source-digest"}})
		}
		return closed("source run is bound to the current source digest")
	case "missing-git-identity":
		if observation.GitIdentity.Name == "" || observation.GitIdentity.Email == "" {
			return unknown(UnknownEvidence{Stage: "TAG", Step: "bind-git-identity", Reason: "annotated tag author identity is absent", UnknownClass: "DIRECT_MISSING", NextOperation: "provide-explicit-git-identity", BlockedBy: []string{"git-user-name", "git-user-email"}})
		}
		return closed("annotated tag identity is explicit")
	case "tag-collision":
		if observation.Tag.Collision || observation.Tag.Exists {
			return refuted("TAG_COLLISION", "release refuses to reuse an existing tag or release")
		}
		return closed("tag and release names are unused")
	case "publish-before-assets":
		if observation.Published && !observation.AssetsUploaded {
			return refuted("PUBLISH_BEFORE_ASSETS", "publish occurred before all release assets were uploaded")
		}
		return closed("publication follows complete asset upload")
	case "published-immutable-false":
		if observation.Published && !observation.PublishedImmutable {
			return refuted("PUBLISHED_IMMUTABLE_FALSE", "public release verification observed immutable=false")
		}
		return closed("public release verification observed immutable=true")
	case "checksum-path-mismatch":
		if observation.ChecksumPathMismatch || !allAssetDigestsMatch(observation.Assets) {
			return refuted("CHECKSUM_PATH_MISMATCH", "checksum verification names a path different from the uploaded asset")
		}
		return closed("checksum paths and uploaded asset names match")
	case "user-token-secret-or-admin-endpoint-in-actions":
		if observation.ActionHasUserTokenSecret || observation.ActionHasAdminEndpoint || !workflowUsesStandardToken(observation.Workflow) {
			return refuted("UNAUTHORIZED_ACTION_CAPABILITY", "workflow contains a user-token secret or an administration endpoint")
		}
		return closed("workflow is limited to standard github.token and public release verification")
	case "resume-existing-exact-draft-by-list-id":
		if observation.ReconciledDraft && observation.DraftID != "" && observation.DraftTag != "" && observation.DraftSourceTarget == observation.Tag.ExpectedCommit && (observation.ExistingDraftAssets == "empty" || observation.ExistingDraftAssets == "exact") {
			return closed("exact draft was found through the release list API and resumed by immutable release ID")
		}
		return refuted("DRAFT_RESUME_BINDING_MISMATCH", "existing draft was not resumed by exact list-derived ID and source target")
	case "existing-draft-target-or-assets-mismatch":
		if observation.ExistingDraftMismatch || observation.DraftSourceTarget != observation.Tag.ExpectedCommit || (observation.ExistingDraftAssets != "empty" && observation.ExistingDraftAssets != "exact") {
			return refuted("EXISTING_DRAFT_MISMATCH", "existing draft target or asset names/digests differ, so the workflow must fail closed")
		}
		return closed("existing draft target and assets match")
	default:
		return refuted("UNDECLARED_SCENARIO", "scenario is not part of the fixed transport denominator")
	}
}

func fixedObservation(id, workflow string) TransportObservation {
	assets := []AssetObservation{
		{Path: "source.tar.gz", Digest: "sha256:source-asset", ChecksumPath: "source.tar.gz"},
		{Path: "conformance-receipt.json", Digest: "sha256:receipt-asset", ChecksumPath: "conformance-receipt.json"},
		{Path: "transport-manifest.json", Digest: "sha256:manifest-asset", ChecksumPath: "transport-manifest.json"},
	}
	observation := TransportObservation{
		Workflow:       workflow,
		OperatorPolicy: &OperatorPolicyReceipt{Enabled: true, Immutable: true, Digest: "sha256:operator-policy"},
		SourceRun:      SourceRunObservation{SourceDigest: "sha256:source-run", CurrentSourceDigest: "sha256:source-run"},
		GitIdentity:    GitIdentityObservation{Name: "github-actions[bot]", Email: "41898282+github-actions[bot]@users.noreply.github.com"},
		Assets:         assets,
		Tag:            TagObservation{Exists: false, Annotated: true, ObjectType: "tag", ObjectSHA: "sha256:annotated-tag-object", TargetCommit: "merge-commit", ExpectedCommit: "merge-commit"},
		DraftID:        "draft-123",
		DraftTag:       "v0.1.1",
		DraftSourceTarget: "merge-commit",
		ExistingDraftAssets: "empty",
		ReconciledDraft: true,
		DraftCreated:   true, AssetsUploaded: true, Published: true, PublishedImmutable: true,
		DeterministicReplay: true,
	}
	switch id {
	case "missing-operator-immutable-policy-receipt":
		observation.OperatorPolicy = nil
	case "stale-source-run":
		observation.SourceRun.CurrentSourceDigest = "sha256:new-source"
		observation.SourceRun.Stale = true
	case "missing-git-identity":
		observation.GitIdentity = GitIdentityObservation{}
	case "tag-collision":
		observation.Tag.Exists = true
		observation.Tag.Collision = true
	case "publish-before-assets":
		observation.AssetsUploaded = false
	case "published-immutable-false":
		observation.PublishedImmutable = false
	case "checksum-path-mismatch":
		observation.ChecksumPathMismatch = true
	case "user-token-secret-or-admin-endpoint-in-actions":
		observation.ActionHasUserTokenSecret = true
	case "resume-existing-exact-draft-by-list-id":
		observation.ExistingDraftAssets = "exact"
	case "existing-draft-target-or-assets-mismatch":
		observation.DraftSourceTarget = "different-commit"
		observation.ExistingDraftAssets = "mismatch"
		observation.ExistingDraftMismatch = true
	}
	return observation
}

func workflowOrderValid(workflow string) bool {
	markers := []string{
		"Create draft release before assets",
		"Reconcile existing draft by list ID",
		"Upload every release asset",
		"Publish release after all uploads",
		"Verify public immutable release",
	}
	last := -1
	for _, marker := range markers {
		index := strings.Index(workflow, marker)
		if index <= last {
			return false
		}
		last = index
	}
	return true
}

func allAssetDigestsMatch(assets []AssetObservation) bool {
	if len(assets) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, asset := range assets {
		if asset.Path == "" || asset.ChecksumPath != asset.Path || !strings.HasPrefix(asset.Digest, "sha256:") || seen[asset.Path] {
			return false
		}
		seen[asset.Path] = true
	}
	return true
}

func workflowUsesStandardToken(workflow string) bool {
	forbidden := []string{"secrets.", "GITHUB_TOKEN", "GH_PAT", "PAT", "/immutable-releases", "administration", "admin:repo"}
	for _, term := range forbidden {
		if strings.Contains(workflow, term) {
			return false
		}
	}
	return strings.Contains(workflow, "GH_TOKEN: ${{ github.token }}")
}

func InventoryRoot(root string) (Inventory, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Inventory{}, err
	}
	var inventory Inventory
	inventory.RootREADMEExcluded = true
	inventory.ExcludedDirectories = []string{".git", "temp", "cache", "vendor", "toolchain"}
	excluded := map[string]bool{".git": true, "temp": true, "cache": true, "vendor": true, "toolchain": true}
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && info.IsDir() && excluded[info.Name()] {
			return filepath.SkipDir
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			inventory.DescendantDirs++
			return nil
		}
		if !info.Mode().IsRegular() || relative == "README.md" {
			return nil
		}
		inventory.RegularFiles++
		switch strings.ToLower(filepath.Ext(info.Name())) {
		case ".go":
			inventory.GoFiles++
			inventory.GoPhysicalLines += physicalLines(path)
		case ".gooo":
			inventory.GoooFiles++
			inventory.GoooPhysicalLines += physicalLines(path)
		}
		return nil
	})
	return inventory, err
}

func physicalLines(path string) int {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return 0
	}
	count := strings.Count(string(raw), "\n")
	if raw[len(raw)-1] != '\n' {
		count++
	}
	return count
}

func EnsureEmptyOutput(output, root string) error {
	if output == "" {
		return errors.New("caller-owned output directory is required")
	}
	absoluteOutput, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if within(absoluteRoot, absoluteOutput) {
		return errors.New("caller-owned output must be outside the observed source repository")
	}
	if err := os.MkdirAll(absoluteOutput, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(absoluteOutput)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty: %s", absoluteOutput)
	}
	return nil
}

func within(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && (relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))))
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
