package agentsdk

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	ConversationCodeToolKey         = "run_code"
	ConversationCodeProtocolVersion = 1
	ConversationCodeLanguageLua     = "lua"
)

// ConversationCodeExecution is the complete, server-created request sent to
// an isolated code Runtime. Tools contains only the frozen definitions already
// selected for this model step; run_code itself is never a callable binding.
type ConversationCodeExecution struct {
	ProtocolVersion int                          `json:"protocol_version"`
	Language        string                       `json:"language"`
	Source          string                       `json:"source"`
	Tools           []ConversationToolDefinition `json:"tools"`
	MaxDispatches   int                          `json:"max_dispatches"`
	MaxOutputBytes  int                          `json:"max_output_bytes"`
	MaxLogBytes     int                          `json:"max_log_bytes"`
}

// ConversationCodeDispatch is emitted synchronously by the isolated Runtime.
// Index starts at zero and must increase by one for deterministic receipt
// linkage. Arguments is complete JSON; Agent validates it against the frozen
// tool definition before any invocation.
type ConversationCodeDispatch struct {
	Index     int             `json:"index"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ConversationCodeDispatcher func(context.Context, ConversationCodeDispatch) (ConversationToolResult, error)

type ConversationCodeResult struct {
	Language   string          `json:"language"`
	Value      json.RawMessage `json:"value"`
	Logs       []string        `json:"logs,omitempty"`
	Dispatches int             `json:"dispatches"`
}

// ConversationCodeRuntime is an independent restricted execution host. It has
// no direct tool host, credentials, filesystem or network port. Every binding
// call must be returned through dispatch so Agent can apply its normal policy,
// budget, confirmation and durable receipt pipeline.
type ConversationCodeRuntime interface {
	ExecuteConversationCode(context.Context, ConversationCodeExecution, ConversationCodeDispatcher) (ConversationCodeResult, error)
}

type ConversationCodeFailure struct{ Code string }

func (e *ConversationCodeFailure) Error() string {
	if e == nil || e.Code == "" {
		return "conversation code execution failed"
	}
	return fmt.Sprintf("conversation code execution failed: %s", e.Code)
}

func ConversationCodeTool() ConversationToolDefinition {
	return ConversationToolDefinition{
		Key:            ConversationCodeToolKey,
		Version:        "1",
		Description:    "Run a bounded Lua program for programmatic tool composition and intermediate JSON data processing. Available tools are exposed under tools by their exact key. Each tools call is authorized, budgeted and receipted separately by the server. No filesystem, shell, environment, network, clock, random or dynamic module loading is available. The chunk must return exactly one JSON-compatible value; use log(value) for bounded progress details.",
		InputSchema:    json.RawMessage(`{"type":"object","properties":{"language":{"enum":["lua"]},"source":{"type":"string","minLength":1,"maxLength":32768}},"required":["language","source"],"additionalProperties":false}`),
		OutputSchema:   json.RawMessage(`{"type":"object","properties":{"language":{"const":"lua"},"value":{},"logs":{"type":"array","maxItems":128,"items":{"type":"string","maxLength":2048}},"dispatches":{"type":"integer","minimum":0,"maximum":64}},"required":["language","value","dispatches"],"additionalProperties":false}`),
		ActionKey:      ConversationToolActionPrefix + ConversationCodeToolKey,
		Effect:         "read",
		Idempotency:    "natural",
		TimeoutMillis:  30000,
		MaxOutputBytes: 64 * 1024,
	}
}
