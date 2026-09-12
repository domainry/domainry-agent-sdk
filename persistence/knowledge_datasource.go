package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
	"time"
)

// The binding is durable and immutable. Disabling/removing host configuration
// does not free its physical KB for a different library. No credentials here.
type KnowledgeDatasourceBinding struct {
	LibraryID          string    `json:"library_id"`
	DatasourceKey      string    `json:"datasource_key"`
	SourceID           string    `json:"source_id"`
	AccessPolicySHA256 string    `json:"access_policy_sha256"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}
type KnowledgeDatasourceAssignment struct {
	DatasourceKey, SourceID, AccessPolicySHA256 string
	ExpectedRevision                            int64
}
type KnowledgeDatasourceRepository interface {
	// Trusted internal resolution also serves previously authorized maintenance
	// after a member leaves. User-facing access must be checked separately.
	KnowledgeDatasourceBinding(context.Context, agentsdk.KnowledgeDocumentStorageScope) (KnowledgeDatasourceBinding, bool, error)
	// Checks live manager membership and expected library revision; commits the
	// binding and exclusive physical-source marker in the same transaction.
	BindKnowledgeDatasource(context.Context, string, KnowledgeDatasourceAssignment, agentsdk.ConversationAuthority) (agentsdk.KnowledgeLibrary, error)
}
