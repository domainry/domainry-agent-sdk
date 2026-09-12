package agentsdk

import (
	"context"
	"encoding/json"
)

// Parameters come from the published action detail and actual business reads.
// Version binds the concrete action contract, not the generic tool version.
type ConversationBusinessAction struct {
	ObjectKey string          `json:"object_key"`
	ActionKey string          `json:"action_key"`
	Version   string          `json:"action_version"`
	RecordID  string          `json:"record_id,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// Server-owned execution metadata. None of the identity, idempotency or
// confirmation fields may be populated from model arguments.
type ConversationBusinessActionRequest struct {
	Authority      ConversationAuthority
	Action         ConversationBusinessAction
	ConversationID string
	RunID          string
	CorrelationID  string
	Step           int
	CallID         string
	IdempotencyKey string
	Confirmation   *ConversationConfirmation
	Arguments      string // exact frozen JSON covered by Confirmation.ArgumentsHash
}

type ConversationBusinessRecordReference struct {
	ObjectKey string `json:"object_key"`
	RecordID  string `json:"record_id"`
}

// Acknowledgements carry actual effect references, not ungoverned handler
// output or raw business records. Read values through the authorized read tools.
type ConversationBusinessActionResult struct {
	Status              string                                `json:"status"` // completed, failed, uncertain
	ErrorCode           string                                `json:"error_code,omitempty"`
	InvocationID        string                                `json:"invocation_id,omitempty"`
	ObjectKey           string                                `json:"object_key"`
	ActionKey           string                                `json:"action_key"`
	RecordID            string                                `json:"record_id,omitempty"`
	CreatedRecords      []ConversationBusinessRecordReference `json:"created_records,omitempty"`
	UpdatedRecords      []ConversationBusinessRecordReference `json:"updated_records,omitempty"`
	DeletedRecords      []ConversationBusinessRecordReference `json:"deleted_records,omitempty"`
	RestoredRecords     []ConversationBusinessRecordReference `json:"restored_records,omitempty"`
	ReferenceCount      int                                   `json:"reference_count"`
	ReferencesTruncated bool                                  `json:"references_truncated,omitempty"`
}

// Optional extension. An empty action in Authorize is a capability discovery
// request only; a nonempty action requires current business/record permission.
// Invoke/Reconcile must enforce the published contract and host idempotency.
// Reconcile must never reclaim or repeat an existing unresolved execution.
// It may start a conclusively absent execution only through an atomic host
// create-or-replay boundary that cannot reclaim an intervening execution.
// Revalidate is read-only and must verify the entire saved acknowledgement
// against the owned host receipt and today's access, without executing again.
type ConversationBusinessActionSource interface {
	AuthorizeBusinessAction(context.Context, ConversationBusinessAction, ConversationAuthority) (ConversationToolAuthorization, error)
	InvokeBusinessAction(context.Context, ConversationBusinessActionRequest) (ConversationBusinessActionResult, error)
	ReconcileBusinessAction(context.Context, ConversationBusinessActionRequest) (ConversationBusinessActionResult, error)
	RevalidateBusinessAction(context.Context, ConversationBusinessEvidence, ConversationAuthority) error
}

func BusinessActionConversationTools() []ConversationToolDefinition {
	return []ConversationToolDefinition{{
		Key: "invoke_action", Version: "1", ActionKey: ConversationToolActionPrefix + "invoke_action", Effect: "write", Idempotency: "reconcile", TimeoutMillis: 60000, MaxOutputBytes: 64 * 1024,
		Description:  "Execute a declared business action after the authenticated user confirms the exact target and payload. First expand the action with business_catalog kind=actions and use its execution_version, object key, target kind and input schema. Use an actual record_id for record actions; omit it for object actions. Supply only the declared business payload, including a published optimistic-concurrency value when required. Never invent IDs, versions, credentials, identity, approval, or idempotency keys. A confirmation is not a substitute for host business permission or additional assurance. completed means the host committed this action; uncertain means the outcome needs reconciliation, not permission to issue another call. Use authorized read tools to inspect resulting records.",
		InputSchema:  json.RawMessage(`{"type":"object","properties":{"object_key":{"type":"string","minLength":1,"maxLength":128},"action_key":{"type":"string","minLength":1,"maxLength":128},"action_version":{"type":"string","minLength":1,"maxLength":256},"record_id":{"type":"string","minLength":1,"maxLength":256},"data":{"type":"object","maxProperties":100}},"required":["object_key","action_key","action_version","data"],"additionalProperties":false}`),
		OutputSchema: json.RawMessage(`{"type":"object"}`),
	}}
}
