package transport

import (
	"os"
	"path/filepath"
	"testing"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFixedContractHasExactActivitiesAndScenarios(t *testing.T) {
	root := repositoryRoot(t)
	contract, _, err := LoadContract(filepath.Join(root, ".gooo", "release-transport.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	if len(contract.Activities) != 12 || len(contract.Scenarios) != 32 || contract.Denominator != 32 {
		t.Fatalf("unexpected fixed contract shape: activities=%d scenarios=%d denominator=%d", len(contract.Activities), len(contract.Scenarios), contract.Denominator)
	}
	for i, activity := range contract.Activities {
		if activity.Name != RequiredActivities[i] || activity.Authority != "READ_ONLY" {
			t.Fatalf("activity %d = %+v", i+1, activity)
		}
	}
}

func TestSemanticIROwnsReleaseTransportBoundary(t *testing.T) {
	root := repositoryRoot(t)
	contract, raw, err := LoadContract(filepath.Join(root, ".gooo", "release-transport.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	ir, err := BuildSemanticIR(contract, raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateSemanticIR(ir); err != nil {
		t.Fatal(err)
	}
	if ir.Version.Version != "0.1.5" || ir.PreviousRelease.ReleaseID != "380375220" || ir.Terminal != "FIXED_POINT" {
		t.Fatalf("semantic IR lost release identity: %+v", ir)
	}
	if !sameStates(ir.States, RequiredStates) || !sameTransitions(ir.Transitions, RequiredTransitions) {
		t.Fatal("semantic IR state machine is not append-only and forward-only")
	}
}

func TestDecisionPrecedenceRefutesUnknownAndClosed(t *testing.T) {
	if ResolveDecision(Closed, Unknown, Refuted) != Refuted || ResolveDecision(Closed, Unknown) != Unknown || ResolveDecision(Closed) != Closed {
		t.Fatal("decision precedence is not REFUTED > UNKNOWN > CLOSED")
	}
}

func TestFixedScenariosEvaluateToDeclaredStates(t *testing.T) {
	root := repositoryRoot(t)
	contract, _, err := LoadContract(filepath.Join(root, ".gooo", "release-transport.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	results, counts, err := EvaluateFixed(contract, ReleaseWorkflow())
	if err != nil {
		t.Fatal(err)
	}
	if counts["CLOSED"] != 12 || counts["UNKNOWN"] != 4 || counts["REFUTED"] != 16 {
		t.Fatalf("unexpected decision counts: %#v", counts)
	}
	for _, result := range results {
		if result.Decision != result.Expected {
			t.Fatalf("scenario %s = %s, want %s", result.ID, result.Decision, result.Expected)
		}
		if result.Decision == Unknown && (result.Unknown == nil || !result.Unknown.Valid()) {
			t.Fatalf("scenario %s lacks six-field unknown evidence", result.ID)
		}
	}
}

func TestWorkflowIsStandardTokenDraftFirstAndAdminFree(t *testing.T) {
	workflow := ReleaseWorkflow()
	if !workflowOrderValid(workflow) || !workflowUsesStandardToken(workflow) {
		t.Fatal("workflow does not satisfy the draft-first standard-token contract")
	}
	for _, forbidden := range []string{"secrets.", "GITHUB_TOKEN", "GH_PAT", "/immutable-releases", "administration"} {
		if contains(workflow, forbidden) {
			t.Fatalf("workflow contains forbidden term %q", forbidden)
		}
	}
}

func TestGenerateWritesExactlySixCallerOwnedFiles(t *testing.T) {
	root := repositoryRoot(t)
	out := t.TempDir()
	if _, err := Generate(GenerateOptions{Root: root, Output: out}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(OutputFiles) {
		t.Fatalf("generated %d files, want %d", len(entries), len(OutputFiles))
	}
	for _, name := range OutputFiles {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("missing output %s: %v", name, err)
		}
	}
}

func TestOutputDirectoryMustBeEmptyAndOutsideSource(t *testing.T) {
	root := repositoryRoot(t)
	inside := t.TempDir()
	if err := EnsureEmptyOutput(inside, inside); err == nil {
		t.Fatal("expected source-boundary rejection")
	}
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "occupied"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureEmptyOutput(outside, root); err == nil {
		t.Fatal("expected non-empty output rejection")
	}
}

func TestUnknownEvidenceHasNoOmittedRequiredField(t *testing.T) {
	evidence := UnknownEvidence{Stage: "stage", Step: "step", Reason: "reason", UnknownClass: "DIRECT_MISSING", NextOperation: "next", BlockedBy: []string{"evidence"}}
	if !evidence.Valid() {
		t.Fatal("valid six-field evidence rejected")
	}
	evidence.BlockedBy = nil
	if evidence.Valid() {
		t.Fatal("empty blocked_by accepted")
	}
}

func contains(value, needle string) bool {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
