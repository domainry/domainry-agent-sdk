package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Side-effect markers are persisted before Connector writes. They survive
// source/parent deletion, lease takeover and response loss. Missing visibility
// alone cannot prove a private remote document was deleted.
type ConversationAttachmentIndex struct {
	LastCheckRevision  int64                          `json:"last_check_revision,omitempty"`
	Actor              agentsdk.ConversationAuthority `json:"actor"`
	PutStarted         bool                           `json:"put_started"`
	PutAcknowledged    bool                           `json:"put_acknowledged"`
	IndexObserved      bool                           `json:"index_observed"`
	DeleteStarted      bool                           `json:"delete_started"`
	DeleteAcknowledged bool                           `json:"delete_acknowledged"`
}

// A check is idempotent for its expected revision and preserves active leases,
// immutable source references and every remote write marker.
type ConversationAttachmentIndexCheckRepository interface {
	RequestAttachmentIndexCheck(context.Context, string, string, int64, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
}

type ConversationAttachmentIndexLease struct {
	Authority      agentsdk.ConversationAuthority `json:"authority"`
	ConversationID string                         `json:"conversation_id"`
	AttachmentID   string                         `json:"attachment_id"`
	Owner          string                         `json:"owner"`
	Token          int64                          `json:"token"`
	ExpiresAt      time.Time                      `json:"expires_at"`
}

type ConversationAttachmentIndexProgress struct {
	Event, IndexStatus, ErrorCode string
	RetryAt                       time.Time
}

// The source registry is runtime/workspace-scoped maintenance. It must reject
// KBs owned by any library/default path; the guard remains after configuration
// removal or the last attachment is deleted. Other methods fence one private
// attachment and always preserve its owner and originating conversation.
type ConversationAttachmentIndexRepository interface {
	KnowledgeSourceRegistry
	AttachmentContent(context.Context, string, agentsdk.ConversationAuthority) ([]byte, error)
	DeleteAttachmentContent(context.Context, string, agentsdk.ConversationAuthority) error
	ActivateAttachmentKnowledgeSource(context.Context, string, string, string) error
	QueueAttachmentIndex(context.Context, string, int64, ConversationAttachmentSource, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	ClaimAttachmentIndexWork(context.Context, string, string, time.Time, time.Duration) (ConversationAttachmentIndexLease, bool, error)
	AttachmentIndexWorkRecord(context.Context, ConversationAttachmentIndexLease) (ConversationAttachmentRecord, error)
	StartAttachmentIndexPut(context.Context, ConversationAttachmentIndexLease) (ConversationAttachmentRecord, bool, error)
	StartAttachmentIndexDelete(context.Context, ConversationAttachmentIndexLease) error
	ApplyAttachmentIndexProgress(context.Context, ConversationAttachmentIndexLease, ConversationAttachmentIndexProgress) error
}
