package transport

import "encoding/json"

const (
	ContractSchema = "gooo/release-transport-conformer/transport-contract/v4"
	ManifestSchema = "gooo/release-transport-conformer/transport-manifest/v4"
	EventsSchema   = "gooo/release-transport-conformer/transport-events/v4"
	ReceiptSchema  = "gooo/release-transport-conformer/conformance-receipt/v4"
	Toolchain      = "go1.27.0"
)

type Decision string

const (
	Closed  Decision = "CLOSED"
	Unknown Decision = "UNKNOWN"
	Refuted Decision = "REFUTED"
)

var Precedence = []Decision{Refuted, Unknown, Closed}

var RequiredActivities = []string{
	"DeclareReleaseTransportProtocol",
	"BindOperatorPolicyReceipt",
	"BindSourceRunArtifact",
	"GenerateDraftFirstWorkflow",
	"ReconcileExistingDraft",
	"ResolveSymbolicReleaseTarget",
	"UseReleaseUploadURL",
	"VerifyAnnotatedTagTarget",
	"VerifyPublicAssetDigests",
	"PreserveTransportCounterexample",
	"EmitTransportReceipt",
}

type Activity struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Proof     string `json:"proof"`
	Artifact  string `json:"artifact"`
	Authority string `json:"authority"`
}

type ScenarioSpec struct {
	Ordinal      int      `json:"ordinal"`
	ID           string   `json:"id"`
	Expected     Decision `json:"expected_decision"`
	UnknownClass string   `json:"unknown_class,omitempty"`
}

type Contract struct {
	Schema        string         `json:"schema"`
	ContractID    string         `json:"contract_id"`
	Authority     string         `json:"authority"`
	Precedence    []Decision     `json:"precedence"`
	UnknownFields []string       `json:"unknown_fields"`
	Denominator   int            `json:"denominator"`
	Activities    []Activity     `json:"activities"`
	Scenarios     []ScenarioSpec `json:"scenarios"`
}

type OperatorPolicyReceipt struct {
	Schema     string `json:"schema"`
	Repository string `json:"repository"`
	Enabled    bool   `json:"enabled"`
	Immutable  bool   `json:"immutable"`
	Digest     string `json:"digest"`
	Endpoint   string `json:"endpoint,omitempty"`
}

type AssetObservation struct {
	Path         string `json:"path"`
	Digest       string `json:"digest"`
	ChecksumPath string `json:"checksum_path"`
}

type TagObservation struct {
	Exists         bool   `json:"exists"`
	Annotated      bool   `json:"annotated"`
	ObjectType     string `json:"object_type"`
	ObjectSHA      string `json:"object_sha"`
	TargetCommit   string `json:"target_commit"`
	ExpectedCommit string `json:"expected_commit"`
	Collision      bool   `json:"collision"`
}

type SourceRunObservation struct {
	SourceDigest        string `json:"source_digest"`
	CurrentSourceDigest string `json:"current_source_digest"`
	Stale               bool   `json:"stale"`
}

type GitIdentityObservation struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TransportObservation struct {
	Workflow                 string                 `json:"workflow"`
	OperatorPolicy           *OperatorPolicyReceipt `json:"operator_policy,omitempty"`
	SourceRun                SourceRunObservation   `json:"source_run"`
	GitIdentity              GitIdentityObservation `json:"git_identity"`
	Assets                   []AssetObservation     `json:"assets"`
	Tag                      TagObservation         `json:"tag"`
	DraftID                  string                 `json:"draft_id"`
	DraftTag                 string                 `json:"draft_tag"`
	DraftSourceTarget        string                 `json:"draft_source_target"`
	DraftTargetCommitish     string                 `json:"draft_target_commitish"`
	PeeledTagTarget          string                 `json:"peeled_tag_target"`
	SymbolicTargetTreatedAsExact bool              `json:"symbolic_target_treated_as_exact"`
	ExistingDraftAssets      string                 `json:"existing_draft_assets"`
	ReconciledDraft          bool                   `json:"reconciled_draft"`
	ExistingDraftMismatch    bool                   `json:"existing_draft_mismatch"`
	DraftUploadURL           string                 `json:"draft_upload_url"`
	UploadURLTemplateRemoved bool                   `json:"upload_url_template_removed"`
	UploadEndpoint           string                 `json:"upload_endpoint"`
	BinaryUploadViaAPI       bool                   `json:"binary_upload_via_api"`
	DraftCreated             bool                   `json:"draft_created"`
	AssetsUploaded           bool                   `json:"assets_uploaded"`
	Published                bool                   `json:"published"`
	PublishedImmutable       bool                   `json:"published_immutable"`
	DeterministicReplay      bool                   `json:"deterministic_replay"`
	ChecksumPathMismatch     bool                   `json:"checksum_path_mismatch"`
	ActionHasUserTokenSecret bool                   `json:"action_has_user_token_secret"`
	ActionHasAdminEndpoint   bool                   `json:"action_has_admin_endpoint"`
}

type UnknownEvidence struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (u UnknownEvidence) Valid() bool {
	return u.Stage != "" && u.Step != "" && u.Reason != "" && u.UnknownClass != "" && u.NextOperation != "" && len(u.BlockedBy) > 0
}

