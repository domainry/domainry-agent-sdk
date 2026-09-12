package agentsdk

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const ConversationTaskStatusQueued = "queued"
const ConversationTaskStatusRunning = "running"
const ConversationTaskStatusCompleted = "completed"
const ConversationTaskStatusFailed = "failed"
const ConversationTaskStatusCancelled = "cancelled"

// ConversationTaskBudget is frozen when task_start is accepted. Every value is
// also bounded by the selected Agent profile and deployment configuration.
type ConversationTaskBudget struct {
	MaxSteps       int `json:"max_steps"`
	MaxToolCalls   int `json:"max_tool_calls"`
	MaxOutputBytes int `json:"max_output_bytes"`
	TimeoutSeconds int `json:"timeout_seconds"`
}

// ConversationTaskToolScope records the exact tool contract and authorization
// revision that was allowed when a task was accepted. Execution still performs
// current authorization and availability checks before each use.
type ConversationTaskToolScope struct {
	Key                   string `json:"key"`
	Version               string `json:"version"`
	ActionKey             string `json:"action_key"`
	DefinitionHash        string `json:"definition_hash"`
	AuthorizationRevision string `json:"authorization_revision,omitempty"`
}

type ConversationTaskStart struct {
	Goal         string                 `json:"goal"`
	Input        string                 `json:"input"`
	AllowedTools []string               `json:"allowed_tools"`
	Budget       ConversationTaskBudget `json:"budget"`
}

