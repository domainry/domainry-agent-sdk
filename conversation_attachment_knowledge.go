package agentsdk

import "context"

// Attachment knowledge owns a dedicated physical KB. DocumentPermissionIDs
// describe the document visibility grants written upstream. ReadPermissionIDs
// describe the current authenticated principal identities presented on reads.
// Neither browser nor model may supply either set. Original bytes go to the
// Connector. The Agent does not parse or index them.
type ConversationAttachmentKnowledge interface {
	AttachmentKnowledgeSourceIdentity() string
	ResolveAttachmentKnowledge(context.Context, string, ConversationAuthority, ConversationAttachmentPermissionScope) (ConversationAttachmentKnowledgeScope, error)
}

// ConversationAttachmentPermissionScope is resolved from the live Identity
// principal. User and workspace identities come from ConversationAuthority;
// OrganizationIDs contains every current organization scope the principal may
// use for KB reads. Raw values never leave the trusted host boundary.
type ConversationAttachmentPermissionScope struct {
	OrganizationIDs []string
}

type ConversationAttachmentPermissionResolver interface {
	ResolveConversationAttachmentPermissions(context.Context, ConversationAuthority) (ConversationAttachmentPermissionScope, error)
}

type ConversationAttachmentKnowledgeSource interface {
	ManagedKnowledgeDocumentSource
	KnowledgeDocumentAccessPolicySource
	KnowledgeDocumentSizeLimitSource
}

type ConversationAttachmentKnowledgeScope struct {
	Source                ConversationAttachmentKnowledgeSource
	DocumentPermissionIDs []string
	ReadPermissionIDs     []string
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
