package agentsdk

import "context"

// ConversationExecutionSubject identifies the actual principal whose current
// permissions authorize a task. It is independent of the Agent configuration
// owner. Roles and access bundles are resolved afresh, never copied into it.
type ConversationExecutionSubject struct {
	RuntimeID   string `json:"runtime_id"`
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
}

// ConversationAgentSubjectValidator validates current, active membership in
// the caller's workspace. The service separately authorizes configuration,
// sharing and use, and the repository checks explicit Agent sharing grants.
// This port cannot select an execution principal or grant source data access.
type ConversationAgentSubjectValidator interface {
	ValidateConversationAgentSubjects(context.Context, ConversationAuthority, []string) error
}