// ConversationTask is an Agent-owned background execution requested from a
// durable conversation. Business effects remain owned by their tool services.
type ConversationTask struct {
	ID                   string                      `json:"id"`
	Status               string                      `json:"status"`
	Goal                 string                      `json:"goal"`
	Input                string                      `json:"input"`
	ToolScope            []ConversationTaskToolScope `json:"tool_scope"`
	Budget               ConversationTaskBudget      `json:"budget"`
	SourceConversationID string                      `json:"source_conversation_id"`
	SourceRunID          string                      `json:"source_run_id"`
	ExecutionRunID       string                      `json:"execution_run_id,omitempty"`
	ResultMessageID      string                      `json:"result_message_id,omitempty"`
	CompletionEventID    string                      `json:"completion_event_id,omitempty"`
	CompletionEventSeq   int64                       `json:"completion_event_seq,omitempty"`
	ErrorCode            string                      `json:"error_code,omitempty"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
	CompletedAt          *time.Time                  `json:"completed_at,omitempty"`
}

// ConversationTaskQuery is an owner-scoped, live task listing. A cursor is
// bound to the exact filters and owner that created it.
type ConversationTaskQuery struct {
	Query                string `json:"query,omitempty"`
	Status               string `json:"status,omitempty"`
	SourceConversationID string `json:"source_conversation_id,omitempty"`
	Cursor               string `json:"cursor,omitempty"`
	Limit                int    `json:"limit,omitempty"`
}

// ConversationTaskProgress is a bounded public projection of the normal
// ConversationRun used by a background task. It deliberately excludes the
// frozen definition hashes and authorization revisions kept by the executor.
type ConversationTaskProgress struct {
	RunStatus    string `json:"run_status,omitempty"`
	Attempt      int    `json:"attempt,omitempty"`
	Steps        int    `json:"steps"`
	ToolCalls    int    `json:"tool_calls"`
	LastEventSeq int64  `json:"last_event_seq,omitempty"`
}

type ConversationTaskWaiting struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Question  string    `json:"question"`
	Tool      string    `json:"tool,omitempty"`
	Revision  int64     `json:"revision"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ConversationTaskResult struct {
	MessageID string `json:"message_id"`
	Preview   string `json:"preview"`
	Bytes     int    `json:"bytes"`
	Complete  bool   `json:"complete"`
}

// ConversationTaskSummary is safe for HTTP and model-facing task_list. The
// source IDs are navigation references, never authority to read that source.
type ConversationTaskSummary struct {
	ID                   string                   `json:"id"`
	Status               string                   `json:"status"`
	Goal                 string                   `json:"goal"`
	AllowedTools         []string                 `json:"allowed_tools"`
	Budget               ConversationTaskBudget   `json:"budget"`
	SourceConversationID string                   `json:"source_conversation_id"`
	SourceRunID          string                   `json:"source_run_id"`
	ExecutionRunID       string                   `json:"execution_run_id,omitempty"`
	Progress             ConversationTaskProgress `json:"progress"`
	Waiting              *ConversationTaskWaiting `json:"waiting,omitempty"`
	Result               *ConversationTaskResult  `json:"result,omitempty"`
	Artifacts            []ConversationArtifact   `json:"artifacts"`
	ArtifactsComplete    bool                     `json:"artifacts_complete"`
	ArtifactsOmitted     bool                     `json:"artifacts_omitted,omitempty"`
	AccessError          string                   `json:"access_error,omitempty"`
	CompletionEventID    string                   `json:"completion_event_id,omitempty"`
	CompletionEventSeq   int64                    `json:"completion_event_seq,omitempty"`
	ErrorCode            string                   `json:"error_code,omitempty"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
	CompletedAt          *time.Time               `json:"completed_at,omitempty"`
}

// ConversationTaskDetail adds the complete bounded input, public execution
// steps and actionable waiting record to one task summary.
type ConversationTaskDetail struct {
	ConversationTaskSummary
	Input       string                   `json:"input"`
	Steps       []ConversationStepView   `json:"steps"`
	Interaction *ConversationInteraction `json:"interaction,omitempty"`
}

type ConversationTaskPage struct {
	Items      []ConversationTaskSummary `json:"items"`
	NextCursor string                    `json:"next_cursor,omitempty"`
	Complete   bool                      `json:"complete"`
}

// ConversationTaskService is an optional ConversationService extension.
type ConversationTaskService interface {
	ConversationTask(context.Context, string, ConversationAuthority) (ConversationTaskDetail, error)
	ConversationTasks(context.Context, ConversationTaskQuery, ConversationAuthority) (ConversationTaskPage, error)
}

type ConversationTaskReceipt struct {
	ID                   string                 `json:"id"`
	Status               string                 `json:"status"`
	AllowedTools         []string               `json:"allowed_tools"`
	Budget               ConversationTaskBudget `json:"budget"`
	SourceConversationID string                 `json:"source_conversation_id"`
	SourceRunID          string                 `json:"source_run_id"`
	CreatedAt            time.Time              `json:"created_at"`
}

func (t ConversationTask) Terminal() bool {
	return t.Status == ConversationTaskStatusCompleted || t.Status == ConversationTaskStatusFailed || t.Status == ConversationTaskStatusCancelled
}

func ConversationTaskPrompt(task ConversationTask) string {
	return "Goal:\n" + strings.TrimSpace(task.Goal) + "\n\nInput:\n" + task.Input
}

// ConversationTaskExecution is frozen onto the conversation run that executes
// a task. It limits the model-facing catalog and run budget; task_start itself
// is never available inside the child run.
type ConversationTaskExecution struct {
	TaskID    string                      `json:"task_id"`
	ToolScope []ConversationTaskToolScope `json:"tool_scope"`
	Budget    ConversationTaskBudget      `json:"budget"`
}

func BackgroundTaskConversationTool() ConversationToolDefinition {
	return ConversationToolDefinition{
		Key:            "task_start",
		Version:        "1",
		ActionKey:      ConversationToolActionPrefix + "task_start",
		Effect:         "write",
		Idempotency:    "key",
		Description:    "Start a separate durable background task only when the user asks work to continue asynchronously. Provide one explicit goal, complete bounded input, an exact subset of currently available tools, and all budget values. The accepted result contains a stable task ID but does not mean the work completed. The background run rechecks current permissions and tool availability; write tools may still wait for user confirmation. A background task cannot start another task.",
		InputSchema:    json.RawMessage(`{"type":"object","properties":{"goal":{"type":"string","minLength":1,"maxLength":2048},"input":{"type":"string","maxLength":8192},"allowed_tools":{"type":"array","maxItems":16,"uniqueItems":true,"items":{"type":"string","minLength":1,"maxLength":64,"pattern":"^[A-Za-z0-9_-]+$"}},"budget":{"type":"object","properties":{"max_steps":{"type":"integer","minimum":1,"maximum":32},"max_tool_calls":{"type":"integer","minimum":1,"maximum":32},"max_output_bytes":{"type":"integer","minimum":256,"maximum":65536},"timeout_seconds":{"type":"integer","minimum":1,"maximum":1800}},"required":["max_steps","max_tool_calls","max_output_bytes","timeout_seconds"],"additionalProperties":false}},"required":["goal","input","allowed_tools","budget"],"additionalProperties":false}`),
		OutputSchema:   json.RawMessage(`{"type":"object","properties":{"task":{"type":"object","properties":{"id":{"type":"string"},"status":{"enum":["queued"]},"allowed_tools":{"type":"array","items":{"type":"string"}},"budget":{"type":"object","properties":{"max_steps":{"type":"integer"},"max_tool_calls":{"type":"integer"},"max_output_bytes":{"type":"integer"},"timeout_seconds":{"type":"integer"}},"required":["max_steps","max_tool_calls","max_output_bytes","timeout_seconds"],"additionalProperties":false},"source_conversation_id":{"type":"string"},"source_run_id":{"type":"string"},"created_at":{"type":"string","format":"date-time"}},"required":["id","status","allowed_tools","budget","source_conversation_id","source_run_id","created_at"],"additionalProperties":false}},"required":["task"],"additionalProperties":false}`),
		TimeoutMillis:  10000,
		MaxOutputBytes: 32768,
	}
}

func BackgroundTaskQueryConversationTools() []ConversationToolDefinition {
	return []ConversationToolDefinition{
		{
			Key: "task_get", Version: "1", ActionKey: ConversationToolActionPrefix + "task_get", Effect: "read", Idempotency: "natural",
			Description:   "Read the current status, progress, waiting item, result preview and saved artifact references for a background task ID returned by task_start or task_list. Source conversation and run IDs are navigation references and do not grant access. A completed status means the durable completion event was committed; accepted or running does not.",
			InputSchema:   json.RawMessage(`{"type":"object","properties":{"id":{"type":"string","minLength":1,"maxLength":96,"pattern":"^[A-Za-z0-9_.:-]+$"}},"required":["id"],"additionalProperties":false}`),
			OutputSchema:  json.RawMessage(`{"type":"object","properties":{"task":{"type":"object"}},"required":["task"],"additionalProperties":false}`),
			TimeoutMillis: 10000, MaxOutputBytes: 65536,
		},
		{
			Key: "task_list", Version: "1", ActionKey: ConversationToolActionPrefix + "task_list", Effect: "read", Idempotency: "natural",
			Description:   "List the current user's durable background tasks and their live progress. Optional query is a literal goal substring; status filters one task state and scope=current_conversation limits results to this conversation. Follow next_cursor until complete=true. Use returned IDs with task_get; do not invent IDs or interpret accepted as completed.",
			InputSchema:   json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","maxLength":256},"status":{"enum":["queued","running","completed","failed","cancelled"]},"scope":{"enum":["all","current_conversation"]},"cursor":{"type":"string","maxLength":2048},"limit":{"type":"integer","minimum":1,"maximum":20}},"additionalProperties":false}`),
			OutputSchema:  json.RawMessage(`{"type":"object","properties":{"items":{"type":"array","items":{"type":"object"}},"next_cursor":{"type":"string"},"complete":{"type":"boolean"}},"required":["items","complete"],"additionalProperties":false}`),
			TimeoutMillis: 10000, MaxOutputBytes: 65536,
		},
	}
}
