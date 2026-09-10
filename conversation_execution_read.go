package agentsdk

import "encoding/json"

// ExecutionRead locates recorded work through server-issued run IDs. It does
// not execute, resume, or authorize the original operation.
type ConversationExecutionRead struct {
	ConversationID string `json:"conversation_id"`
	RunID          string `json:"run_id"`
	Cursor         string `json:"cursor,omitempty"`
	Limit          int    `json:"limit,omitempty"`
}

type ConversationExecutionEntry struct {
	Step       int                          `json:"step"`
	CallID     string                       `json:"call_id"`
	Tool       string                       `json:"tool"`
	RecordHash string                       `json:"record_hash"`
	State      string                       `json:"state"`
	Status     string                       `json:"status,omitempty"`
	ResourceID string                       `json:"resource_id,omitempty"`
	ErrorCode  string                       `json:"error_code,omitempty"`
	Reference  *ConversationResultReference `json:"reference,omitempty"`
}

type ConversationExecutionReadResult struct {
	ConversationID string                       `json:"conversation_id"`
	RunID          string                       `json:"run_id"`
	RunStatus      string                       `json:"run_status"`
	Items          []ConversationExecutionEntry `json:"items"`
	NextCursor     string                       `json:"next_cursor,omitempty"`
	Complete       bool                         `json:"complete"`
	Omitted        bool                         `json:"omitted"`
}

func ConversationExecutionReadDefinition() ConversationToolDefinition {
	return ConversationToolDefinition{Key: "execution_read", Version: "1", ActionKey: ConversationToolActionPrefix + "execution_read", Effect: "read", Idempotency: "natural", TimeoutMillis: 10000, MaxOutputBytes: 32 * 1024,
		Description: "Inspect recorded tool outcomes for a run_id and conversation_id supplied by recent execution links or history_search/history_read. Follow next_cursor until complete=true. Entries are historical observations, not current business status; completed pagination does not mean the work succeeded. Omitted=true means some entries cannot currently be disclosed. Read full result content with tool_result_read using an entry's exact reference. No operation is repeated or resumed. If execution_cursor_changed is returned restart from the first page. Do not guess IDs.",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"conversation_id":{"type":"string","minLength":1,"maxLength":96},"run_id":{"type":"string","minLength":1,"maxLength":96},"cursor":{"type":"string","maxLength":2048},"limit":{"type":"integer","minimum":1,"maximum":5}},"required":["conversation_id","run_id"],"additionalProperties":false}`), OutputSchema: json.RawMessage(`{"type":"object"}`)}
}
