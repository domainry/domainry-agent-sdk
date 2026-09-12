package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Reads metadata only, scoped to the current owner and parent conversation.
// Implementations return at most 50 ready records (the per-conversation upload
// quota), fail on overflow, and never include deleted attachments or originals.
type ConversationAttachmentKnowledgeRepository interface {
	AttachmentKnowledgeRecords(context.Context, string, agentsdk.ConversationAuthority) ([]ConversationAttachmentRecord, error)
}
