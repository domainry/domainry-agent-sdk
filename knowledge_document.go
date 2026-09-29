package agentsdk

import (
	"context"
	"time"
)

// One immutable uploaded file belongs to one library. Sharing/moving creates a
// separate authorized lifecycle; conversation attachments are not library files.
type KnowledgeDocument struct {
	ID              string    `json:"id"`
	LibraryID       string    `json:"library_id"`
	Filename        string    `json:"filename"`
	ContentType     string    `json:"content_type"`
	Bytes           int64     `json:"bytes"`
	SHA256          string    `json:"sha256"`
	CreatedByUserID string    `json:"created_by_user_id"`
	State           string    `json:"state"` // uploading, queued, indexing, ready, failed, needs_reconcile, deleting, deleted
	IndexStatus     string    `json:"index_status,omitempty"`
	ErrorCode       string    `json:"error_code,omitempty"`
	Revision        int64     `json:"revision"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type KnowledgeDocumentUpload struct {
	ClientID string `json:"client_id"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

// KnowledgeDocumentSourceAccess binds a managed document to the source-owned
// business resource whose current read policy governs the document. It is
// supplied only through a trusted host port; ordinary library uploads do not
// accept this value from browser or model input.
type KnowledgeDocumentSourceAccess struct {
	Namespace    string `json:"namespace"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

const KnowledgeDocumentSourceNamespaceRuntimeRecord = "domainry.runtime.record.v1"

// KnowledgeDocumentSourceAuthorizer rechecks a persisted source binding for
// the actual reader. Library membership alone must not make a source-backed
// document readable after the underlying business resource is revoked.
type KnowledgeDocumentSourceAuthorizer interface {
	AuthorizeKnowledgeDocumentSource(context.Context, KnowledgeDocumentSourceAccess, ConversationAuthority) error
}

// KnowledgeDocumentSourceUploadService is the trusted host-only upload path
// for source-backed documents. Public Knowledge HTTP uploads continue to use
// KnowledgeDocumentService and cannot attach a forged source binding.
type KnowledgeDocumentSourceUploadService interface {
	UploadKnowledgeDocumentForSource(context.Context, string, KnowledgeDocumentUpload, KnowledgeDocumentSourceAccess, ConversationAuthority) (KnowledgeDocument, error)
}
type KnowledgeDocumentPage struct {
	Items     []KnowledgeDocument `json:"items"`
	NextAfter string              `json:"next_after,omitempty"`
	Complete  bool                `json:"complete"`
}
type KnowledgeDocumentDownload struct {
	Document KnowledgeDocument `json:"document"`
	Data     []byte            `json:"data"`
}
type KnowledgeDocumentService interface {
	UploadKnowledgeDocument(context.Context, string, KnowledgeDocumentUpload, ConversationAuthority) (KnowledgeDocument, error)
	KnowledgeDocuments(context.Context, string, string, int, ConversationAuthority) (KnowledgeDocumentPage, error)
	KnowledgeDocument(context.Context, string, string, ConversationAuthority) (KnowledgeDocument, error)
	DownloadKnowledgeDocument(context.Context, string, string, ConversationAuthority) (KnowledgeDocumentDownload, error)
	DeleteKnowledgeDocument(context.Context, string, string, int64, ConversationAuthority) (KnowledgeDocument, error)
}

// Storage is library-scoped, private, and independent of uploader membership.
// Runtime/workspace/library/document are trusted application facts. Deletion
// must permanently fence late puts, including after host restarts.
type KnowledgeDocumentStorageScope struct{ RuntimeID, WorkspaceID, LibraryID string }
type KnowledgeDocumentStorage interface {
	PutKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string, string, []byte) (string, error)
	ReadKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string, string) ([]byte, error)
	DeleteKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string) error
}

// Provider projection includes only the explicitly mapped source fields of one
// document passage. It must exclude unrelated/global response data. The Agent
// applies its local document allowlist before constructing model evidence.
type KnowledgeDocumentPassage struct {
	DocumentID string                 `json:"doc_id"`
	Title      string                 `json:"title,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Location   *DocumentLocation      `json:"location,omitempty"`
	Cells      []KnowledgePassageCell `json:"cells,omitempty"`
}
type ManagedKnowledgeDocumentSource interface {
	KnowledgeDocumentSource
	// Physical remote scope identity excludes credentials and user input. It
	// must stay stable across key rotation and field-mapping changes.
	KnowledgeDocumentSourceIdentity() string
	KnowledgeDocumentManagementReady() error
	SearchKnowledgeDocumentPassages(context.Context, string, ConversationAuthority) ([]KnowledgeDocumentPassage, error)
	ReadKnowledgeDocumentPassages(context.Context, string, ConversationAuthority) ([]KnowledgeDocumentPassage, error)
}

// Import creates an independent library copy of an owned conversation attachment.
// Source IDs never confer access. The target library and source download actions
// are authorized separately; deleting the source later does not delete the copy.
type KnowledgeAttachmentImport struct {
	ClientID         string `json:"client_id"`
	ConversationID   string `json:"conversation_id"`
	AttachmentID     string `json:"attachment_id"`
	ExpectedRevision int64  `json:"expected_revision"`
}
type KnowledgeAttachmentImportService interface {
	ImportConversationAttachment(context.Context, string, KnowledgeAttachmentImport, ConversationAuthority) (KnowledgeDocument, error)
}

// Copy/move creates an immutable target document in another library. A move
// revokes the source when the durable target original is committed, while both
// remote indexing and source cleanup proceed asynchronously.
type KnowledgeDocumentTransfer struct {
	ClientID         string `json:"client_id"`
	SourceLibraryID  string `json:"source_library_id"`
	SourceDocumentID string `json:"source_document_id"`
	ExpectedRevision int64  `json:"expected_revision"`
	Mode             string `json:"mode"` // copy, move
}
type KnowledgeDocumentTransferService interface {
	TransferKnowledgeDocument(context.Context, string, KnowledgeDocumentTransfer, ConversationAuthority) (KnowledgeDocument, error)
}
