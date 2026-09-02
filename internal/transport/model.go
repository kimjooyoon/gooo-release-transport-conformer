package transport

import "encoding/json"

const (
	ContractSchema = "gooo/release-transport-conformer/transport-contract/v6"
	ManifestSchema = "gooo/release-transport-conformer/transport-manifest/v6"
	EventsSchema   = "gooo/release-transport-conformer/transport-events/v6"
	ReceiptSchema  = "gooo/release-transport-conformer/conformance-receipt/v6"
	SemanticIRSchema = "gooo/release-transport-conformer/semantic-ir/v6"
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
	"BindReleaseLineage",
	"PreflightMutationPayload",
	"AdvanceReleaseState",
	"BindExpectedAssetManifest",
	"PublishDraftAtFixedPoint",
	"PreserveOperationalRefutation",
	"BurnFailedVersion",
	"EnforceRefutedUnknownPrecedence",
	"EnforceZeroRuntimeAuthority",
}

var RequiredStates = []ReleaseState{
	PrecheckState,
	TaggedState,
	DraftCreatedState,
	AssetsUploadedState,
	AssetsAuditedState,
	PublishedImmutableState,
}

var RequiredTransitions = []Transition{
	{From: PrecheckState, To: TaggedState},
	{From: TaggedState, To: DraftCreatedState},
	{From: DraftCreatedState, To: AssetsUploadedState},
	{From: AssetsUploadedState, To: AssetsAuditedState},
	{From: AssetsAuditedState, To: PublishedImmutableState},
}

var RequiredIndicatorFamilies = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
var RequiredIndicatorRoles = []string{"DRIVER", "OUTCOME", "GUARDRAIL"}

type ReleaseState string

const (
	PrecheckState            ReleaseState = "PRECHECK"
	TaggedState              ReleaseState = "TAGGED"
	DraftCreatedState        ReleaseState = "DRAFT_CREATED"
	AssetsUploadedState      ReleaseState = "ASSETS_UPLOADED"
	AssetsAuditedState       ReleaseState = "ASSETS_AUDITED"
	PublishedImmutableState  ReleaseState = "PUBLISHED_IMMUTABLE"
)

type Transition struct {
	From ReleaseState `json:"from"`
	To   ReleaseState `json:"to"`
}

type VersionDeclaration struct {
	Version     string `json:"version"`
	Tag         string `json:"tag"`
	Stream      string `json:"stream"`
	PatchOnly   bool   `json:"patch_only"`
	NextVersion string `json:"next_version"`
}

type ImmutableReleaseIdentity struct {
	Repository       string `json:"repository"`
	Version          string `json:"version"`
	Tag              string `json:"tag"`
	ReleaseID        string `json:"release_id"`
	TagObjectSHA     string `json:"tag_object_sha"`
	TargetCommitSHA  string `json:"target_commit_sha"`
	Immutable        bool   `json:"immutable"`
	Assets           []ExpectedAsset `json:"assets"`
}

type MergedPRLineage struct {
	Repository       string `json:"repository"`
	Number           string `json:"number"`
	BaseBranch       string `json:"base_branch"`
	HeadSHA          string `json:"head_sha"`
	MergeCommitSHA   string `json:"merge_commit_sha"`
	Merged           bool   `json:"merged"`
}

type ExactTargetCommit struct {
	Value  string `json:"value"`
	Source string `json:"source"`
	Exact  bool   `json:"exact"`
}

type AnnotatedTagDeclaration struct {
	Tag            string `json:"tag"`
	ObjectSHA      string `json:"object_sha"`
	TargetCommitSHA string `json:"target_commit_sha"`
	Annotated      bool   `json:"annotated"`
}

type DraftReleaseDeclaration struct {
	ID             string `json:"id"`
	Tag            string `json:"tag"`
	TargetCommitish string `json:"target_commitish"`
	Draft          bool   `json:"draft"`
}

type ExpectedAsset struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Size   string `json:"size"`
	Digest string `json:"digest"`
}

type APIReceipt struct {
	ID             string `json:"id"`
	Method         string `json:"method"`
	Endpoint       string `json:"endpoint"`
	Status         int    `json:"status"`
	ResponseDigest string `json:"response_digest"`
	ResourceID     string `json:"resource_id"`
	Fixture        string `json:"fixture"`
}

type BurnedVersion struct {
	Version      string `json:"version"`
	AttemptID    string `json:"attempt_id"`
	Reason       string `json:"reason"`
	ReceiptDigest string `json:"receipt_digest"`
}

type IndicatorSpec struct {
	Ordinal  int    `json:"ordinal"`
	ID       string `json:"id"`
	Family   string `json:"family"`
	Role     string `json:"role"`
	Activity string `json:"activity"`
}

type Activity struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Proof     string `json:"proof"`
	Family    string `json:"family"`
	Indicator string `json:"indicator"`
	Artifact  string `json:"artifact"`
	Authority string `json:"authority"`
}

