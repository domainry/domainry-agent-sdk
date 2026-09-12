package agentsdk

import "context"

// ConversationExecutionAuthorizationRequest contains trusted owner identifiers,
// never a browser credential or a previously resolved policy bundle. Stage is
// assigned by the application: send, resume, respond, execute, model, summary,
// tool or commit. RunID is empty only before a new run is enqueued.
type ConversationExecutionAuthorizationRequest struct {
	Authority      ConversationAuthority
	ConversationID string
	RunID          string
	Stage          string
}

// ConversationExecutionAuthorizer re-resolves the current owner according to
// the host's execution admission policy, independently of tool access. Current
// conversation send/resume actions require an authenticated active principal;
// response and tool/resource permissions are additional checks. Admission to
// enqueue is not a grant for a later worker attempt.
// Implementations must honor ctx; errors and false both stop execution.
type ConversationExecutionAuthorizer interface {
	AuthorizeConversationExecution(context.Context, ConversationExecutionAuthorizationRequest) (bool, error)
}
