package agentsdk

import (
	"context"
	"encoding/json"
)

const ConversationCodingToolPrefix = "coding_"

// ConversationCodingScope is assigned by Agent from the authenticated run.
// The model can address only opaque resources already owned by this scope.
type ConversationCodingScope struct {
	RuntimeID, WorkspaceID, UserID string
	ConversationID, RunID          string
}

// ConversationCodingRequest is the trusted Runtime call produced after Agent
// has validated the model arguments, checked current authorization and, for a
// mutation, resolved the durable user confirmation.
type ConversationCodingRequest struct {
	Scope          ConversationCodingScope
	Tool           string
	Arguments      json.RawMessage
	IdempotencyKey string
	Reconcile      bool
}

// ConversationCodingRuntime owns the restricted filesystem/process execution
// world. CloseScope ends PTYs and background processes without deleting files.
type ConversationCodingRuntime interface {
	ExecuteConversationCoding(context.Context, ConversationCodingRequest) (ConversationToolResult, error)
	CloseConversationCodingScope(context.Context, ConversationCodingScope) error
}

func IsConversationCodingTool(key string) bool {
	for _, definition := range ConversationCodingTools() {
		if definition.Key == key {
			return true
		}
	}
	return false
}

func ConversationCodingTools() []ConversationToolDefinition {
	type spec struct {
		key, description, input, effect, idempotency string
		timeout                                      int
	}
	const path = `"path":{"type":"string","minLength":1,"maxLength":2048}`
	specs := []spec{
		{"coding_file_read", "Read a regular UTF-8 workspace file after resolving it inside the configured workspace. The result includes a SHA-256 version used by write/edit guards.", `{"type":"object","properties":{` + path + `,"offset":{"type":"integer","minimum":0,"maximum":1073741824},"max_bytes":{"type":"integer","minimum":1,"maximum":65536}},"required":["path"],"additionalProperties":false}`, "read", "natural", 10000},
		{"coding_file_search", "Search literal text in regular workspace files. Results are bounded and paths never escape the configured workspace.", `{"type":"object","properties":{"query":{"type":"string","minLength":1,"maxLength":1024},"path":{"type":"string","maxLength":2048},"glob":{"type":"string","maxLength":256},"limit":{"type":"integer","minimum":1,"maximum":200}},"required":["query"],"additionalProperties":false}`, "read", "natural", 30000},
		{"coding_file_write", "Create or replace a UTF-8 workspace file. expected_sha256 must be the exact version returned by coding_file_read, or 'missing' when creating a new file.", `{"type":"object","properties":{` + path + `,"content":{"type":"string","maxLength":1048576},"expected_sha256":{"type":"string","minLength":7,"maxLength":64},"create_directories":{"type":"boolean"}},"required":["path","content","expected_sha256"],"additionalProperties":false}`, "write", "reconcile", 30000},
		{"coding_file_edit", "Replace exact text in a UTF-8 workspace file using the SHA-256 version returned by coding_file_read. The edit fails on stale versions or ambiguous matches unless replace_all is true.", `{"type":"object","properties":{` + path + `,"old_text":{"type":"string","minLength":1,"maxLength":262144},"new_text":{"type":"string","maxLength":262144},"expected_sha256":{"type":"string","minLength":64,"maxLength":64},"replace_all":{"type":"boolean"}},"required":["path","old_text","new_text","expected_sha256"],"additionalProperties":false}`, "write", "reconcile", 30000},
		{"coding_terminal_open", "Open a persistent sandboxed PTY using the deployment's fixed shell. The terminal belongs to this exact Agent run.", `{"type":"object","properties":{"columns":{"type":"integer","minimum":20,"maximum":400},"rows":{"type":"integer","minimum":5,"maximum":200}},"additionalProperties":false}`, "write", "reconcile", 10000},
		{"coding_terminal_send", "Write text to an existing run-owned PTY. Use coding_terminal_read with its cursor to observe output.", `{"type":"object","properties":{"terminal_id":{"type":"string","minLength":1,"maxLength":96},"input":{"type":"string","minLength":1,"maxLength":65536}},"required":["terminal_id","input"],"additionalProperties":false}`, "write", "reconcile", 10000},
		{"coding_terminal_read", "Read bounded output from an existing run-owned PTY using a byte cursor.", `{"type":"object","properties":{"terminal_id":{"type":"string","minLength":1,"maxLength":96},"cursor":{"type":"integer","minimum":0},"max_bytes":{"type":"integer","minimum":1,"maximum":65536}},"required":["terminal_id"],"additionalProperties":false}`, "read", "natural", 10000},
		{"coding_terminal_close", "Close an existing run-owned PTY and terminate its shell process.", `{"type":"object","properties":{"terminal_id":{"type":"string","minLength":1,"maxLength":96}},"required":["terminal_id"],"additionalProperties":false}`, "write", "reconcile", 10000},
		{"coding_process_start", "Start a sandboxed background process from an argv array inside the workspace. The process belongs to this exact Agent run.", `{"type":"object","properties":{"argv":{"type":"array","minItems":1,"maxItems":64,"items":{"type":"string","minLength":1,"maxLength":4096}},"working_directory":{"type":"string","maxLength":2048}},"required":["argv"],"additionalProperties":false}`, "write", "reconcile", 10000},
		{"coding_process_read", "Read bounded stdout/stderr from a run-owned background process using a byte cursor.", `{"type":"object","properties":{"process_id":{"type":"string","minLength":1,"maxLength":96},"cursor":{"type":"integer","minimum":0},"max_bytes":{"type":"integer","minimum":1,"maximum":65536}},"required":["process_id"],"additionalProperties":false}`, "read", "natural", 10000},
		{"coding_process_kill", "Terminate a run-owned background process.", `{"type":"object","properties":{"process_id":{"type":"string","minLength":1,"maxLength":96}},"required":["process_id"],"additionalProperties":false}`, "write", "reconcile", 10000},
		{"coding_lsp", "Use the configured language server for exact definition or reference navigation. Lines and characters are one-based UTF-16 positions.", `{"type":"object","properties":{"operation":{"enum":["definition","references"]},` + path + `,"line":{"type":"integer","minimum":1,"maximum":10000000},"character":{"type":"integer","minimum":1,"maximum":10000000}},"required":["operation","path","line","character"],"additionalProperties":false}`, "read", "natural", 30000},
	}
	out := make([]ConversationToolDefinition, 0, len(specs))
	for _, item := range specs {
		parallelism := ""
		if item.effect == "read" && item.key != "coding_terminal_read" && item.key != "coding_process_read" {
			parallelism = "independent_read"
		}
		out = append(out, ConversationToolDefinition{Key: item.key, Version: "1", Description: item.description, InputSchema: json.RawMessage(item.input), OutputSchema: json.RawMessage(`{"type":"object"}`), ActionKey: ConversationToolActionPrefix + item.key, Effect: item.effect, Idempotency: item.idempotency, Parallelism: parallelism, TimeoutMillis: item.timeout, MaxOutputBytes: 512 * 1024})
	}
	return out
}