type ScenarioSpec struct {
	Ordinal      int      `json:"ordinal"`
	ID           string   `json:"id"`
	Expected     Decision `json:"expected_decision"`
	UnknownClass string   `json:"unknown_class,omitempty"`
	Family       string   `json:"family"`
	Indicator    string   `json:"indicator"`
	Resolution   string   `json:"resolution"`
}

type Contract struct {
	Schema        string         `json:"schema"`
	ContractID    string         `json:"contract_id"`
	Authority     string         `json:"authority"`
	Precedence    []Decision     `json:"precedence"`
	UnknownFields []string       `json:"unknown_fields"`
	Denominator   int            `json:"denominator"`
	PreviousDenominator int      `json:"previous_denominator"`
	AppendOnly    bool           `json:"append_only"`
	Activities    []Activity     `json:"activities"`
	Scenarios     []ScenarioSpec `json:"scenarios"`
	Version       VersionDeclaration       `json:"version_declaration"`
	PreviousRelease ImmutableReleaseIdentity `json:"previous_release"`
	MergedPR      MergedPRLineage          `json:"merged_pr_lineage"`
	ExactTarget   ExactTargetCommit        `json:"exact_target_commit"`
	AnnotatedTag  AnnotatedTagDeclaration  `json:"annotated_tag"`
	DraftRelease  DraftReleaseDeclaration  `json:"draft_release"`
	ExpectedAssets []ExpectedAsset          `json:"expected_asset_manifest"`
	APIReceipts   []APIReceipt              `json:"observed_api_receipts"`
	BurnedVersions []BurnedVersion          `json:"burned_versions"`
	States        []ReleaseState            `json:"states"`
	Transitions   []Transition              `json:"transitions"`
	Terminal      string                    `json:"terminal"`
	Indicators    []IndicatorSpec            `json:"indicators"`
}

type SemanticIR struct {
	Schema          string                    `json:"schema"`
	ContractID      string                    `json:"contract_id"`
	ContractDigest  string                    `json:"contract_digest"`
	SourceDigest    string                    `json:"source_digest"`
	PreviousDenominator int                   `json:"previous_denominator"`
	AppendOnly      bool                      `json:"append_only"`
	Version         VersionDeclaration        `json:"version"`
	PreviousRelease ImmutableReleaseIdentity  `json:"previous_release"`
	MergedPR        MergedPRLineage           `json:"merged_pr_lineage"`
	ExactTarget     ExactTargetCommit         `json:"exact_target_commit"`
	AnnotatedTag    AnnotatedTagDeclaration   `json:"annotated_tag"`
	DraftRelease    DraftReleaseDeclaration   `json:"draft_release"`
	ExpectedAssets  []ExpectedAsset            `json:"expected_asset_manifest"`
	APIReceipts     []APIReceipt               `json:"observed_api_receipts"`
	BurnedVersions  []BurnedVersion            `json:"burned_versions"`
	States          []ReleaseState              `json:"states"`
	Transitions     []Transition                `json:"transitions"`
	Terminal        string                      `json:"terminal"`
	Activities      []Activity                  `json:"activities"`
	Indicators      []IndicatorSpec             `json:"indicators"`
	Scenarios       []ScenarioSpec              `json:"scenarios"`
	Authority       Authority                    `json:"authority"`
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
	CreatedDraftID           string                 `json:"created_draft_id"`
	UsedCreatedDraftID       bool                   `json:"used_created_draft_id"`
	ImmediateDraftListRequired bool                 `json:"immediate_draft_list_required"`
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
	StateHistory             []ReleaseState         `json:"state_history"`
	StateMachineForwardOnly  bool                   `json:"state_machine_forward_only"`
	PreMutationFixtureConformance bool              `json:"pre_mutation_fixture_conformance"`
	MutationStarted          bool                   `json:"mutation_started"`
	FailurePreserved         bool                   `json:"failure_preserved"`
	BurnedVersion            bool                   `json:"burned_version"`
	CreatedPublicObjectIDsExact bool                `json:"created_public_object_ids_exact"`
	DirectMainReleaseTarget  bool                   `json:"direct_main_release_target"`
	CompareResultAmbiguous   bool                   `json:"compare_result_ambiguous"`
	WrongTargetCommitish     bool                   `json:"wrong_target_commitish"`
	TagReleaseOrderingError  bool                   `json:"tag_release_ordering_error"`
	ImmutableSettingEvidenceMissing bool            `json:"immutable_setting_evidence_missing"`
	MalformedImmutableSettingEvidence bool           `json:"malformed_immutable_setting_evidence"`
	AssetManifestMismatch    bool                   `json:"asset_manifest_mismatch"`
	DuplicateOrBurnedVersionReuse bool               `json:"duplicate_or_burned_version_reuse"`
	MutablePublishedRelease  bool                   `json:"mutable_published_release"`
	FixedPoint               bool                   `json:"fixed_point"`
	OperatorAuthoritySeparate bool                  `json:"operator_authority_separate"`
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
	Family     string           `json:"family"`
	Indicator  string           `json:"indicator"`
	Resolution string           `json:"resolution"`
	Numerator  int              `json:"numerator"`
	Denominator int             `json:"denominator"`
	Unknown    *UnknownEvidence `json:"unknown,omitempty"`
	Refutation *Refutation      `json:"refutation,omitempty"`
}

