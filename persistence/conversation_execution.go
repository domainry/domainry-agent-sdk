package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Execution records are internal persistence contracts. Product APIs expose
// explicit views so model input and opaque provider state stay server-side.
type ConversationExecutionStep struct {
	Number    int                              `json:"number"`
	Input     agentsdk.ConversationStepRequest `json:"input"`
	Result    *agentsdk.ConversationStepResult `json:"result,omitempty"`
	CreatedAt time.Time                        `json:"created_at"`
	UpdatedAt time.Time                        `json:"updated_at"`
}

type ConversationToolExecution struct {
	Step           int                                    `json:"step"`
	Call           agentsdk.ConversationToolCall          `json:"call"`
	Definition     agentsdk.ConversationToolDefinition    `json:"definition"`
	IdempotencyKey string                                 `json:"idempotency_key"`
	Authorization  agentsdk.ConversationToolAuthorization `json:"authorization"`
	State          string                                 `json:"state"` // started, uncertain or completed
	Result         *agentsdk.ConversationToolResult       `json:"result,omitempty"`
	CreatedAt      time.Time                              `json:"created_at"`
	UpdatedAt      time.Time                              `json:"updated_at"`
}

// Optional repository extension; every mutation validates the current run
// lease/fence and commits its event with the record in one transaction.
type ConversationExecutionRepository interface {
	ExecutionStep(context.Context, ConversationClaim, int, *agentsdk.ConversationStepRequest) (ConversationExecutionStep, bool, error)
	CompleteExecutionStep(context.Context, ConversationClaim, int, agentsdk.ConversationStepResult) error
	ExecutionTools(context.Context, ConversationClaim, int) ([]ConversationToolExecution, error)
	BeginExecutionTool(context.Context, ConversationClaim, int, string, agentsdk.ConversationToolAuthorization) (ConversationToolExecution, bool, error)
	FinishExecutionTool(context.Context, ConversationClaim, int, string, agentsdk.ConversationToolResult) error
}
