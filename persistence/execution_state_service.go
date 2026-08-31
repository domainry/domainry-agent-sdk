package persistence

import (
	"context"
	"time"

	agentmodel "github.com/domainry/domainry-agent-sdk/state"
)

// AgentTaskActor is authenticated host evidence used by Agent-owned operator
// transitions. The host authenticates it; Agent validates scope and records it.
type AgentTaskActor struct {
	Known       bool
	WorkspaceID string
	UserID      string
	RoleKey     string
}

type AgentTaskAuditRequest struct {
	Event, ObjectKey, RecordID, Summary string
	Actor                               AgentTaskActor
	Before, After, Metadata             map[string]any
}

// AgentTaskAuditAppender is a Runtime-owned side-effect port. Agent decides
// which state transition requires audit evidence and supplies the full event.
type AgentTaskAuditAppender interface {
	AppendAgentTaskAudit(context.Context, AgentTaskAuditRequest) error
}

type AgentTaskRunCompletion struct {
	Status         agentmodel.AgentTaskRunStatus
	Outcome        string
	Output         map[string]any
	ExternalRunID  string
	RawEvidenceRef string
	ErrorCode      string
	Evidence       agentmodel.AgentTaskExecutionEvidence
	Reconciliation agentmodel.AgentTaskReconciliation
}

type AgentTaskApprovalResolution struct {
	ProposalID string
	Decision   string
	Actor      string
	Reason     string
	Execution  map[string]any
}

// AgentTaskStateService is the Agent-owned task state machine. Repository
// mutations, claims, leases, fencing and transition rules must not be
// reimplemented by Runtime application services.
type AgentTaskStateService interface {
	Create(context.Context, agentmodel.AgentTaskRun) (agentmodel.AgentTaskRun, bool, error)
	Get(context.Context, string, string) (agentmodel.AgentTaskRun, bool, error)
	List(context.Context, string, AgentTaskRunFilter) ([]agentmodel.AgentTaskRun, error)
	ClaimNext(context.Context, string, string, time.Duration) (AgentTaskClaim, bool, error)
	ClaimTask(context.Context, string, string, string, time.Duration) (AgentTaskClaim, bool, error)
	ClaimNextForWorker(context.Context, SystemScope, string, time.Duration) (AgentTaskClaim, bool, error)
	ListForWorker(context.Context, SystemScope, AgentTaskRunFilter) ([]agentmodel.AgentTaskRun, error)
	Heartbeat(context.Context, string, string, string, int64, time.Duration) (AgentTaskHeartbeatResult, error)
	WaitForApproval(context.Context, agentmodel.AgentTaskRun, string, int64, string, agentmodel.AgentTaskExecutionEvidence) (agentmodel.AgentTaskRun, error)
	ResolveApproval(context.Context, string, string, AgentTaskApprovalResolution) (agentmodel.AgentTaskRun, bool, error)
	FailAttempt(context.Context, agentmodel.AgentTaskRun, string, int64, string, string, bool, time.Time) (agentmodel.AgentTaskRun, error)
	Complete(context.Context, agentmodel.AgentTaskRun, string, int64, AgentTaskRunCompletion) (agentmodel.AgentTaskRun, error)
	MarkWorkflowCompletionDelivered(context.Context, string, string) (agentmodel.AgentTaskRun, error)
	RequestCancel(context.Context, string, string, string) (agentmodel.AgentTaskRun, bool, error)
	ExpireApprovals(context.Context, string, time.Time) (int, error)
	OverrideOutput(context.Context, string, string, map[string]any, AgentTaskActor, string, AgentTaskAuditAppender) (agentmodel.AgentTaskRun, error)
	Operate(context.Context, string, string, string, string, string, AgentTaskActor, AgentTaskAuditAppender) (agentmodel.AgentTaskRun, bool, error)
}

type AgentInteractiveAuthority struct {
	Known                 bool
	WorkspaceID           string
	UserID                string
	RoleKey               string
	AuthorizationRevision string
}

type AgentInteractiveRoute struct {
	RouteType, TargetKey, TargetVersion, IdempotencyKey string
}

// AgentInteractiveStateService owns interactive-run identity, authorization
// evidence, transitions, handoff state and tool invocation evidence.
type AgentInteractiveStateService interface {
	Create(context.Context, agentmodel.AgentInteractiveRun, AgentInteractiveAuthority) (agentmodel.AgentInteractiveRun, bool, error)
	Get(context.Context, string, AgentInteractiveAuthority) (agentmodel.AgentInteractiveRun, bool, error)
	List(context.Context, AgentInteractiveAuthority, AgentInteractiveRunFilter) ([]agentmodel.AgentInteractiveRun, error)
	Complete(context.Context, agentmodel.AgentInteractiveRun, agentmodel.AgentInteractiveRunStatus, map[string]any, string) (agentmodel.AgentInteractiveRun, error)
	HandoffTask(context.Context, agentmodel.AgentInteractiveRun, AgentInteractiveRoute, agentmodel.AgentTaskRun) (agentmodel.AgentInteractiveRun, bool, error)
	HandoffWorkflow(context.Context, agentmodel.AgentInteractiveRun, AgentInteractiveRoute, string) (agentmodel.AgentInteractiveRun, bool, error)
	RecordToolInvocation(context.Context, agentmodel.AgentInteractiveRun, agentmodel.AgentTaskToolInvocationEvidence) (agentmodel.AgentInteractiveRun, error)
	ObservePermissionDenied(context.Context)
	OpenMetrics(context.Context) string
}

// ExecutionStateBinding exposes deployment-neutral Agent application
// capabilities to the Agent-owned HTTP surface and worker assembly. Hosts do
// not mutate these repositories or reimplement their state transitions.
type ExecutionStateBinding interface {
	AgentTaskState() AgentTaskStateService
	AgentInteractiveState() AgentInteractiveStateService
}
