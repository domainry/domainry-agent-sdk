package agentsdk

import (
	"context"
	"time"
)

// Libraries belong to a trusted Runtime/Workspace scope. Role is the current
// caller's membership, not a transferable grant. Personal libraries cannot be
// given members. Creating a library does not configure a remote knowledge source.
type KnowledgeLibrary struct {
	KnowledgeConfigured bool      `json:"knowledge_configured,omitempty"` // Host binding exists; not a remote health/indexing assertion.
	DocumentsConfigured bool      `json:"documents_configured,omitempty"` // Host enables document management; not an Identity permission grant.
	ID                  string    `json:"id"`
	Kind                string    `json:"kind"` // personal, shared
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	OwnerUserID         string    `json:"owner_user_id"`
	Role                string    `json:"role"` // reader, editor, manager
	Archived            bool      `json:"archived"`
	Revision            int64     `json:"revision"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
type KnowledgeLibraryCreate struct {
	ClientID    string `json:"client_id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type KnowledgeLibraryUpdate struct {
	ExpectedRevision int64  `json:"expected_revision"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Archived         bool   `json:"archived"`
}
type KnowledgeLibraryPage struct {
	Items     []KnowledgeLibrary `json:"items"`
	NextAfter string             `json:"next_after,omitempty"`
	Complete  bool               `json:"complete"`
}
type KnowledgeLibraryMember struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	UpdatedAt time.Time `json:"updated_at"`
}
type KnowledgeLibraryMembers struct {
	Items     []KnowledgeLibraryMember `json:"items"`
	NextAfter string                   `json:"next_after,omitempty"`
	Complete  bool                     `json:"complete"`
	Revision  int64                    `json:"revision"`
}
type KnowledgeLibraryMemberWrite struct {
	Role             string `json:"role"`
	ExpectedRevision int64  `json:"expected_revision"`
}
type KnowledgeLibraryService interface {
	CreateKnowledgeLibrary(context.Context, KnowledgeLibraryCreate, ConversationAuthority) (KnowledgeLibrary, error)
	KnowledgeLibraries(context.Context, string, int, ConversationAuthority) (KnowledgeLibraryPage, error)
	KnowledgeLibrary(context.Context, string, ConversationAuthority) (KnowledgeLibrary, error)
	UpdateKnowledgeLibrary(context.Context, string, KnowledgeLibraryUpdate, ConversationAuthority) (KnowledgeLibrary, error)
	KnowledgeLibraryMembers(context.Context, string, string, int, ConversationAuthority) (KnowledgeLibraryMembers, error)
	SetKnowledgeLibraryMember(context.Context, string, string, KnowledgeLibraryMemberWrite, ConversationAuthority) (KnowledgeLibrary, error)
	RemoveKnowledgeLibraryMember(context.Context, string, string, int64, ConversationAuthority) (KnowledgeLibrary, error)
}

// The host resolves live Identity on every operation, evaluating facts loaded
// by the application. Member validation accepts only active users of this
// workspace. It must not create accounts or grant global Identity actions.
type KnowledgeLibraryAuthorizer interface {
	AuthorizeKnowledgeLibrary(context.Context, string, KnowledgeLibrary, ConversationAuthority) error
	ValidateKnowledgeLibraryMember(context.Context, string, ConversationAuthority) error
}
