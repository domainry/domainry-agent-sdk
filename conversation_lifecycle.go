package agentsdk

import (
	"context"
	"time"
)

const ConversationLifecycleContractVersion = "domainry-agent-conversation-lifecycle-v1"

type ConversationLifecycleStage string

const (
	ConversationLifecycleInputReceived     ConversationLifecycleStage = "input.received"
	ConversationLifecycleInputAdmitting    ConversationLifecycleStage = "input.admitting"
	ConversationLifecycleContextAssembling ConversationLifecycleStage = "context.assembling"
	ConversationLifecycleContextCompacting ConversationLifecycleStage = "context.compacting"
	ConversationLifecycleContextAssembled  ConversationLifecycleStage = "context.assembled"
	ConversationLifecycleModelRequest      ConversationLifecycleStage = "model.request"
	ConversationLifecycleModelCompleted    ConversationLifecycleStage = "model.completed"
	ConversationLifecycleModelFailed       ConversationLifecycleStage = "model.failed"
	ConversationLifecycleModelRetry        ConversationLifecycleStage = "model.retry"
	ConversationLifecycleToolBefore        ConversationLifecycleStage = "tool.before"
	ConversationLifecycleToolCompleted     ConversationLifecycleStage = "tool.completed"
	ConversationLifecycleToolFailed        ConversationLifecycleStage = "tool.failed"
	ConversationLifecycleRunFinished       ConversationLifecycleStage = "run.finished"
	ConversationLifecycleTaskFinished      ConversationLifecycleStage = "task.finished"
)

const (
	ConversationLifecycleKindPolicy   = "policy"
	ConversationLifecycleKindObserver = "observer"

	ConversationLifecycleFailureFailClosed = "fail_closed"
	ConversationLifecycleFailureContinue   = "continue"
)

// ConversationLifecycleExtensionDefinition is read once during service
// assembly. Order is ascending, then Key. Version identifies implementation
// behavior; ConfigurationVersion is a non-secret identifier for deployment-owned configuration.
// Changing either requires a newly admitted run. Runtime hot swap is outside
// this contract.
type ConversationLifecycleExtensionDefinition struct {
	Key                  string                       `json:"key"`
	Version              string                       `json:"version"`
	ConfigurationVersion string                       `json:"configuration_version"`
	Order                int                          `json:"order"`
	Kind                 string                       `json:"kind"`         // policy or observer
	FailureMode          string                       `json:"failure_mode"` // fail_closed or continue
	Stages               []ConversationLifecycleStage `json:"stages"`
}

type ConversationLifecycleExtensionSnapshot struct {
	Key                  string                       `json:"key"`
	Version              string                       `json:"version"`
	ConfigurationVersion string                       `json:"configuration_version"`
	Order                int                          `json:"order"`
	Kind                 string                       `json:"kind"`
	FailureMode          string                       `json:"failure_mode"`
	Stages               []ConversationLifecycleStage `json:"stages"`
}

// ConversationLifecycleManifest is frozen on every foreground or background
// run. Digest binds the ordered extension definitions and configuration
// versions used at admission.
type ConversationLifecycleManifest struct {
	ContractVersion string                                   `json:"contract_version"`
	Digest          string                                   `json:"digest"`
	Extensions      []ConversationLifecycleExtensionSnapshot `json:"extensions"`
}

type ConversationLifecycleInput struct {
	ClientMessageID string                     `json:"client_message_id,omitempty"`
	Message         string                     `json:"message,omitempty"`
	Content         []ConversationContentBlock `json:"content,omitempty"`
}

type ConversationLifecycleFailure struct {
	ErrorCode              string `json:"error_code"`
	Retryable              bool   `json:"retryable"`
	RetryAfterMilliseconds int64  `json:"retry_after_ms,omitempty"`
}

type ConversationLifecycleOutcome struct {
	Status    string         `json:"status"`
	ErrorCode string         `json:"error_code,omitempty"`
	Model     string         `json:"model,omitempty"`
	Usage     map[string]any `json:"usage,omitempty"`
}

// ConversationLifecycleEvent is a trusted in-process extension event. Model
// and step requests are owned copies, so an extension cannot mutate the
// persisted request or bypass engine checks by retaining a reference. A stage
// may be redelivered with the same EventID after recovery; policy handlers must
// return the same decision for that ID and observers should deduplicate effects.
type ConversationLifecycleEvent struct {
	ContractVersion string                         `json:"contract_version"`
	EventID         string                         `json:"event_id"`
	Stage           ConversationLifecycleStage     `json:"stage"`
	Manifest        *ConversationLifecycleManifest `json:"manifest,omitempty"`
	Authority       ConversationAuthority          `json:"authority"`
	ConversationID  string                         `json:"conversation_id,omitempty"`
	RunID           string                         `json:"run_id,omitempty"`
	TaskID          string                         `json:"task_id,omitempty"`
	RunAttempt      int                            `json:"run_attempt,omitempty"`
	Step            int                            `json:"step,omitempty"`
	ModelAttempt    int                            `json:"model_attempt,omitempty"`
	Purpose         string                         `json:"purpose,omitempty"`
	ModelKey        string                         `json:"model_key,omitempty"`
	ToolCallID      string                         `json:"tool_call_id,omitempty"`
	ToolKey         string                         `json:"tool_key,omitempty"`
	ToolCall        *ConversationToolCall          `json:"tool_call,omitempty"`
	ToolDefinition  *ConversationToolDefinition    `json:"tool_definition,omitempty"`
	ToolResult      *ConversationToolResult        `json:"tool_result,omitempty"`
	Input           *ConversationLifecycleInput    `json:"input,omitempty"`
	ModelRequest    *ConversationModelRequest      `json:"model_request,omitempty"`
	StepRequest     *ConversationStepRequest       `json:"step_request,omitempty"`
	Failure         *ConversationLifecycleFailure  `json:"failure,omitempty"`
	Outcome         *ConversationLifecycleOutcome  `json:"outcome,omitempty"`
	OccurredAt      time.Time                      `json:"occurred_at"`
}

// ConversationLifecycleRetryDecision can only narrow the engine retry policy.
// Retry=false disables another attempt. MaxAttempts may lower the configured
// maximum. DelayMilliseconds may increase the engine-computed backoff, while
// the engine retains its configured maximum delay.
type ConversationLifecycleRetryDecision struct {
	Retry             bool  `json:"retry"`
	MaxAttempts       int   `json:"max_attempts,omitempty"`
	DelayMilliseconds int64 `json:"delay_ms,omitempty"`
}

// Decisions are stage-scoped. PlanningInstructions apply only while the
// initial context is assembled; ContextLimitBytes applies only before safe
// execution compaction; Retry applies only after a failed model request.
// Extensions never grant authorization, confirmation, tool availability,
// execution success or access to a saved result.
type ConversationLifecycleDecision struct {
	PlanningInstructions []string                            `json:"planning_instructions,omitempty"`
	ContextLimitBytes    int                                 `json:"context_limit_bytes,omitempty"`
	Retry                *ConversationLifecycleRetryDecision `json:"retry,omitempty"`
}

type ConversationLifecycleExtension interface {
	ConversationLifecycleDefinition() ConversationLifecycleExtensionDefinition
	// HandleConversationLifecycle must honor cancellation on ctx. The host calls
	// extensions synchronously in their declared order.
	HandleConversationLifecycle(context.Context, ConversationLifecycleEvent) (ConversationLifecycleDecision, error)
}
