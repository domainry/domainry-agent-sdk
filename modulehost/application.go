package modulehost

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
	agentmodel "github.com/domainry/domainry-agent-sdk/state"
)

// Principal is authenticated host evidence. Agent may persist the stable
// reference needed by a run, but it never receives Runtime authorization
// objects or imports Runtime implementation packages.
type Principal struct {
	Known                 bool
	WorkspaceID           string
	UserID                string
	RoleKey               string
	AuthorizationRevision string
	RequestID             string
	CorrelationID         string
	CausationID           string
}

func (p Principal) Reference() agentsdk.PrincipalReference {
	return agentsdk.PrincipalReference{
		WorkspaceID:           p.WorkspaceID,
		UserID:                p.UserID,
		RoleKey:               p.RoleKey,
		AuthorizationRevision: p.AuthorizationRevision,
	}
}

type InteractiveAuthorizationRequest struct {
	Context   agentsdk.GlobalContext
	Principal Principal
}

type InteractiveContextRequest struct {
	Principal             Principal
	EntrypointKey         string
	RouteKey              string
	ObjectKey             string
	RecordID              string
	SelectedRecordIDs     []string
	Locale                string
	Timezone              string
	AvailableOperationIDs []string
}

// InteractiveAuthorization contains only the current facts Agent needs to
// execute its conversation use case. Runtime remains the authority that
// resolves identity, schema visibility and business permissions.
type InteractiveAuthorization struct {
	Principal    Principal
	Context      agentsdk.GlobalContext
	Agent        agentsdk.AgentSchema
	Candidates   []agentsdk.RouteCandidate
	AllowedTools []string
}

type InteractiveTaskAuthorizationRequest struct {
	Principal   Principal
	TaskKey     string
	TaskVersion string
}

type InteractiveTaskAuthorization struct {
	Identity agentsdk.ExecutionIdentity
	Task     agentsdk.AgentTaskDefinition
	Evidence agentmodel.AgentAuthorizationEvidence
}

type InteractiveToolInvocationRequest struct {
	Context                                   agentsdk.GlobalContext
	Principal                                 Principal
	RunID, SessionID, EntrypointKey, RouteKey string
	CorrelationID, IdempotencyKey             string
	Route                                     agentsdk.RouteResult
}

// InteractiveToolInvocationResult is the host effect plus its authorization
// evidence. Agent records that evidence in its own run state.
type InteractiveToolInvocationResult struct {
	Tool          string
	Status        string
	Output        any
	Proposal      any
	Authorization agentmodel.AgentAuthorizationEvidence
}

type InteractiveWorkflowHandoffRequest struct {
	WorkflowKey    string
	Input          map[string]any
	InteractiveID  string
	IdempotencyKey string
	Principal      Principal
}

// InteractiveHost is the narrow Runtime capability boundary used by the
// Agent-owned interactive application. It deliberately exposes host facts and
// effects, not an Agent conversation use case.
type InteractiveHost interface {
	ResolveInteractiveContext(context.Context, InteractiveContextRequest) (agentsdk.GlobalContext, error)
	AuthorizeInteractive(context.Context, InteractiveAuthorizationRequest) (InteractiveAuthorization, error)
	AuthorizeInteractiveTask(context.Context, InteractiveTaskAuthorizationRequest) (InteractiveTaskAuthorization, error)
	InvokeInteractiveTool(context.Context, InteractiveToolInvocationRequest) (InteractiveToolInvocationResult, error)
	StartInteractiveWorkflow(context.Context, InteractiveWorkflowHandoffRequest) (string, error)
	WakeAgentTask(context.Context, string, string)
}

// TaskAuthorizationRequest asks the Runtime host to refresh the business
// authorization used by an Agent-owned asynchronous task. It is deliberately
// independent of Runtime workflow and application types.
type TaskAuthorizationRequest struct {
	TaskRunID, WorkspaceID, ProcessID, NodeInstanceID string
	TaskKey, TaskVersion                              string
	Identity                                          agentsdk.ExecutionIdentity
	CorrelationID                                     string
	PreviousEvidence                                  []agentmodel.AgentAuthorizationEvidence
}

type TaskAuthorization struct {
	Principal    Principal
	Identity     agentsdk.ExecutionIdentity
	Task         agentsdk.AgentTaskDefinition
	Evidence     agentmodel.AgentAuthorizationEvidence
	AllowedTools []string
}

