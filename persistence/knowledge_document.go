package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Import provenance stays server-side: shared readers must not learn private
// source conversation identifiers. Once committed this is an independent copy.
type KnowledgeAttachmentOrigin struct {
	ConversationID string `json:"conversation_id"`
	AttachmentID   string `json:"attachment_id"`
	Revision       int64  `json:"revision"`
}
type KnowledgeAttachmentImportRepository interface {
	FindKnowledgeAttachmentImport(context.Context, string, string, KnowledgeAttachmentOrigin, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, bool, error)
}
type KnowledgeDocumentOrigin struct {
	LibraryID  string `json:"library_id"`
	DocumentID string `json:"document_id"`
	Revision   int64  `json:"revision"`
	Mode       string `json:"mode"`
}
type KnowledgeDocumentTransferRepository interface {
	FindKnowledgeDocumentTransfer(context.Context, string, string, KnowledgeDocumentOrigin, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, bool, error)
}
type KnowledgeDocumentRecord struct {
	DocumentOrigin     *KnowledgeDocumentOrigin       `json:"document_origin,omitempty"`
	AttachmentOrigin   *KnowledgeAttachmentOrigin     `json:"attachment_origin,omitempty"`
	Document           agentsdk.KnowledgeDocument     `json:"document"`
	RequestSHA256      string                         `json:"request_sha256"`
	SourceID           string                         `json:"source_id"`
	AccessPolicySHA256 string                         `json:"access_policy_sha256,omitempty"`
	PutRequestID       string                         `json:"put_request_id,omitempty"`
	RemoteID           string                         `json:"remote_id"`
	BodyRef            string                         `json:"body_ref,omitempty"`
	Actor              agentsdk.ConversationAuthority `json:"actor"`
	PutStarted         bool                           `json:"put_started"`
	PutAcknowledged    bool                           `json:"put_acknowledged"`
	IndexObserved      bool                           `json:"index_observed"`
	DeleteStarted      bool                           `json:"delete_started"`
	DeleteAcknowledged bool                           `json:"delete_acknowledged,omitempty"`
}
type KnowledgeDocumentReserve struct {
	DocumentOrigin                                               *KnowledgeDocumentOrigin   `json:",omitempty"`
	AttachmentOrigin                                             *KnowledgeAttachmentOrigin `json:",omitempty"`
	LibraryID, ClientID, Filename, ContentType, SHA256, SourceID string
	Bytes                                                        int64
	AccessPolicySHA256                                           string `json:",omitempty"`
}
type KnowledgeDocumentLease struct {
	RuntimeID   string    `json:"runtime_id"`
	WorkspaceID string    `json:"workspace_id"`
	LibraryID   string    `json:"library_id"`
	DocumentID  string    `json:"document_id"`
	Owner       string    `json:"owner"`
	Token       int64     `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
type KnowledgeDocumentProgress struct {
	Event, IndexStatus, ErrorCode string
	RetryAt                       time.Time
}

// Management markers and tombstones are durable. A previously managed KB may
// not silently become an unrestricted legacy source when configuration changes.
// Worker methods are trusted runtime maintenance, never user HTTP operations.
type KnowledgeDocumentRepository interface {
	ActivateKnowledgeDocumentSource(context.Context, agentsdk.KnowledgeDocumentStorageScope, string) error
	KnowledgeDocumentLibrarySource(context.Context, string, agentsdk.ConversationAuthority) (string, error)
	KnowledgeSourceManaged(context.Context, string) (bool, error)
	ReserveKnowledgeDocument(context.Context, KnowledgeDocumentReserve, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, error)
	CommitKnowledgeDocumentContent(context.Context, string, int64, string, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, error)
	KnowledgeDocumentRecord(context.Context, string, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, error)
	KnowledgeDocumentByRemoteID(context.Context, string, string, string, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, error)
	KnowledgeDocuments(context.Context, string, string, int, agentsdk.ConversationAuthority) (agentsdk.KnowledgeDocumentPage, error)
	RequestKnowledgeDocumentDeletion(context.Context, string, int64, agentsdk.ConversationAuthority) (KnowledgeDocumentRecord, error)
	ClaimKnowledgeDocumentWork(context.Context, string, string, time.Time, time.Duration) (KnowledgeDocumentLease, bool, error)
	KnowledgeDocumentWorkRecord(context.Context, KnowledgeDocumentLease) (KnowledgeDocumentRecord, error)
	StartKnowledgeDocumentPut(context.Context, KnowledgeDocumentLease) (KnowledgeDocumentRecord, bool, error)
	StartKnowledgeDocumentDelete(context.Context, KnowledgeDocumentLease) error
	ApplyKnowledgeDocumentProgress(context.Context, KnowledgeDocumentLease, KnowledgeDocumentProgress) error
}
