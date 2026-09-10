package agentsdk

import (
	"context"
	"encoding/json"
)

// Conversation execution is an optional capability. The existing text-only
// ConversationModel and Interactive/Task contracts retain their semantics.
const CapabilityConversationExecutionV1 = "conversation.execution.v1"

type ConversationExecutionStatusProvider interface{ ConversationExecutionEnabled() bool }

// Bounded public projections. These never contain model input snapshots,
// policy bundles, credentials, or opaque provider continuation blocks.
type ConversationToolView struct {
	Citations       []ConversationCitation       `json:"citations,omitempty"`
	ID              string                       `json:"id"`
	Name            string                       `json:"name"`
	Arguments       string                       `json:"arguments"`
	Status          string                       `json:"status"`
	ErrorCode       string                       `json:"error_code,omitempty"`
	ResourceID      string                       `json:"resource_id,omitempty"`
	ResultPreview   string                       `json:"result_preview,omitempty"`
	ResultTruncated bool                         `json:"result_truncated,omitempty"`
	ResultReference *ConversationResultReference `json:"result_reference,omitempty"`
}
type ConversationStepView struct {
	Number  int                    `json:"number"`
	Attempt int                    `json:"attempt"`
	Status  string                 `json:"status"`
	Text    string                 `json:"text"`
	Calls   []ConversationToolView `json:"calls"`
}

// Tool definitions come from trusted host registration. A model chooses a key
// and arguments; it cannot supply the authorization, effect or retry policy.
type ConversationToolDefinition struct {
	Key            string          `json:"key"`
	Version        string          `json:"version"`
	Description    string          `json:"description"`
	InputSchema    json.RawMessage `json:"input_schema"`
	OutputSchema   json.RawMessage `json:"output_schema"`
	ActionKey      string          `json:"action_key"`
	Effect         string          `json:"effect"`      // read or write
	Idempotency    string          `json:"idempotency"` // natural, key or reconcile
	TimeoutMillis  int             `json:"timeout_ms"`
	MaxOutputBytes int             `json:"max_output_bytes"`
}

type ConversationToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // Complete JSON text; the executor validates the schema.
}

// ProviderState carries protocol continuation blocks (e.g. signed thinking or
// encrypted reasoning). It is server-side execution state, never a browser
// input, tool result, memory, or user-visible reasoning transcript.
type ConversationStepMessage struct {
	Role          string                 `json:"role"`
	Content       string                 `json:"content"`
	ToolCalls     []ConversationToolCall `json:"tool_calls,omitempty"`
	ToolCallID    string                 `json:"tool_call_id,omitempty"`
	IsError       bool                   `json:"is_error,omitempty"`
	ProviderState json.RawMessage        `json:"provider_state,omitempty"`
	// Server-created linkage for compaction. Providers receive only Content;
	// model output must never supply or overwrite this field.
	ResultReference *ConversationResultReference `json:"result_reference,omitempty"`
}

type ConversationModelIdentity struct {
	Provider    string `json:"provider"`
	Protocol    string `json:"protocol"`
	Model       string `json:"model"`
	Fingerprint string `json:"fingerprint"` // Non-secret configuration identity for recovery.
}

type ConversationStepRequest struct {
	Messages         []ConversationStepMessage      `json:"messages"`
	Tools            []ConversationToolDefinition   `json:"tools"`
	ModelIdentity    ConversationModelIdentity      `json:"model_identity"`
	IdempotencyKey   string                         `json:"idempotency_key"`
	MaxOutputBytes   int                            `json:"max_output_bytes"`
	MaxArgumentBytes int                            `json:"max_argument_bytes"`
	MaxToolCalls     int                            `json:"max_tool_calls"`
	Compaction       *ConversationContextCompaction `json:"compaction,omitempty"`
}

type ConversationStepResult struct {
	Message      ConversationStepMessage `json:"message"`
	FinishReason string                  `json:"finish_reason"` // stop or tool_calls
	Model        string                  `json:"model"`
	Usage        map[string]any          `json:"usage,omitempty"`
}

// Events are previews of one model step, not authorization to execute a tool.
// Only a normally completed StepResult can be submitted to the tool executor.
// Offset counts UTF-8 bytes within Text or this tool's argument stream.
type ConversationModelEvent struct {
	Type   string `json:"type"` // text.delta, tool.started, tool.arguments.delta
	Index  int    `json:"index,omitempty"`
	CallID string `json:"call_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Offset int    `json:"offset"`
	Delta  string `json:"delta,omitempty"`
}

// Callbacks are sequential and synchronous; a callback failure cancels the
// request. Incomplete/truncated streams must return an error and no result.
type ConversationAgentModel interface {
	ConversationModelIdentity() ConversationModelIdentity
	StreamConversationStep(context.Context, ConversationStepRequest, func(ConversationModelEvent) error) (ConversationStepResult, error)
}

// Execution authority is resolved by the host on each call, including resume.
// Granted means authorized now, and ConfirmationRequired is a separate policy
// decision; neither can be inferred from a tool's visibility in the catalog.
type ConversationToolAuthorization struct {
	UserTimezone         string         `json:"user_timezone,omitempty"`
	Granted              bool           `json:"granted"`
	ConfirmationRequired bool           `json:"confirmation_required"`
	Revision             string         `json:"revision"`
	Evidence             map[string]any `json:"evidence,omitempty"`
}

type ConversationToolRequest struct {
	Authority      ConversationAuthority
	ConversationID string
	RunID          string
	Step           int
	Call           ConversationToolCall
	Definition     ConversationToolDefinition
	IdempotencyKey string
	ConfirmationID string
	Confirmation   *ConversationConfirmation
	// Server-only worker guard. Local transactional effects validate this
	// against persisted state; it is never accepted from model arguments.
	LeaseOwner string
	Fence      int64
}

type ConversationToolResult struct {
	Status     string          `json:"status"` // completed, failed, pending or uncertain
	Content    json.RawMessage `json:"content,omitempty"`
	ErrorCode  string          `json:"error_code,omitempty"`
	ResourceID string          `json:"resource_id,omitempty"`
}

// Narrow execution port for Conversation. It does not require a Workflow or
// an existing TaskDefinition; concrete business effects remain host-owned.
type ConversationToolHost interface {
	ConversationTools(context.Context, ConversationAuthority) ([]ConversationToolDefinition, error)
	AuthorizeConversationTool(context.Context, ConversationToolRequest) (ConversationToolAuthorization, error)
	InvokeConversationTool(context.Context, ConversationToolRequest) (ConversationToolResult, error)
	ReconcileConversationTool(context.Context, ConversationToolRequest) (ConversationToolResult, error)
}
