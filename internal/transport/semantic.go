package transport

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var semanticVersionPattern = regexp.MustCompile(`^0\.([0-9]+)\.([0-9]+)$`)

var requiredScenarioIDs = []string{
	"draft-assets-publish-immutable",
	"deterministic-replay",
	"all-asset-digests-match",
	"exact-annotated-tag-target",
	"missing-operator-immutable-policy-receipt",
	"stale-source-run",
	"missing-git-identity",
	"tag-collision",
	"publish-before-assets",
	"published-immutable-false",
	"checksum-path-mismatch",
	"user-token-secret-or-admin-endpoint-in-actions",
	"resume-existing-exact-draft-by-list-id",
	"existing-draft-target-or-assets-mismatch",
	"upload-assets-via-release-upload-url",
	"upload-assets-via-api-endpoint",
	"reconcile-symbolic-target-with-peeled-tag-target",
	"treat-symbolic-target-commitish-as-exact-commit",
	"continue-with-create-response-draft-id",
	"require-immediate-draft-list-visibility-after-create",
	"linear-forward-state-machine",
	"pre-mutation-fixture-conformance",
	"failed-attempt-preserves-objects-and-burns-version",
	"direct-main-release-target",
	"ambiguous-compare-lineage",
	"wrong-target-commitish",
	"tag-release-ordering-error",
	"missing-immutable-setting-evidence",
	"asset-manifest-count-name-size-digest-mismatch",
	"duplicate-or-burned-version-reuse",
	"mutable-published-release",
	"fixed-point-only-publication",
}

var requiredScenarioDecisions = []Decision{
	Closed, Closed, Closed, Closed, Unknown, Unknown, Unknown, Refuted,
	Refuted, Refuted, Refuted, Refuted, Closed, Refuted, Closed, Refuted,
	Closed, Refuted, Closed, Refuted, Closed, Closed, Closed, Refuted,
	Refuted, Refuted, Refuted, Unknown, Refuted, Refuted, Refuted, Closed,
}

var requiredScenarioUnknownClasses = []string{
	"", "", "", "", "DIRECT_MISSING", "STALE", "DIRECT_MISSING", "",
	"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"", "", "", "DIRECT_MISSING", "", "", "", "",
}

func BuildSemanticIR(contract Contract, sourceRaw []byte) (SemanticIR, error) {
	if err := ValidateContract(contract); err != nil {
		return SemanticIR{}, err
	}
	if len(sourceRaw) == 0 {
		return SemanticIR{}, errors.New("semantic source is empty")
	}
	authority := zeroRuntimeAuthority()
	return SemanticIR{
		Schema:          SemanticIRSchema,
		ContractID:      contract.ContractID,
		ContractDigest:  DigestBytes(sourceRaw),
		SourceDigest:    DigestBytes(sourceRaw),
		PreviousDenominator: contract.PreviousDenominator,
		AppendOnly:      contract.AppendOnly,
		Version:         contract.Version,
		PreviousRelease: contract.PreviousRelease,
		MergedPR:        contract.MergedPR,
		ExactTarget:     contract.ExactTarget,
		AnnotatedTag:    contract.AnnotatedTag,
		DraftRelease:    contract.DraftRelease,
		ExpectedAssets:  append([]ExpectedAsset(nil), contract.ExpectedAssets...),
		APIReceipts:     append([]APIReceipt(nil), contract.APIReceipts...),
		BurnedVersions:  append([]BurnedVersion(nil), contract.BurnedVersions...),
		States:          append([]ReleaseState(nil), contract.States...),
		Transitions:     append([]Transition(nil), contract.Transitions...),
		Terminal:        contract.Terminal,
		Activities:      append([]Activity(nil), contract.Activities...),
		Indicators:      append([]IndicatorSpec(nil), contract.Indicators...),
		Scenarios:       append([]ScenarioSpec(nil), contract.Scenarios...),
		Authority:       authority,
	}, nil
}

