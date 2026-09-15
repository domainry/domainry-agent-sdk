package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// ConversationForkSeed is private model context copied from one completed run.
// The source is provenance and a current-access dependency, never authority.
type ConversationForkSeed struct {
	Version          int                                 `json:"version"`
	Source           agentsdk.ConversationRunReference   `json:"source"`
	BoundaryEventSeq int64                               `json:"boundary_event_seq"`
	SourceSHA256     string                              `json:"source_sha256"`
	Messages         []agentsdk.ConversationModelMessage `json:"messages"`
	CreatedAt        time.Time                           `json:"created_at"`
}

// ConversationForkRepository atomically creates an independent conversation
// and its private seed, and reads that seed only in the owning principal scope.
type ConversationForkRepository interface {
	ForkConversation(context.Context, agentsdk.ConversationForkRequest, ConversationForkSeed, agentsdk.ConversationAuthority) (agentsdk.Conversation, error)
	ConversationForkSeed(context.Context, string, agentsdk.ConversationAuthority) (ConversationForkSeed, bool, error)
}
