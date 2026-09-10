package agentsdk

import "context"

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

type KnowledgeDocumentContent struct {
	DocumentID string
	Filename   string
	Data       []byte
}

type KnowledgeDocumentState struct {
	DocumentID string `json:"document_id"`
	// Exists is visibility in the current upstream permission scope. A hidden
	// document can be reported as missing; false alone is not global deletion proof.
	Exists      bool   `json:"exists"`
	IndexStatus string `json:"index_status,omitempty"` // Actual upstream value; unknown is not ready.
}
