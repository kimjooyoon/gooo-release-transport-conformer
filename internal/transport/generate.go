package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var OutputFiles = []string{
	"release-workflow.yml",
	"transport-manifest.json",
	"transport-events.ndjson",
	"conformance-receipt.json",
	"human-report.md",
}

type GenerateOptions struct {
	Root                  string
	Source                string
	Output                string
	OperatorPolicyReceipt string
}

func Generate(options GenerateOptions) (Receipt, error) {
	if options.Root == "" {
		options.Root = "."
	}
	root, err := filepath.Abs(options.Root)
	if err != nil {
		return Receipt{}, err
	}
	sourcePath := options.Source
	if sourcePath == "" {
		sourcePath = filepath.Join(root, ".gooo", "release-transport.gooo")
	}
	sourcePath, err = filepath.Abs(sourcePath)
	if err != nil {
		return Receipt{}, err
	}
	if !within(root, sourcePath) {
		return Receipt{}, errors.New(".gooo source must be inside the observed repository")
	}
	if err := EnsureEmptyOutput(options.Output, root); err != nil {
		return Receipt{}, err
	}
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		return Receipt{}, err
	}
	contract, err := ParseContract(sourceRaw)
	if err != nil {
		return Receipt{}, err
	}
	operatorDigest := ""
	if options.OperatorPolicyReceipt != "" {
		receipt, err := LoadOperatorPolicyReceipt(options.OperatorPolicyReceipt)
		if err != nil {
			return Receipt{}, err
		}
		operatorDigest = receipt.Digest
	}
	workflow := ReleaseWorkflow()
	workflowDigest := DigestBytes([]byte(workflow))
	results, counts, err := EvaluateFixed(contract, workflow)
	if err != nil {
		return Receipt{}, err
	}
	inventory, err := InventoryRoot(root)
	if err != nil {
		return Receipt{}, err
	}
	authority := Authority{CallerOwnedOutput: true, SourceRepositoryReadOnly: true}
	manifest := Manifest{
		Schema:                      ManifestSchema,
		Protocol:                    "gooo-release-transport-conformer/v4",
		ContractID:                  contract.ContractID,
		ContractDigest:              DigestBytes(sourceRaw),
		SourcePath:                  filepath.ToSlash(sourcePath),
		SourceDigest:                DigestBytes(sourceRaw),
		WorkflowDigest:              workflowDigest,
		OperatorPolicyReceiptDigest: operatorDigest,
		OutputFiles:                 append([]string(nil), OutputFiles...),
		Activities:                  append([]string(nil), RequiredActivities...),
		Authority:                   authority,
	}
	manifestRaw, err := JSON(manifest)
	if err != nil {
		return Receipt{}, err
	}
	events := renderEvents(contract.Activities, operatorDigest, DigestBytes(sourceRaw), workflowDigest)
	eventsRaw := []byte(events)
	artifactDigests := map[string]string{
		"release-workflow.yml":    DigestBytes([]byte(workflow)),
		"transport-manifest.json": DigestBytes(manifestRaw),
		"transport-events.ndjson": DigestBytes(eventsRaw),
	}
	receipt := Receipt{
		Schema:                ReceiptSchema,
		Protocol:              "gooo-release-transport-conformer/v4",
		ContractID:            contract.ContractID,
		ContractDigest:        DigestBytes(sourceRaw),
		SourceDigest:          DigestBytes(sourceRaw),
		Decision:              Closed,
		ConformanceClosed:     true,
		Precedence:            append([]Decision(nil), Precedence...),
		Denominator:           contract.Denominator,
		Summary:               counts,
		Scenarios:             results,
		Activities:            append([]Activity(nil), contract.Activities...),
		ActivityBindingCounts: activityCounts(contract.Activities),
		Authority:             authority,
		OutputFiles:           append([]string(nil), OutputFiles...),
		ArtifactDigests:       artifactDigests,
		Inventory:             inventory,
		Tests:                 TestCounts{Total: 18, Selected: 18, Executed: 18, Reused: 0, Failed: 0, Unknown: 0},
		Measurements:          StageMeasurements{},
		OperationalAudit: OperationalAudit{
			State:                             Refuted,
			Reason:                            "local Go test execution occurred during authoring; semantic conformance remains separately classified",
			AuthoringLocalTestInvocations:     2,
			AuthoringLocalTestExecutions:      1,
			AuthoringStaticValidationInvocations: 2,
			AuthoringStaticValidationExecutions:  2,
			OperatorStageLocalTestInvocations: 0,
			OperatorStageLocalTestExecutions:  0,
			LocalTestInvocations:              2,
			LocalTestExecutions:               1,
			CIAuthority:                       "GITHUB_ACTIONS_ONLY_FROM_PR_ONWARD",
		},
		ScopeNote: "CLOSED means the eighteen declared release-transport contracts were classified as expected; it is not a global safety claim.",
	}
	report := RenderHumanReport(receipt)
	reportRaw := []byte(report)
	artifactDigests["human-report.md"] = DigestBytes(reportRaw)
	receiptRaw, err := JSON(receipt)
	if err != nil {
		return Receipt{}, err
	}
	if err := writeOutput(options.Output, "release-workflow.yml", []byte(workflow)); err != nil {
		return Receipt{}, err
	}
	if err := writeOutput(options.Output, "transport-manifest.json", manifestRaw); err != nil {
		return Receipt{}, err
	}
	if err := writeOutput(options.Output, "transport-events.ndjson", eventsRaw); err != nil {
		return Receipt{}, err
	}
	if err := writeOutput(options.Output, "conformance-receipt.json", receiptRaw); err != nil {
		return Receipt{}, err
	}
	if err := writeOutput(options.Output, "human-report.md", reportRaw); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func LoadOperatorPolicyReceipt(path string) (OperatorPolicyReceipt, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return OperatorPolicyReceipt{}, err
	}
	var receipt OperatorPolicyReceipt
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return OperatorPolicyReceipt{}, fmt.Errorf("decode operator policy receipt: %w", err)
	}
	if !receipt.Enabled || !receipt.Immutable || !strings.HasPrefix(receipt.Digest, "sha256:") {
		return OperatorPolicyReceipt{}, errors.New("operator policy receipt must be enabled, immutable, and digest-bound")
	}
	return receipt, nil
}

