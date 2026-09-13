package agentsdk

import "context"

const BusinessResultReadUnsupportedCode = "agent.conversation.business_result_read_unsupported"

// This optional source-owner policy reads an explicitly released, immutable
// business result independently of the producing Agent tool's execution grant.
// The owner must validate the complete evidence and current object, record,
// field and process access under the actual reader. It must never execute or
// reconcile a write. Unsupported operations retain the ordinary execution path.
type ConversationBusinessResultReadSource interface {
	AuthorizeBusinessResultRead(context.Context, ConversationBusinessEvidence, ConversationAuthority) error
}

func AuthorizeBusinessResultRead(ctx context.Context, source any, saved ConversationBusinessEvidence, a ConversationAuthority) error {
	reader, ok := source.(ConversationBusinessResultReadSource)
	if !ok {
		return &Error{Class: "unavailable", Code: BusinessResultReadUnsupportedCode}
	}
	return reader.AuthorizeBusinessResultRead(ctx, saved, a)
}
