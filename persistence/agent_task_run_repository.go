package persistence

import (
	"context"
	"errors"
	"time"

	agentmodel "github.com/domainry/domainry-agent-sdk/state"
)

var ErrSystemScopeRequired = errors.New("Agent system scope is required")

const AgentSystemScopeKindGlobal = "agent_global"

// SystemScope is Agent-owned worker authority for payload-free cross-workspace
// task discovery. It never carries a Runtime principal or grants host effects;
// those are reauthorized separately through TaskHost after a fenced claim.
type SystemScope struct {
	Kind    string
	Purpose string
}

type AgentTaskRunFilter struct {
	Statuses  []agentmodel.AgentTaskRunStatus
	ProcessID string
	TaskKey   string
	Limit     int
}

type AgentTaskRunSystemWorkerRepository interface {
	ListAgentTaskRunsForWorker(context.Context, SystemScope, AgentTaskRunFilter) ([]agentmodel.AgentTaskRun, error)
	ClaimNextAgentTaskRunForWorker(context.Context, SystemScope, string, time.Time, time.Duration) (AgentTaskClaim, bool, error)
}

// AgentTaskRunDirectClaimRepository lets a post-commit wakeup claim the exact
// durable task it refers to. Discovery remains separate and is only a recovery
// path for lost wakeups, retries and expired leases.
type AgentTaskRunDirectClaimRepository interface {
	ClaimAgentTaskRun(context.Context, string, string, string, time.Time, time.Duration) (AgentTaskClaim, bool, error)
}

type AgentTaskClaim struct {
	Run   agentmodel.AgentTaskRun
	Lease agentmodel.AgentTaskLease
}

type AgentTaskHeartbeatResult struct {
	Lost  bool
	Lease agentmodel.AgentTaskLease
}

type AgentToolCallStart struct {
	WorkspaceID, ProcessID, TaskRunID, Tool, InputHash string
	Owner                                              string
	FencingToken                                       int64
	MaxToolCalls                                       int
	CostUnits, MaxCostUnits                            int
	Authorization                                      agentmodel.AgentAuthorizationEvidence
}

type AgentToolCallFinish struct {
	WorkspaceID, TaskRunID, CallRef, Status, ErrorCode string
	Owner                                              string
	FencingToken                                       int64
	Evidence                                           map[string]any
	Authorization                                      agentmodel.AgentAuthorizationEvidence
}

type AgentToolCallLedger interface {
	BeginAgentToolCall(context.Context, AgentToolCallStart) (string, int, error)
	FinishAgentToolCall(context.Context, AgentToolCallFinish) error
}

type AgentTaskRunRepository interface {
	Create(context.Context, agentmodel.AgentTaskRun) (agentmodel.AgentTaskRun, bool, error)
	Get(context.Context, string, string) (agentmodel.AgentTaskRun, bool, error)
	List(context.Context, string, AgentTaskRunFilter) ([]agentmodel.AgentTaskRun, error)
	ClaimNext(context.Context, string, string, time.Time, time.Duration) (AgentTaskClaim, bool, error)
	Heartbeat(context.Context, string, string, string, int64, time.Time, time.Duration) (AgentTaskHeartbeatResult, error)
	SaveRunning(context.Context, agentmodel.AgentTaskRun, string, int64) error
	SaveWaitingApproval(context.Context, agentmodel.AgentTaskRun, int64) error
	SaveTerminalOverride(context.Context, agentmodel.AgentTaskRun, int64) error
	SaveOperationalTransition(context.Context, agentmodel.AgentTaskRun, agentmodel.AgentTaskRunStatus, int64) error
	RequestCancel(context.Context, string, string, string, time.Time) (agentmodel.AgentTaskRun, bool, error)
}

type AgentInteractiveRunFilter struct {
	Statuses []agentmodel.AgentInteractiveRunStatus
	Limit    int
}

type AgentInteractiveRunRepository interface {
	CreateInteractiveRun(context.Context, agentmodel.AgentInteractiveRun) (agentmodel.AgentInteractiveRun, bool, error)
	GetInteractiveRun(context.Context, string, string) (agentmodel.AgentInteractiveRun, bool, error)
	ListInteractiveRuns(context.Context, string, string, string, AgentInteractiveRunFilter) ([]agentmodel.AgentInteractiveRun, error)
	SaveInteractiveRun(context.Context, agentmodel.AgentInteractiveRun, int64) (bool, error)
	CommitInteractiveTaskHandoff(context.Context, agentmodel.AgentInteractiveRun, int64, agentmodel.AgentTaskRun) (agentmodel.AgentInteractiveRun, bool, error)
}
