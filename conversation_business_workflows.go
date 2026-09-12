package agentsdk

import (
	"context"
	"encoding/json"
)

type ConversationWorkflowStart struct {
	WorkflowKey string          `json:"workflow_key"`
	Version     string          `json:"workflow_version"`
	Data        json.RawMessage `json:"data"`
}

type ConversationWorkflowGet struct {
	WorkflowKey string `json:"workflow_key"`
	ProcessID   string `json:"process_id"`
}

// Trusted call metadata is never populated from model arguments.
type ConversationWorkflowStartRequest struct {
	Authority      ConversationAuthority
	Start          ConversationWorkflowStart
	ConversationID string
	RunID          string
	CorrelationID  string
	Step           int
	CallID         string
	IdempotencyKey string
	Confirmation   *ConversationConfirmation
	Arguments      string
}

type ConversationWorkflowReceipt struct {
	Status       string `json:"status"` // accepted, failed, uncertain
	ErrorCode    string `json:"error_code,omitempty"`
	InvocationID string `json:"invocation_id,omitempty"`
	WorkflowKey  string `json:"workflow_key"`
	ExecutionID  string `json:"execution_id,omitempty"`
	ProcessID    string `json:"process_id,omitempty"`
}

type ConversationWorkflowState struct {
	WorkflowKey      string                               `json:"workflow_key"`
	ProcessID        string                               `json:"process_id"`
	Name             string                               `json:"name"`
	Status           string                               `json:"status"` // host's actual workflow state
	Terminal         bool                                 `json:"terminal"`
	BusinessOutcome  string                               `json:"business_outcome,omitempty"`
	CurrentSteps     []string                             `json:"current_steps"`
	CurrentStepCount int                                  `json:"current_step_count"`
	StepsTruncated   bool                                 `json:"steps_truncated,omitempty"`
	Record           *ConversationBusinessRecordReference `json:"record,omitempty"`
	UpdatedAt        string                               `json:"updated_at"`
	CompletedAt      string                               `json:"completed_at,omitempty"`
}

// Optional business extension. Start returns acceptance, not workflow
// completion. Get uses the host's current instance permission and state.
// Reconciliation never reclaims an existing unresolved start; an absent start
// may proceed only through an atomic no-reclaim claim. Revalidation is read-only.
type ConversationBusinessWorkflowSource interface {
	AuthorizeWorkflowStart(context.Context, ConversationWorkflowStart, ConversationAuthority) (ConversationToolAuthorization, error)
	StartBusinessWorkflow(context.Context, ConversationWorkflowStartRequest) (ConversationWorkflowReceipt, error)
	ReconcileBusinessWorkflow(context.Context, ConversationWorkflowStartRequest) (ConversationWorkflowReceipt, error)
	GetBusinessWorkflow(context.Context, ConversationWorkflowGet, ConversationAuthority) (ConversationWorkflowState, error)
	RevalidateBusinessWorkflow(context.Context, ConversationBusinessEvidence, ConversationAuthority) error
}

func BusinessWorkflowConversationTools() []ConversationToolDefinition {
	return []ConversationToolDefinition{
		{Key: "workflow_start", Version: "1", ActionKey: ConversationToolActionPrefix + "workflow_start", Effect: "write", Idempotency: "reconcile", TimeoutMillis: 60000, MaxOutputBytes: 64 * 1024,
			Description: "Start a declared business workflow after the user confirms the exact workflow and input. First expand business_catalog kind=workflows and use its execution_version as workflow_version. Supply only published business fields and defaults; never supply identity, credentials, run-as roles or idempotency metadata. accepted means the host accepted this start, not that the workflow finished. Save the returned process_id and use workflow_get to read its actual progress and outcome. An uncertain start requires reconciliation, never a second start with a different call.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"workflow_key":{"type":"string","minLength":1,"maxLength":128},"workflow_version":{"type":"string","minLength":1,"maxLength":256},"data":{"type":"object","maxProperties":100}},"required":["workflow_key","workflow_version","data"],"additionalProperties":false}`), OutputSchema: json.RawMessage(`{"type":"object"}`)},
		{Key: "workflow_get", Version: "1", ActionKey: ConversationToolActionPrefix + "workflow_get", Effect: "read", Idempotency: "natural", TimeoutMillis: 30000, MaxOutputBytes: 64 * 1024,
			Description: "Read current progress and business outcome of a permitted workflow instance. Use an actual process_id returned by workflow_start or an authorized source; never invent IDs. Report accepted/running/waiting as still in progress. terminal also includes failure or cancellation, so check status and business_outcome before claiming success. This reads current state; it does not resume, approve, cancel or restart the workflow. Referenced business records must be read through authorized record tools.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"workflow_key":{"type":"string","minLength":1,"maxLength":128},"process_id":{"type":"string","minLength":1,"maxLength":256}},"required":["workflow_key","process_id"],"additionalProperties":false}`), OutputSchema: json.RawMessage(`{"type":"object"}`)},
	}
}
