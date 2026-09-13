package agentsdk

import "context"

const KnowledgeResultReadUnsupportedCode = "agent.conversation.knowledge_result_read_unsupported"

// An explicit source-owner policy for reading a previously released receipt.
// It must validate all saved data, current source access and owner/configuration
// identity independently of the producing tool's execution grant. This port
// does not authorize running a tool or expose private conversation attachments.
type ConversationKnowledgeResultReadSource interface {
	AuthorizeKnowledgeResultRead(context.Context, ConversationKnowledgeResult, ConversationAuthority) error
}

func AuthorizeKnowledgeResultRead(ctx context.Context, source any, saved ConversationKnowledgeResult, a ConversationAuthority) error {
	if saved.ConversationID != "" {
		return &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
	}
	for _, citation := range saved.Citations {
		if citation.ConversationID != "" {
			return &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
		}
	}
	reader, ok := source.(ConversationKnowledgeResultReadSource)
	if !ok {
		return &Error{Class: "unavailable", Code: KnowledgeResultReadUnsupportedCode}
	}
	return reader.AuthorizeKnowledgeResultRead(ctx, saved, a)
}
