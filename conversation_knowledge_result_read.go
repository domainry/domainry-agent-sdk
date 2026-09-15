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

// Reader remains the current request identity. Producer binds original scope
// and citations; it never grants execution or private attachment access.
type ConversationKnowledgeSharedResultReadSource interface {
	AuthorizeSharedKnowledgeResultRead(context.Context, ConversationKnowledgeResult, ConversationAuthority, ConversationAuthority) error
}

type KnowledgeSharedExtractionContentSource interface {
	SharedKnowledgeExtractionPassages(context.Context, ConversationKnowledgeResult, ConversationAuthority, ConversationAuthority) ([]KnowledgeDocumentPassage, error)
}

func validSharedKnowledgeRead(saved ConversationKnowledgeResult, reader, producer ConversationAuthority) bool {
	if !reader.Known || !producer.Known || reader.RuntimeID == "" || reader.WorkspaceID == "" || reader.UserID == "" || producer.RuntimeID != reader.RuntimeID || producer.WorkspaceID != reader.WorkspaceID || producer.UserID == "" || saved.ConversationID != "" {
		return false
	}
	for _, citation := range saved.Citations {
		if citation.ConversationID != "" {
			return false
		}
	}
	return true
}

func AuthorizeSharedKnowledgeResultRead(ctx context.Context, source any, saved ConversationKnowledgeResult, reader, producer ConversationAuthority) error {
	if !validSharedKnowledgeRead(saved, reader, producer) {
		return &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
	}
	if reader == producer {
		return AuthorizeKnowledgeResultRead(ctx, source, saved, reader)
	}
	if owner, ok := source.(ConversationKnowledgeSharedResultReadSource); ok {
		return owner.AuthorizeSharedKnowledgeResultRead(ctx, saved, reader, producer)
	}
	if reader.UserID == producer.UserID {
		return AuthorizeKnowledgeResultRead(ctx, source, saved, reader)
	}
	return &Error{Class: "unavailable", Code: KnowledgeResultReadUnsupportedCode}
}

func SharedKnowledgeExtractionPassages(ctx context.Context, source any, saved ConversationKnowledgeResult, reader, producer ConversationAuthority) ([]KnowledgeDocumentPassage, error) {
	if !validSharedKnowledgeRead(saved, reader, producer) {
		return nil, &Error{Class: "forbidden", Code: "agent.conversation.knowledge_access_denied"}
	}
	if reader != producer {
		if owner, ok := source.(KnowledgeSharedExtractionContentSource); ok {
			return owner.SharedKnowledgeExtractionPassages(ctx, saved, reader, producer)
		}
		if reader.UserID != producer.UserID {
			return nil, &Error{Class: "unavailable", Code: KnowledgeResultReadUnsupportedCode}
		}
	}
	owner, ok := source.(KnowledgeExtractionContentSource)
	if !ok {
		return nil, &Error{Class: "unavailable", Code: "agent.conversation.knowledge_extraction_content_unavailable"}
	}
	return owner.KnowledgeExtractionPassages(ctx, saved, reader)
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