type CaseVector struct {
	Ordinal     int      `json:"ordinal"`
	ID          string   `json:"id"`
	Family      string   `json:"family"`
	Indicator   string   `json:"indicator"`
	Decision    Decision `json:"decision"`
	Numerator   int      `json:"numerator"`
	Denominator int      `json:"denominator"`
}

type IndicatorVector struct {
	Ordinal     int      `json:"ordinal"`
	ID          string   `json:"id"`
	Family      string   `json:"family"`
	Role        string   `json:"role"`
	Activity    string   `json:"activity"`
	Decision    Decision `json:"decision"`
	Numerator   int      `json:"numerator"`
	Denominator int      `json:"denominator"`
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
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
	GithubToken              string `json:"github_token"`
	OperatorRelease          OperatorReleaseAuthority `json:"operator_release"`
}

type OperatorReleaseAuthority struct {
	Workflow       string   `json:"workflow"`
	AllowedMutations []string `json:"allowed_mutations"`
	SeparatedFromRuntime bool `json:"separated_from_runtime"`
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
	MutationStarted                   bool     `json:"mutation_started"`
	CreatedPublicObjectIDs            []string `json:"created_public_object_ids"`
	IncidentDisposition               string   `json:"incident_disposition"`
	BurnedVersion                     bool     `json:"burned_version"`
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
	SemanticIRSchema             string    `json:"semantic_ir_schema"`
	ReleaseVersion               string    `json:"release_version"`
	PreviousReleaseID            string    `json:"previous_release_id"`
	PreviousDenominator          int       `json:"previous_denominator"`
	AppendOnly                   bool      `json:"append_only"`
	StateMachine                 []ReleaseState `json:"state_machine"`
	ExpectedAssetManifest        []ExpectedAsset `json:"expected_asset_manifest"`
}

type Event struct {
	Ordinal             int            `json:"ordinal"`
	Activity            string         `json:"activity"`
	Event               string         `json:"event"`
	Decision            Decision       `json:"decision"`
	Proof               string         `json:"proof"`
	Evidence            string         `json:"evidence"`
	Family              string         `json:"family"`
	Indicator           string         `json:"indicator"`
	StateHistory        []ReleaseState `json:"state_history"`
	ObservedAPIReceipts []string       `json:"observed_api_receipts"`
	MutationStarted     bool           `json:"mutation_started"`
	CreatedObjectIDs    []string       `json:"created_object_ids"`
	IncidentDisposition string         `json:"incident_disposition"`
}

type Receipt struct {
	Schema                string            `json:"schema"`
	Protocol              string            `json:"protocol"`
	ContractID            string            `json:"contract_id"`
	ContractDigest        string            `json:"contract_digest"`
	SourceDigest          string            `json:"source_digest"`
	Decision              Decision          `json:"decision"`
	Terminal              string            `json:"terminal"`
	State                 ReleaseState      `json:"state"`
	ConformanceClosed     bool              `json:"conformance_closed"`
	Precedence            []Decision        `json:"precedence"`
	Denominator           int               `json:"denominator"`
	Summary               map[string]int    `json:"summary"`
	Scenarios             []ScenarioResult  `json:"scenarios"`
	CaseDenominator       int               `json:"case_denominator"`
	CaseVectors           []CaseVector      `json:"case_vectors"`
	IndicatorDenominator  int               `json:"indicator_denominator"`
	IndicatorVectors      []IndicatorVector `json:"indicator_vectors"`
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
	SemanticIR             SemanticIR       `json:"semantic_ir"`
	Attempt                MutationAttempt  `json:"attempt"`
}

type MutationAttempt struct {
	Schema                    string                 `json:"schema"`
	AttemptID                 string                 `json:"attempt_id"`
	Version                   string                 `json:"version"`
	Tag                       string                 `json:"tag"`
	State                     ReleaseState           `json:"state"`
	Decision                  Decision               `json:"decision"`
	MutationStarted           bool                   `json:"mutation_started"`
	CreatedPublicObjectIDs    CreatedPublicObjectIDs `json:"created_public_object_ids"`
	PreserveNeverDelete       bool                   `json:"preserve_never_delete"`
	BurnedVersion             bool                   `json:"burned_version"`
	NextVersion               string                 `json:"next_version"`
	IncidentDisposition       string                 `json:"incident_disposition"`
}

type CreatedPublicObjectIDs struct {
	TagRefID       string   `json:"tag_ref_id,omitempty"`
	TagObjectID    string   `json:"tag_object_id,omitempty"`
	DraftReleaseID string   `json:"draft_release_id,omitempty"`
	ReleaseID      string   `json:"release_id,omitempty"`
	AssetIDs       []string `json:"asset_ids"`
}

func JSON(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}