func zeroRuntimeAuthority() Authority {
	return Authority{
		RepositoryWrites:         0,
		Commits:                  0,
		Pushes:                   0,
		Merges:                   0,
		Tags:                     0,
		Releases:                 0,
		LocalGoTests:             0,
		CallerOwnedOutput:        true,
		SourceRepositoryReadOnly: true,
		CrossProjectRequiredGates: 0,
		GithubToken:              "github.token",
		OperatorRelease: OperatorReleaseAuthority{
			Workflow: "operator-release-workflow",
			AllowedMutations: []string{"annotated_tag", "draft_release", "release_asset_upload", "publish_draft"},
			SeparatedFromRuntime: true,
		},
	}
}

func ValidateSemanticIR(ir SemanticIR) error {
	if ir.Schema != SemanticIRSchema || ir.ContractID == "" || ir.ContractDigest == "" || ir.SourceDigest == "" || ir.PreviousDenominator != 20 || !ir.AppendOnly {
		return errors.New("semantic IR header is incomplete")
	}
	if ir.Terminal != "FIXED_POINT" {
		return errors.New("semantic IR requires explicit FIXED_POINT")
	}
	if !sameStates(ir.States, RequiredStates) || !sameTransitions(ir.Transitions, RequiredTransitions) {
		return errors.New("semantic IR state machine is not the fixed forward-only machine")
	}
	if len(ir.ExpectedAssets) == 0 || len(ir.APIReceipts) == 0 {
		return errors.New("semantic IR must retain expected assets and observed API receipts")
	}
	if len(ir.Scenarios) != len(requiredScenarioIDs) || len(ir.Indicators) != 12 {
		return errors.New("semantic IR denominator vectors are incomplete")
	}
	if ir.Authority.RepositoryWrites != 0 || ir.Authority.Commits != 0 || ir.Authority.Merges != 0 || ir.Authority.Tags != 0 || ir.Authority.Releases != 0 || ir.Authority.LocalGoTests != 0 || ir.Authority.CrossProjectRequiredGates != 0 {
		return errors.New("semantic IR runtime authority must be zero")
	}
	return nil
}

func ResolveDecision(decisions ...Decision) Decision {
	for _, precedence := range Precedence {
		for _, decision := range decisions {
			if decision == precedence {
				return decision
			}
		}
	}
	return Unknown
}

func ValidateReleaseDeclaration(contract Contract) error {
	if !semanticVersionPattern.MatchString(contract.Version.Version) || contract.Version.Tag != "v"+contract.Version.Version || !contract.Version.PatchOnly {
		return errors.New("release declaration must be a new 0.x.y patch version")
	}
	if contract.Version.NextVersion == "" || !semanticVersionPattern.MatchString(contract.Version.NextVersion) {
		return errors.New("release declaration must declare the next fresh patch version")
	}
	if contract.PreviousRelease.Repository == "" || contract.PreviousRelease.Version == "" || contract.PreviousRelease.Tag == "" || contract.PreviousRelease.ReleaseID == "" || contract.PreviousRelease.TagObjectSHA == "" || contract.PreviousRelease.TargetCommitSHA == "" || !contract.PreviousRelease.Immutable {
		return errors.New("previous immutable release identity is incomplete")
	}
	if len(contract.PreviousRelease.Assets) != 3 {
		return errors.New("previous immutable release identity must retain its exact three assets")
	}
	previousAssetNames := map[string]bool{}
	for _, asset := range contract.PreviousRelease.Assets {
		if asset.ID == "" || asset.Name == "" || asset.Size == "" || !isSHA256Digest(asset.Digest) || previousAssetNames[asset.Name] {
			return errors.New("previous immutable release assets must retain exact IDs, names, sizes, and digests")
		}
		previousAssetNames[asset.Name] = true
	}
	if contract.MergedPR.Repository == "" || contract.MergedPR.BaseBranch != "main" || contract.MergedPR.Number == "" || contract.MergedPR.MergeCommitSHA == "" || !contract.MergedPR.Merged {
		return errors.New("merged PR lineage declaration is incomplete")
	}
	if contract.ExactTarget.Value == "" || contract.ExactTarget.Source == "" || !contract.ExactTarget.Exact || (!isExactSHA(contract.ExactTarget.Value) && contract.ExactTarget.Source != "workflow_dispatch.expected_sha") {
		return errors.New("exact target commit declaration is incomplete")
	}
	if contract.AnnotatedTag.Tag != contract.Version.Tag || contract.AnnotatedTag.TargetCommitSHA == "" || !contract.AnnotatedTag.Annotated {
		return errors.New("annotated tag declaration is incomplete")
	}
	if contract.DraftRelease.ID == "" || contract.DraftRelease.Tag != contract.Version.Tag || !contract.DraftRelease.Draft {
		return errors.New("draft release identity declaration is incomplete")
	}
	if len(contract.ExpectedAssets) != 3 {
		return errors.New("expected asset manifest must contain exactly three assets")
	}
	assetNames := map[string]bool{}
	for _, asset := range contract.ExpectedAssets {
		if asset.Name == "" || asset.Size == "" || asset.Digest == "" || assetNames[asset.Name] {
			return errors.New("expected asset manifest entries require name, size, and digest")
		}
		assetNames[asset.Name] = true
	}
	for _, receipt := range contract.APIReceipts {
		if receipt.ID == "" || receipt.Method == "" || receipt.Endpoint == "" || receipt.Status == 0 || !isSHA256Digest(receipt.ResponseDigest) || receipt.Fixture == "" {
			return errors.New("observed API receipts require identity, request, status, digest, and fixture")
		}
	}
	return nil
}

