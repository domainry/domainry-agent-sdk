package agentsdk

import (
	"context"
	"io"
)

// KnowledgeDocumentSource is a trusted host port for document lifecycle work,
// not a model tool or browser API. The application must first authorize the
// current library action and resolve the document's recorded source binding.
// Put/Delete only acknowledge the request; Inspect reports upstream state.
// An uncertain write is not permission to retry or declare cleanup complete.
type KnowledgeDocumentSource interface {
	PutKnowledgeDocument(context.Context, KnowledgeDocumentContent, ConversationAuthority) error
	InspectKnowledgeDocument(context.Context, string, ConversationAuthority) (KnowledgeDocumentState, error)
	DeleteKnowledgeDocument(context.Context, string, ConversationAuthority) error
}

// Optional recovery of the SAME immutable document deletion. A nil error is
// a positive deletion acknowledgement: the source must either retrieve a
// definitive operation receipt or execute a verified idempotent delete. A
// permission-scoped lookup returning "not found" is never sufficient. The
// caller must preserve the recorded source, document ID and cleanup intent.
// Sources without this port retain their uncertain state without another write.
type KnowledgeDocumentDeleteRecoverySource interface {
	RecoverKnowledgeDocumentDelete(context.Context, string, ConversationAuthority) error
}

type KnowledgeDocumentContent struct {
	DocumentID string
	Filename   string
	Bytes      int64
	SHA256     string
	Content    io.Reader
	// RequestID is created and durably recorded by the application before any
	// write. It identifies one logical command, not permission to retry it.
	RequestID          string
	AccessPolicySHA256 string
	// PermissionIDs is a trusted per-document projection override. It is used
	// for source-owned page ACL tokens and is never populated from public input.
	PermissionIDs []string
}

// A source may expose an immutable access-policy fingerprint independently of
// its physical source identity. The application persists this at reservation
// and rejects stale work/evidence when configuration changes. Empty preserves
// legacy sources that do not implement a separate private upload policy.
type KnowledgeDocumentAccessPolicySource interface {
	KnowledgeDocumentAccessPolicySHA256() string
}

// Optional upstream indexing limit. This is deliberately separate from the
// original-file storage limit: Knowledge may retain, preview and download an
// original that is larger than one provider indexing request.
type KnowledgeDocumentIndexLimitSource interface {
	KnowledgeDocumentIndexMaximumBytes() int64
}

// KnowledgeDocumentFormat describes one exact Builder parser route and its
// route-specific processing budgets. The Connector transport ceiling is a
// source-wide value exposed separately by KnowledgeDocumentIndexLimitSource.
// Asset/enrichment/analysis budgets are not whole-file upload limits.
type KnowledgeDocumentFormat struct {
	Extension              string `json:"extension"`
	ContentType            string `json:"content_type"`
	ParserRoute            string `json:"parser_route"`
	BuilderMaxBytes        int64  `json:"builder_max_bytes,omitempty"`
	AssetMaxBytes          int64  `json:"asset_max_bytes,omitempty"`
	EnrichmentMaxBytes     int64  `json:"enrichment_max_bytes,omitempty"`
	AnalysisOutputMaxBytes int64  `json:"analysis_output_max_bytes,omitempty"`
}

// KnowledgeDocumentFormatSource is the authoritative file capability surface
// supplied by the configured Connector. Knowledge must not duplicate a
// Builder format table in its application layer.
type KnowledgeDocumentFormatSource interface {
	KnowledgeDocumentFormats() []KnowledgeDocumentFormat
}

type KnowledgeDocumentState struct {
	DocumentID string `json:"document_id"`
	// Exists is visibility in the current upstream permission scope. A hidden
	// document can be reported as missing; false alone is not global deletion proof.
	Exists      bool   `json:"exists"`
	IndexStatus string `json:"index_status,omitempty"` // Actual upstream value; unknown is not ready.
}
