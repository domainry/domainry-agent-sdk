package agentsdk

import "context"

// Attachment knowledge owns a dedicated physical KB. Every resolved source
// must use a private scope derived from the trusted runtime/workspace/user/
// conversation tuple; neither browser nor model may supply upstream ACL IDs.
// Original bytes go to the Connector. The Agent does not parse or index them.
type ConversationAttachmentKnowledge interface {
	AttachmentKnowledgeSourceIdentity() string
	ResolveAttachmentKnowledge(context.Context, string, ConversationAuthority) (ConversationAttachmentKnowledgeScope, error)
}

type ConversationAttachmentKnowledgeSource interface {
	ManagedKnowledgeDocumentSource
	KnowledgeDocumentAccessPolicySource
	KnowledgeDocumentSizeLimitSource
}

type ConversationAttachmentKnowledgeScope struct {
	Source       ConversationAttachmentKnowledgeSource
	PermissionID string
}

type ConversationAttachmentKnowledgeBinding struct {
	WorkspaceID string
	Knowledge   ConversationAttachmentKnowledge
}

// Indexing is an explicit file-management action, separate from uploading an
// original or saving an independent personal/shared library copy.
type ConversationAttachmentIndexService interface {
	IndexAttachment(context.Context, string, string, int64, ConversationAuthority) (ConversationAttachment, error)
}

// Expedites the existing durable task; it never clears write markers, chooses
// a new remote ID or blindly repeats a write whose result is unknown.
type ConversationAttachmentIndexCheckService interface {
	CheckAttachmentIndex(context.Context, string, string, int64, ConversationAuthority) (ConversationAttachment, error)
}
