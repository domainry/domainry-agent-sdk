package agentsdk

import (
	"context"
	"io"
	"time"
)

// KnowledgeDocumentMaximumBytes is the host safety ceiling for one immutable
// original. Provider-specific limits may be smaller, but the core contract is
// intentionally not tied to the former in-memory attachment limit.
const KnowledgeDocumentMaximumBytes int64 = 1 << 30

// KnowledgeDocumentUploadChunkBytes is the wire/storage part size for a
// resumable browser upload. Every non-final part has this exact size, which
// makes an upload position deterministic without trusting a client cursor.
const KnowledgeDocumentUploadChunkBytes int64 = 8 << 20

// One immutable uploaded file belongs to one library. Sharing/moving creates a
// separate authorized lifecycle; conversation attachments are not library files.
type KnowledgeDocument struct {
	ID                   string    `json:"id"`
	LibraryID            string    `json:"library_id"`
	ProjectionKind       string    `json:"projection_kind"` // file, page
	OriginPageID         string    `json:"origin_page_id,omitempty"`
	OriginPageVersion    int64     `json:"origin_page_version,omitempty"`
	Filename             string    `json:"filename"`
	ContentType          string    `json:"content_type"`
	Bytes                int64     `json:"bytes"`
	SHA256               string    `json:"sha256"`
	ContentVersion       int64     `json:"content_version"` // Immutable-original version used by citations and stale-read checks.
	IndexGeneration      int64     `json:"index_generation"`
	Reindexing           bool      `json:"reindexing,omitempty"`
	CreatedByUserID      string    `json:"created_by_user_id"`
	State                string    `json:"state"`            // uploading, quarantined, scanning, queued, indexing, ready, failed, needs_reconcile, deleting, deleted
	ProcessingStage      string    `json:"processing_stage"` // upload, scan, parse, chunk, index, complete, failed, delete
	SecurityStatus       string    `json:"security_status"`  // pending, scanning, clean, rejected
	SecurityScanner      string    `json:"security_scanner,omitempty"`
	SecurityReportSHA256 string    `json:"security_report_sha256,omitempty"`
	IndexStatus          string    `json:"index_status,omitempty"`
	ErrorCode            string    `json:"error_code,omitempty"`
	Revision             int64     `json:"revision"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
type KnowledgeDocumentUpload struct {
	ClientID string    `json:"client_id"`
	Filename string    `json:"filename"`
	Bytes    int64     `json:"bytes"`
	SHA256   string    `json:"sha256"`
	Content  io.Reader `json:"-"`
}

// KnowledgeDocumentUploadChunk carries one independently verified part of an
// immutable original. Bytes and SHA256 describe the complete original;
// ChunkBytes and ChunkSHA256 describe only Content at Offset.
type KnowledgeDocumentUploadChunk struct {
	ClientID    string    `json:"client_id"`
	Filename    string    `json:"filename"`
	Bytes       int64     `json:"bytes"`
	SHA256      string    `json:"sha256"`
	Offset      int64     `json:"offset"`
	ChunkBytes  int64     `json:"chunk_bytes"`
	ChunkSHA256 string    `json:"chunk_sha256"`
	Content     io.Reader `json:"-"`
}

// KnowledgeDocumentUploadProgress is returned after every part. Document is
// the current uploading or committed record; Complete is true only after the
// complete original has been verified and committed to the library.
type KnowledgeDocumentUploadProgress struct {
	NextOffset int64              `json:"next_offset"`
	Complete   bool               `json:"complete"`
	Document   *KnowledgeDocument `json:"document,omitempty"`
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
	Document KnowledgeDocument
	Content  io.ReadSeekCloser
}
type KnowledgeDocumentService interface {
	UploadKnowledgeDocument(context.Context, string, KnowledgeDocumentUpload, ConversationAuthority) (KnowledgeDocument, error)
	KnowledgeDocuments(context.Context, string, string, int, ConversationAuthority) (KnowledgeDocumentPage, error)
	KnowledgeDocument(context.Context, string, string, ConversationAuthority) (KnowledgeDocument, error)
	DownloadKnowledgeDocument(context.Context, string, string, ConversationAuthority) (KnowledgeDocumentDownload, error)
	DeleteKnowledgeDocument(context.Context, string, string, int64, ConversationAuthority) (KnowledgeDocument, error)
}

// KnowledgeDocumentChunkUploadService is the public large-file upload path.
// Whole-file UploadKnowledgeDocument remains the trusted in-process path used
// when bytes already belong to another server-side resource.
type KnowledgeDocumentChunkUploadService interface {
	UploadKnowledgeDocumentChunk(context.Context, string, KnowledgeDocumentUploadChunk, ConversationAuthority) (KnowledgeDocumentUploadProgress, error)
}

// KnowledgeDocumentRetryService retries only the durable stage already owned
// by Knowledge. It never turns an uncertain provider write into a second write.
type KnowledgeDocumentRetryService interface {
	RetryKnowledgeDocument(context.Context, string, string, int64, ConversationAuthority) (KnowledgeDocument, error)
}

// KnowledgeDocumentReindexService builds a shadow index generation from the
// retained original. The old generation remains readable until the new one is
// observed as indexed and switched atomically.
type KnowledgeDocumentReindexService interface {
	ReindexKnowledgeDocument(context.Context, string, string, int64, ConversationAuthority) (KnowledgeDocument, error)
}

// Storage is library-scoped, private, and independent of uploader membership.
// Runtime/workspace/library/document are trusted application facts. Deletion
// must permanently fence late puts, including after host restarts.
type KnowledgeDocumentStorageScope struct{ RuntimeID, WorkspaceID, LibraryID string }
type KnowledgeDocumentStorage interface {
	PutKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string, string, int64, io.Reader) (string, error)
	OpenKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string, string, string, int64) (io.ReadSeekCloser, error)
	DeleteKnowledgeDocumentContent(context.Context, KnowledgeDocumentStorageScope, string) error
}

// KnowledgeDocumentChunkStorage durably stages deterministic parts and
// publishes the immutable original only after all parts and the whole-file
// digest match. An empty reference means the upload is still incomplete.
type KnowledgeDocumentChunkStorage interface {
	PutKnowledgeDocumentContentChunk(context.Context, KnowledgeDocumentStorageScope, string, string, int64, int64, string, int64, io.Reader) (nextOffset int64, reference string, err error)
}

// Provider projection includes only the explicitly mapped source fields of one
// document passage. It must exclude unrelated/global response data. The Agent
// applies its local document allowlist before constructing model evidence.
type KnowledgeDocumentPassage struct {
	DocumentID      string                 `json:"doc_id"`
	PageID          string                 `json:"page_id,omitempty"`
	PageVersion     int64                  `json:"page_version,omitempty"`
	ChunkID         string                 `json:"chunk_id,omitempty"`
	DocumentVersion int64                  `json:"document_version,omitempty"`
	ContentSHA256   string                 `json:"content_sha256,omitempty"`
	SourceRevision  string                 `json:"source_revision,omitempty"`
	Score           float64                `json:"score,omitempty"`
	Title           string                 `json:"title,omitempty"`
	URL             string                 `json:"url,omitempty"`
	Content         string                 `json:"content,omitempty"`
	Location        *DocumentLocation      `json:"location,omitempty"`
	Cells           []KnowledgePassageCell `json:"cells,omitempty"`
}

// KnowledgeDocumentSearch preserves the actual retrieval strategy alongside
// the mapped passages. Managed hosts use it for user-visible diagnostics and
// durable result receipts; callers must not infer these values from passages.
type KnowledgeDocumentSearch struct {
	Passages       []KnowledgeDocumentPassage `json:"passages"`
	TopK           int                        `json:"top_k"`
	Mode           string                     `json:"mode,omitempty"`
	Reranked       *bool                      `json:"reranked,omitempty"`
	DegradedReason string                     `json:"degraded_reason,omitempty"`
}

type ManagedKnowledgeDocumentSearchSource interface {
	SearchKnowledgeDocuments(context.Context, string, ConversationAuthority) (KnowledgeDocumentSearch, error)
}

// ManagedKnowledgeDocumentPermissionSource executes retrieval with additional
// server-owned page permission tokens. The implementation must batch without
// truncation when the current user's token set exceeds one upstream request,
// then globally merge and rank the complete result set.
type ManagedKnowledgeDocumentPermissionSource interface {
	SearchKnowledgeDocumentsWithPermissions(context.Context, string, []string, ConversationAuthority) (KnowledgeDocumentSearch, error)
	ReadKnowledgeDocumentPassagesWithPermissions(context.Context, string, []string, ConversationAuthority) ([]KnowledgeDocumentPassage, error)
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