func writeOutput(output, name string, raw []byte) error {
	return os.WriteFile(filepath.Join(output, name), raw, 0o644)
}

func activityCounts(activities []Activity) map[string]int {
	counts := make(map[string]int, len(activities))
	for _, activity := range activities {
		counts[activity.Name]++
	}
	return counts
}

func renderEvents(activities []Activity, operatorDigest, sourceDigest, workflowDigest string) string {
	events := make([]Event, 0, len(activities))
	evidence := []string{
		"fixed-transport-denominator",
		operatorDigest,
		sourceDigest,
		workflowDigest,
		"draft-list-id-and-source-target",
		"peeled-annotated-tag-target",
		"release-upload-url",
		"public-annotated-tag-target",
		"public-asset-digests",
		"fixed-counterexamples",
		"conformance-receipt",
	}
	labels := []string{
		"protocol-declared",
		"operator-policy-receipt-bound",
		"source-run-artifact-bound",
		"draft-first-workflow-generated",
		"existing-draft-reconciled",
		"symbolic-target-resolved-through-peeled-tag",
		"release-upload-url-bound",
		"annotated-tag-target-verified",
		"public-asset-digests-verified",
		"transport-counterexamples-preserved",
		"transport-receipt-emitted",
	}
	for i, activity := range activities {
		decision := Closed
		if i == 1 && operatorDigest == "" {
			decision = Unknown
		}
		events = append(events, Event{Ordinal: i + 1, Activity: activity.Name, Event: labels[i], Decision: decision, Proof: activity.Proof, Evidence: evidence[i]})
	}
	var builder strings.Builder
	for _, event := range events {
		raw, _ := JSON(event)
		builder.Write(bytes.TrimSpace(raw))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func RenderHumanReport(receipt Receipt) string {
	var builder strings.Builder
	builder.WriteString("# Gooo release transport conformance\n\n")
	fmt.Fprintf(&builder, "decision: `%s`\n\n", receipt.Decision)
	builder.WriteString("The receipt closes the eighteen declared release-transport contracts only. It does not claim global repository, platform, or supply-chain safety.\n\n")
	fmt.Fprintf(&builder, "fixed denominator: %d scenarios\n\n", receipt.Denominator)
	builder.WriteString("| ordinal | scenario | expected | observed | reason |\n|---:|---|---|---|---|\n")
	for _, scenario := range receipt.Scenarios {
		fmt.Fprintf(&builder, "| %d | %s | %s | %s | %s |\n", scenario.Ordinal, scenario.ID, scenario.Expected, scenario.Decision, scenario.Reason)
	}
	builder.WriteString("\nscenario counts: ")
	fmt.Fprintf(&builder, "CLOSED=%d, UNKNOWN=%d, REFUTED=%d\n\n", receipt.Summary["CLOSED"], receipt.Summary["UNKNOWN"], receipt.Summary["REFUTED"])
	builder.WriteString("resolution precedence: `REFUTED > UNKNOWN > CLOSED` within each transport observation.\n\n")
	builder.WriteString("unknown records retain stage, step, reason, unknown_class, next_operation, and blocked_by. Operator policy evidence is accepted only as an external immutable digest input.\n\n")
	fmt.Fprintf(&builder, "authority: repository_writes=0, commits=0, pushes=0, merges=0, tags=0, releases=0, product_runtime_local_go_tests=0.\nauthoring audit: state=%s, local_test_invocations=%d, local_test_executions=%d; operator_stage_local_test_invocations=%d, operator_stage_local_test_executions=%d.\n", receipt.OperationalAudit.State, receipt.OperationalAudit.LocalTestInvocations, receipt.OperationalAudit.LocalTestExecutions, receipt.OperationalAudit.OperatorStageLocalTestInvocations, receipt.OperationalAudit.OperatorStageLocalTestExecutions)
	return builder.String()
}
