package persistence

import (
	"strings"
	"time"

	agentmodel "github.com/domainry/domainry-agent-sdk/state"
)

// AgentTaskRunView is the stable, redacted task-run projection shared by the
// Agent-owned read Surface and Runtime-owned operator mutation responses.
type AgentTaskRunView struct {
	ID                     string                `json:"id"`
	ProcessID              string                `json:"process_id,omitempty"`
	NodeInstanceID         string                `json:"node_instance_id,omitempty"`
	InteractiveRunID       string                `json:"interactive_run_id,omitempty"`
	TaskKey                string                `json:"task_key"`
	TaskVersion            string                `json:"task_version"`
	Status                 string                `json:"status"`
	Outcome                string                `json:"outcome,omitempty"`
	ProposalID             string                `json:"proposal_id,omitempty"`
	ErrorCode              string                `json:"error_code,omitempty"`
	ReconciliationRequired bool                  `json:"reconciliation_required"`
	ReconciliationState    string                `json:"reconciliation_state,omitempty"`
	Identity               AgentTaskIdentityView `json:"identity"`
	Output                 map[string]any        `json:"output,omitempty"`
	ActionReceipts         []string              `json:"action_receipts,omitempty"`
	AuditRefs              []string              `json:"audit_refs,omitempty"`
	Attempt                int                   `json:"attempt"`
	MaxAttempts            int                   `json:"max_attempts"`
	Revision               int64                 `json:"revision"`
	CreatedAt              string                `json:"created_at"`
	UpdatedAt              string                `json:"updated_at"`
}

type AgentTaskIdentityView struct {
	Mode                string `json:"mode"`
	ExecutionUserID     string `json:"execution_user_id"`
	ExecutionRoleKey    string `json:"execution_role_key"`
	ServicePrincipalKey string `json:"service_principal_key,omitempty"`
}

// ProjectAgentTaskRun deliberately excludes task input, raw evidence, lease
// fencing details and tool invocation evidence from external HTTP responses.
func ProjectAgentTaskRun(run agentmodel.AgentTaskRun) AgentTaskRunView {
	proposalID := ""
	if run.Approval != nil {
		proposalID = run.Approval.ProposalID
	}
	return AgentTaskRunView{
		ID: run.ID, ProcessID: run.ProcessID, NodeInstanceID: run.NodeInstanceID, InteractiveRunID: run.InteractiveRunID,
		TaskKey: run.TaskKey, TaskVersion: run.TaskVersion, Status: string(run.Status), Outcome: run.Outcome, ProposalID: proposalID,
		ErrorCode: run.LastErrorCode, Output: cloneTaskRunMap(run.Output), ActionReceipts: taskRunActionReceipts(run), AuditRefs: append([]string(nil), run.Evidence.AuditRefs...),
		ReconciliationRequired: run.Reconciliation.Required, ReconciliationState: run.Reconciliation.State,
		Identity: AgentTaskIdentityView{Mode: run.Identity.Mode, ExecutionUserID: run.Identity.Execution.UserID, ExecutionRoleKey: run.Identity.Execution.RoleKey, ServicePrincipalKey: run.Identity.ServicePrincipalKey},
		Attempt:  run.Attempt, MaxAttempts: run.MaxAttempts, Revision: run.Revision, CreatedAt: run.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: run.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func cloneTaskRunMap(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}
	return out
}

func taskRunActionReceipts(run agentmodel.AgentTaskRun) []string {
	refs := []string{}
	if run.Approval == nil {
		return refs
	}
	for _, key := range []string{"receipt_id", "action_receipt_id", "idempotency_receipt_id"} {
		if value, ok := run.Approval.Execution[key].(string); ok {
			if value = strings.TrimSpace(value); value != "" {
				refs = append(refs, value)
			}
		}
	}
	return refs
}