func sameStates(left, right []ReleaseState) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameTransitions(left, right []Transition) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validateIndicators(indicators []IndicatorSpec, activities []Activity) error {
	if len(indicators) != 12 || len(activities) != 12 {
		return errors.New("indicator denominator must contain exactly twelve entries")
	}
	families := map[string]int{}
	roles := map[string]int{}
	for index, indicator := range indicators {
		if indicator.Ordinal != index+1 || indicator.ID == "" || indicator.Activity != activities[index].Name {
			return fmt.Errorf("indicator %d is not bound to its activity", index+1)
		}
		if indicator.Family != RequiredIndicatorFamilies[index/4] || indicator.Role != RequiredIndicatorRoles[index/4] {
			return fmt.Errorf("indicator %d has invalid family or role", index+1)
		}
		families[indicator.Family]++
		roles[indicator.Role]++
	}
	if families["FOUNDATION"] != 4 || families["COHERENCE"] != 4 || families["REGRESSION"] != 4 || roles["DRIVER"] != 4 || roles["OUTCOME"] != 4 || roles["GUARDRAIL"] != 4 {
		return errors.New("indicator vectors must be 4/4/4 by family and role")
	}
	return nil
}

func validateBurnedVersions(versions []BurnedVersion, current string) error {
	seen := map[string]bool{}
	for _, item := range versions {
		if item.Version == "" || item.AttemptID == "" || item.Reason == "" || item.ReceiptDigest == "" || !semanticVersionPattern.MatchString(item.Version) || seen[item.Version] {
			return errors.New("burned version ledger must be append-only and digest-bound")
		}
		if item.Version == current {
			return errors.New("current release version cannot already be burned")
		}
		seen[item.Version] = true
	}
	return nil
}

func scenarioFamilies(count int) (string, string) {
	switch {
	case count < 12:
		return "FOUNDATION", "DRIVER"
	case count < 28:
		return "COHERENCE", "OUTCOME"
	default:
		return "REGRESSION", "GUARDRAIL"
	}
}

func sortAssets(assets []ExpectedAsset) []ExpectedAsset {
	copyAssets := append([]ExpectedAsset(nil), assets...)
	sort.Slice(copyAssets, func(left, right int) bool { return copyAssets[left].Name < copyAssets[right].Name })
	return copyAssets
}

func isExactSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, char := range strings.ToLower(value) {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func isSHA256Digest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	return isExactSHA(value[len("sha256:"):])
}