type TaskCredentialRequest struct {
	WorkspaceID, ProcessID, TaskRunID string
	Principal                         Principal
	AllowedTools                      []string
	TTLSeconds                        int
}

type TaskToolRequest struct {
	Credential, WorkspaceID, ProcessID, TaskRunID string
	Owner                                         string
	FencingToken                                  int64
	Identity                                      agentsdk.ExecutionIdentity
	TaskKey, TaskVersion, Tool, IdempotencyKey    string
	Input                                         map[string]any
	PreviousEvidence                              []agentmodel.AgentAuthorizationEvidence
}

type TaskToolResult struct {
	Status, Tool     string
	Output, Proposal any
	Authorization    agentmodel.AgentAuthorizationEvidence
}

// WorkflowTaskCompletion is a one-way asynchronous result notification from
// Agent to Runtime. Runtime never calls back into Agent while committing it.
type WorkflowTaskCompletion struct {
	WorkspaceID, TaskRunID, ProcessID, NodeInstanceID string
	TaskKey, TaskVersion                              string
	Identity                                          agentsdk.ExecutionIdentity
	Status                                            string
	Outcome                                           string
	Output                                            map[string]any
	ErrorCode                                         string
	Evidence                                          agentmodel.AgentTaskExecutionEvidence
}

type TaskHost interface {
	AuthorizeTask(context.Context, TaskAuthorizationRequest) (TaskAuthorization, error)
	IssueTaskCredential(context.Context, TaskCredentialRequest) (string, error)
	InvokeTaskTool(context.Context, TaskToolRequest) (TaskToolResult, error)
	CompleteWorkflowTask(context.Context, WorkflowTaskCompletion) error
}

type GuardedWriteContract struct {
	ObjectKey, Operation, ActionKey, Endpoint string
	RequiresRecord                            bool
}

type ProposalActionRequest struct {
	ActionKey, ObjectKey, RecordID, IdempotencyKey string
	Input                                          map[string]any
	Principal                                      Principal
}

type ProposalActionResult struct{ Record, Object any }

type ProposalWorkflowRequest struct {
	WorkflowKey string
	Payload     map[string]any
	Principal   Principal
}

// ProposalHost exposes Runtime-owned business effects while proposal policy,
// decision CAS and Agent task approval lifecycle remain in Agent.
type ProposalHost interface {
	GuardedWrites(context.Context, Principal) []GuardedWriteContract
	ResolveProposalPrincipal(context.Context, string, string) (Principal, error)
	InvokeProposalAction(context.Context, ProposalActionRequest) (ProposalActionResult, error)
	RunProposalWorkflow(context.Context, ProposalWorkflowRequest) (any, error)
}

type AuditRequest struct {
	Event, ObjectKey, RecordID, Summary string
	Principal                           Principal
	Before, After, Metadata             map[string]any
}

type AuditHost interface {
	AppendAgentAudit(context.Context, AuditRequest) error
	ListAgentAudit(context.Context, Principal, int) ([]AuditEvent, error)
}

type AuditEvent struct {
	Event, ObjectKey, RecordID, Summary string
	Metadata                            map[string]any
	CreatedAt                           string
}

type AnalysisCatalog struct {
	ObjectKeys   []string
	ReportKeys   []string
	MaskedFields map[string][]string
}

type AnalysisRecord struct {
	ID   string
	Data map[string]any
}

type AnalysisRecordPage struct {
	Items   []AnalysisRecord
	Total   int
	HasNext bool
}

type AnalysisHost interface {
	ResolveAnalysisCatalog(context.Context, Principal) (AnalysisCatalog, error)
	ListAnalysisRecords(context.Context, string, map[string]any, int, Principal) (AnalysisRecordPage, error)
}

// ApplicationHost is bound only after Runtime business services exist.
// Persistence is opened first so Agent-owned migrations and definition
// restoration can run without creating a Runtime/Agent construction cycle.
type ApplicationHost interface {
	InteractiveAgent() InteractiveHost
	TaskAgent() TaskHost
	ProposalAgent() ProposalHost
	AuditAgent() AuditHost
	AnalysisAgent() AnalysisHost
}

type ApplicationHostBinder interface {
	BindApplicationHost(ApplicationHost) error
}
