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

// The request authority remains the current reader. Producer identifies the
// original signed snapshot or immutable ledger, never an execution grant.
type ConversationBusinessSharedResultReadSource interface {
	AuthorizeSharedBusinessResultRead(context.Context, ConversationBusinessEvidence, ConversationAuthority, ConversationAuthority) error
}

func AuthorizeSharedBusinessResultRead(ctx context.Context, source any, saved ConversationBusinessEvidence, reader, producer ConversationAuthority) error {
	if !reader.Known || !producer.Known || reader.RuntimeID == "" || reader.WorkspaceID == "" || reader.UserID == "" || producer.RuntimeID != reader.RuntimeID || producer.WorkspaceID != reader.WorkspaceID || producer.UserID == "" {
		return &Error{Class: "forbidden", Code: "agent.conversation.business_access_denied"}
	}
	if producer == reader {
		return AuthorizeBusinessResultRead(ctx, source, saved, reader)
	}
	owner, ok := source.(ConversationBusinessSharedResultReadSource)
	if !ok {
		// Existing user-bound owners already authorize the current role. They
		// remain valid for the same user, but never substitute a foreign user.
		if producer.UserID == reader.UserID {
			return AuthorizeBusinessResultRead(ctx, source, saved, reader)
		}
		return &Error{Class: "unavailable", Code: BusinessResultReadUnsupportedCode}
	}
	return owner.AuthorizeSharedBusinessResultRead(ctx, saved, reader, producer)
}

func AuthorizeBusinessResultRead(ctx context.Context, source any, saved ConversationBusinessEvidence, a ConversationAuthority) error {
	reader, ok := source.(ConversationBusinessResultReadSource)
	if !ok {
		return &Error{Class: "unavailable", Code: BusinessResultReadUnsupportedCode}
	}
	return reader.AuthorizeBusinessResultRead(ctx, saved, a)
}
