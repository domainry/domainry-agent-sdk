package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// ConversationProvenanceRepository atomically appends a completed externally
// executed turn to an Agent-owned conversation. Implementations must enforce
// idempotency using publication.ClientID and bind the resulting Run to the
// supplied Conversation owner.
type ConversationProvenanceRepository interface {
	ImportCompletedConversationRun(context.Context, string, agentsdk.ConversationProvenancePublication, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
}