type Refutation struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

type ScenarioResult struct {
	Ordinal    int              `json:"ordinal"`
	ID         string           `json:"id"`
	Expected   Decision         `json:"expected_decision"`
	Decision   Decision         `json:"decision"`
	Reason     string           `json:"reason"`
	Unknown    *UnknownEvidence `json:"unknown,omitempty"`
	Refutation *Refutation      `json:"refutation,omitempty"`
}

type Inventory struct {
	DescendantDirs      int      `json:"descendant_dirs"`
	RegularFiles        int      `json:"regular_files"`
	GoFiles             int      `json:"go_files"`
	GoPhysicalLines     int      `json:"go_physical_lines"`
	GoooFiles           int      `json:"gooo_files"`
	GoooPhysicalLines   int      `json:"gooo_physical_lines"`
	RootREADMEExcluded  bool     `json:"root_readme_excluded"`
	ExcludedDirectories []string `json:"excluded_directories"`
}

type StageMeasurement struct {
	WallMS     int64 `json:"wall_ms"`
	PeakRSSKiB int64 `json:"peak_rss_kib"`
}

type StageMeasurements struct {
	Compile     StageMeasurement `json:"compile"`
	Build       StageMeasurement `json:"build"`
	Test        StageMeasurement `json:"test"`
	Conformance StageMeasurement `json:"conformance"`
	Integration StageMeasurement `json:"integration"`
}

type TestCounts struct {
	Total    int `json:"total"`
	Selected int `json:"selected"`
	Executed int `json:"executed"`
	Reused   int `json:"reused"`
	Failed   int `json:"failed"`
	Unknown  int `json:"unknown"`
}

type Authority struct {
	RepositoryWrites         int  `json:"repository_writes"`
	Commits                  int  `json:"commits"`
	Pushes                   int  `json:"pushes"`
	Merges                   int  `json:"merges"`
	Tags                     int  `json:"tags"`
	Releases                 int  `json:"releases"`
	LocalGoTests             int  `json:"local_go_tests"`
	CallerOwnedOutput        bool `json:"caller_owned_output"`
	SourceRepositoryReadOnly bool `json:"source_repository_read_only"`
}

type OperationalAudit struct {
	State                             Decision `json:"state"`
	Reason                            string   `json:"reason"`
	AuthoringLocalTestInvocations     int      `json:"authoring_local_test_invocations"`
	AuthoringLocalTestExecutions      int      `json:"authoring_local_test_executions"`
	AuthoringStaticValidationInvocations int   `json:"authoring_static_validation_invocations"`
	AuthoringStaticValidationExecutions  int   `json:"authoring_static_validation_executions"`
	OperatorStageLocalTestInvocations int      `json:"operator_stage_local_test_invocations"`
	OperatorStageLocalTestExecutions  int      `json:"operator_stage_local_test_executions"`
	LocalTestInvocations              int      `json:"local_test_invocations"`
	LocalTestExecutions               int      `json:"local_test_executions"`
	CIAuthority                       string   `json:"ci_authority"`
}

type Manifest struct {
	Schema                      string    `json:"schema"`
	Protocol                    string    `json:"protocol"`
	ContractID                  string    `json:"contract_id"`
	ContractDigest              string    `json:"contract_digest"`
	SourcePath                  string    `json:"source_path"`
	SourceDigest                string    `json:"source_digest"`
	WorkflowDigest              string    `json:"workflow_digest"`
	OperatorPolicyReceiptDigest string    `json:"operator_policy_receipt_digest,omitempty"`
	OutputFiles                 []string  `json:"output_files"`
	Activities                  []string  `json:"activities"`
	Authority                   Authority `json:"authority"`
}

type Event struct {
	Ordinal  int      `json:"ordinal"`
	Activity string   `json:"activity"`
	Event    string   `json:"event"`
	Decision Decision `json:"decision"`
	Proof    string   `json:"proof"`
	Evidence string   `json:"evidence"`
}

type Receipt struct {
	Schema                string            `json:"schema"`
	Protocol              string            `json:"protocol"`
	ContractID            string            `json:"contract_id"`
	ContractDigest        string            `json:"contract_digest"`
	SourceDigest          string            `json:"source_digest"`
	Decision              Decision          `json:"decision"`
	ConformanceClosed     bool              `json:"conformance_closed"`
	Precedence            []Decision        `json:"precedence"`
	Denominator           int               `json:"denominator"`
	Summary               map[string]int    `json:"summary"`
	Scenarios             []ScenarioResult  `json:"scenarios"`
	Activities            []Activity        `json:"activities"`
	ActivityBindingCounts map[string]int    `json:"activity_binding_counts"`
	Authority             Authority         `json:"authority"`
	OutputFiles           []string          `json:"output_files"`
	ArtifactDigests       map[string]string `json:"artifact_digests"`
	Inventory             Inventory         `json:"inventory"`
	Tests                 TestCounts        `json:"tests"`
	Measurements          StageMeasurements `json:"measurements"`
	OperationalAudit      OperationalAudit  `json:"operational_audit"`
	ScopeNote             string            `json:"scope_note"`
}

func JSON(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
