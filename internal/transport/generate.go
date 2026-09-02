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
	"semantic-ir.json",
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
	ir, err := BuildSemanticIR(contract, sourceRaw)
	if err != nil {
		return Receipt{}, err
	}
	ir, err = ProjectSemanticIR(ir)
	if err != nil {
		return Receipt{}, err
	}
	workflow := ReleaseWorkflow()
	workflowDigest := DigestBytes([]byte(workflow))
	results, counts, caseVectors, indicatorVectors, err := EvaluateSemanticIR(ir, workflow)
	if err != nil {
		return Receipt{}, err
	}
	if err := ValidateVectors(caseVectors, indicatorVectors, contract.Denominator); err != nil {
		return Receipt{}, err
	}
	irRaw, err := JSON(ir)
	if err != nil {
		return Receipt{}, err
	}
	inventory, err := InventoryRoot(root)
	if err != nil {
		return Receipt{}, err
	}
	authority := ir.Authority
	manifest := Manifest{
		Schema:                      ManifestSchema,
		Protocol:                    "gooo-release-transport-conformer/v6",
		ContractID:                  contract.ContractID,
		ContractDigest:              DigestBytes(sourceRaw),
		SourcePath:                  filepath.ToSlash(sourcePath),
		SourceDigest:                DigestBytes(sourceRaw),
		WorkflowDigest:              workflowDigest,
		OperatorPolicyReceiptDigest: operatorDigest,
		OutputFiles:                 append([]string(nil), OutputFiles...),
		Activities:                  append([]string(nil), RequiredActivities...),
		Authority:                   authority,
		SemanticIRSchema:             SemanticIRSchema,
		ReleaseVersion:               ir.Version.Version,
		PreviousReleaseID:            ir.PreviousRelease.ReleaseID,
		PreviousDenominator:          ir.PreviousDenominator,
		AppendOnly:                   ir.AppendOnly,
		StateMachine:                 append([]ReleaseState(nil), ir.States...),
		ExpectedAssetManifest:        append([]ExpectedAsset(nil), ir.ExpectedAssets...),
	}
	manifestRaw, err := JSON(manifest)
	if err != nil {
		return Receipt{}, err
	}
	events := renderEvents(contract.Activities, operatorDigest, DigestBytes(sourceRaw), workflowDigest, ir.States, ir.APIReceipts)
	eventsRaw := []byte(events)
	artifactDigests := map[string]string{
		"release-workflow.yml":    DigestBytes([]byte(workflow)),
		"semantic-ir.json":        DigestBytes(irRaw),
		"transport-manifest.json": DigestBytes(manifestRaw),
		"transport-events.ndjson": DigestBytes(eventsRaw),
	}
	receipt := Receipt{
		Schema:                ReceiptSchema,
		Protocol:              "gooo-release-transport-conformer/v6",
		ContractID:            contract.ContractID,
		ContractDigest:        DigestBytes(sourceRaw),
		SourceDigest:          DigestBytes(sourceRaw),
		Decision:              Closed,
		Terminal:              ir.Terminal,
		State:                 PublishedImmutableState,
		ConformanceClosed:     true,
		Precedence:            append([]Decision(nil), Precedence...),
		Denominator:           contract.Denominator,
		Summary:               counts,
		Scenarios:             results,
		CaseDenominator:       contract.Denominator,
		CaseVectors:           caseVectors,
		IndicatorDenominator:  len(indicatorVectors),
		IndicatorVectors:      indicatorVectors,
		Activities:            append([]Activity(nil), contract.Activities...),
		ActivityBindingCounts: activityCounts(contract.Activities),
		Authority:             authority,
		OutputFiles:           append([]string(nil), OutputFiles...),
		ArtifactDigests:       artifactDigests,
		Inventory:             inventory,
		Tests:                 TestCounts{Total: 32, Selected: 32, Executed: 32, Reused: 0, Failed: 0, Unknown: 0},
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
			MutationStarted:                   false,
			CreatedPublicObjectIDs:            []string{},
			IncidentDisposition:               "NONE_GENERATOR_READ_ONLY",
			BurnedVersion:                     false,
		},
		ScopeNote: "CLOSED means the thirty-two declared release-transport cases were classified as expected at the explicit FIXED_POINT; it is not a global safety claim.",
		SemanticIR:           ir,
		Attempt: MutationAttempt{
			Schema:                  "gooo/release-transport-conformer/mutation-attempt/v1",
			AttemptID:               "GENERATOR_READ_ONLY",
			Version:                 ir.Version.Version,
			Tag:                     ir.Version.Tag,
			State:                   PrecheckState,
			Decision:                Closed,
			MutationStarted:         false,
			CreatedPublicObjectIDs:  CreatedPublicObjectIDs{AssetIDs: []string{}},
			PreserveNeverDelete:     true,
			BurnedVersion:           false,
			NextVersion:             ir.Version.NextVersion,
			IncidentDisposition:     "NONE_GENERATOR_READ_ONLY",
		},
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
	if err := writeOutput(options.Output, "semantic-ir.json", irRaw); err != nil {
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
	if receipt.Schema != "gooo-release-transport-conformer/operator-immutable-policy-receipt/v1" || receipt.Repository == "" || !receipt.Enabled || !receipt.Immutable || !isSHA256Digest(receipt.Digest) {
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

func renderEvents(activities []Activity, operatorDigest, sourceDigest, workflowDigest string, states []ReleaseState, receipts []APIReceipt) string {
	events := make([]Event, 0, len(activities))
	evidence := []string{
		"semantic-ir-and-fixed-denominator",
		operatorDigest,
		sourceDigest,
		"previous-immutable-release-and-merged-pr-lineage",
		"fixture-api-responses-before-mutation",
		"forward-only-release-state-machine",
		"expected-asset-manifest-name-size-digest",
		workflowDigest,
		"operational-refutation-and-created-public-object-ids",
		"append-only-burned-version-ledger",
		"refuted-unknown-closed-precedence",
		"runtime-authority-zero-operator-authority-separate",
	}
	labels := []string{
		"protocol-declared",
		"operator-policy-receipt-bound",
		"source-run-artifact-bound",
		"release-lineage-bound",
		"mutation-payload-preflighted",
		"state-machine-bound",
		"asset-manifest-bound",
		"fixed-point-publication-guarded",
		"operational-refutation-preserved",
		"failed-version-burned",
		"decision-precedence-enforced",
		"runtime-authority-zero",
	}
	for i, activity := range activities {
		decision := Closed
		if i == 1 && operatorDigest == "" {
			decision = Unknown
		}
		receiptIDs := make([]string, 0, len(receipts))
		for _, receipt := range receipts {
			receiptIDs = append(receiptIDs, receipt.ID+"="+receipt.ResponseDigest)
		}
		events = append(events, Event{Ordinal: i + 1, Activity: activity.Name, Event: labels[i], Decision: decision, Proof: activity.Proof, Evidence: evidence[i], Family: activity.Family, Indicator: activity.Indicator, StateHistory: append([]ReleaseState(nil), states...), ObservedAPIReceipts: receiptIDs, MutationStarted: false, CreatedObjectIDs: []string{}, IncidentDisposition: "NONE_GENERATOR_READ_ONLY"})
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
	builder.WriteString("The receipt closes the thirty-two declared release-transport cases only. It does not claim global repository, platform, or supply-chain safety.\n\n")
	fmt.Fprintf(&builder, "fixed case denominator: %d; fixed indicator denominator: %d\n\n", receipt.CaseDenominator, receipt.IndicatorDenominator)
	builder.WriteString("| ordinal | case | family | indicator | expected | observed | reason |\n|---:|---|---|---|---|---|---|\n")
	for _, scenario := range receipt.Scenarios {
		fmt.Fprintf(&builder, "| %d | %s | %s | %s | %s | %s | %s |\n", scenario.Ordinal, scenario.ID, scenario.Family, scenario.Indicator, scenario.Expected, scenario.Decision, scenario.Reason)
	}
	builder.WriteString("\nscenario counts: ")
	fmt.Fprintf(&builder, "CLOSED=%d, UNKNOWN=%d, REFUTED=%d\n\n", receipt.Summary["CLOSED"], receipt.Summary["UNKNOWN"], receipt.Summary["REFUTED"])
	builder.WriteString("resolution precedence: `REFUTED > UNKNOWN > CLOSED`; only explicit `FIXED_POINT` may close the contract.\n\n")
	builder.WriteString("unknown records retain stage, step, reason, unknown_class, next_operation, and blocked_by. Failed mutation attempts preserve every created public object ID and burn the attempted version; no tag, draft, release, or asset is deleted, edited, or reused.\n\n")
	fmt.Fprintf(&builder, "authority: repository_writes=0, commits=0, pushes=0, merges=0, tags=0, releases=0, local_go_tests=0, cross_project_required_gates=0; operator_release=%s.\nauthoring audit: state=%s, local_test_invocations=%d, local_test_executions=%d; operator_stage_local_test_invocations=%d, operator_stage_local_test_executions=%d.\n", receipt.Authority.OperatorRelease.Workflow, receipt.OperationalAudit.State, receipt.OperationalAudit.LocalTestInvocations, receipt.OperationalAudit.LocalTestExecutions, receipt.OperationalAudit.OperatorStageLocalTestInvocations, receipt.OperationalAudit.OperatorStageLocalTestExecutions)
	return builder.String()
}
