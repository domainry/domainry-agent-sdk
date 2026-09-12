package agentsdk

import "context"

// A trusted host owns connections, credentials and the allowed catalog. The
// application only resolves catalog keys in a known library storage scope.
// None of these definition/source interfaces are model or browser inputs.
type KnowledgeDatasourceDefinition struct {
	Key, Name, Description, SourceID string
}
type KnowledgeDatasourceSource interface {
	ConversationKnowledgeSource
	ManagedKnowledgeDocumentSource
	KnowledgeDocumentAccessPolicySource
	KnowledgeDocumentSizeLimitSource
}
type KnowledgeDatasourceCatalog interface {
	KnowledgeDatasources(context.Context, KnowledgeDocumentStorageScope) ([]KnowledgeDatasourceDefinition, error)
	OpenKnowledgeDatasource(context.Context, string, KnowledgeDocumentStorageScope) (KnowledgeDatasourceSource, error)
}

type KnowledgeDatasourceChoice struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Available   bool   `json:"available"`
}
type KnowledgeLibrarySources struct {
	LibraryID  string                      `json:"library_id"`
	Revision   int64                       `json:"revision"`
	CurrentKey string                      `json:"current_key,omitempty"`
	Status     string                      `json:"status"` // unbound, connected, unavailable, host_managed
	Items      []KnowledgeDatasourceChoice `json:"items"`
	NextAfter  string                      `json:"next_after,omitempty"`
	Complete   bool                        `json:"complete"`
	CanBind    bool                        `json:"can_bind"`
}
type KnowledgeLibrarySourceWrite struct {
	DatasourceKey    string `json:"datasource_key"`
	ExpectedRevision int64  `json:"expected_revision"`
}
type KnowledgeDatasourceService interface {
	KnowledgeLibrarySources(context.Context, string, string, int, ConversationAuthority) (KnowledgeLibrarySources, error)
	BindKnowledgeLibrarySource(context.Context, string, KnowledgeLibrarySourceWrite, ConversationAuthority) (KnowledgeLibrary, error)
}
