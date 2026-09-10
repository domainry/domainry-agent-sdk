package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// These types are internal host inputs, never browser/model request envelopes.
// References and permission IDs are chosen and verified by the trusted host.
type ConversationAttachmentRecord struct {
	Attachment    agentsdk.ConversationAttachment `json:"attachment"`
	RequestSHA256 string                          `json:"request_sha256"`
	BodyRef       string                          `json:"body_ref,omitempty"`
	Source        *ConversationAttachmentSource   `json:"source,omitempty"`
}

type ConversationAttachmentSource struct {
	Identity     string `json:"identity"`
	DocID        string `json:"doc_id"`
	PermissionID string `json:"permission_id"`
}

// Reserve is idempotent by owner, conversation and ClientID. The same key must
// bind identical filename/type/bytes/hash. Replays return current state, never
// resurrecting a deleted attachment or resetting its indexing work.
type ConversationAttachmentReserve struct {
	ClientID       string
	ConversationID string
	Filename       string
	ContentType    string
	SHA256         string
	Bytes          int64
}

// Revision fences stale upload/index workers. A stored upload is not ready.
// Source is attached before the first remote attempt and stays immutable;
// ready requires positive indexing and private-ACL verification by the host.
type ConversationAttachmentTransition struct {
	State     string
	BodyRef   string
	Source    *ConversationAttachmentSource
	ErrorCode string
}

type ConversationAttachmentRepository interface {
	ReserveAttachment(context.Context, ConversationAttachmentReserve, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	AttachmentRecord(context.Context, string, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	Attachments(context.Context, string, string, int, agentsdk.ConversationAuthority) (agentsdk.ConversationAttachmentPage, error)
	TransitionAttachment(context.Context, string, int64, ConversationAttachmentTransition, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	// Runtime-scoped host maintenance, never exposed through a user RPC.
	AttachmentCleanupCandidates(context.Context, string, time.Time, int) ([]ConversationAttachmentCleanup, error)
	DeferAttachmentCleanup(context.Context, string, time.Time, agentsdk.ConversationAuthority) error
	// Tombstones remain readable through the trusted record interface after
	// deleting a conversation so cleanup can resume without its parent row.
}

// Enqueued atomically when access is revoked. This authority identifies the
// original private storage namespace; it does not authorize new user actions.
type ConversationAttachmentCleanup struct {
	AttachmentID string                         `json:"attachment_id"`
	Authority    agentsdk.ConversationAuthority `json:"authority"`
}
