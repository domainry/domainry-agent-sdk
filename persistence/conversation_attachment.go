package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// These types are internal host inputs, never browser/model request envelopes.
// References and permission IDs are chosen and verified by the trusted host.
type ConversationAttachmentRecord struct {
	Attachment    agentsdk.ConversationAttachment `json:"attachment"`
	RequestSHA256 string                          `json:"request_sha256"`
	Source        *ConversationAttachmentSource   `json:"source,omitempty"`
	Index         *ConversationAttachmentIndex    `json:"index,omitempty"`
}

type ConversationAttachmentSource struct {
	Identity           string `json:"identity"`
	DocID              string `json:"doc_id"`
	PermissionID       string `json:"permission_id"`
	AccessPolicySHA256 string `json:"access_policy_sha256,omitempty"`
	RequestID          string `json:"request_id,omitempty"`
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
	// Content is trusted application input. It is written through the
	// deployment Artifact ContentWriter and is never serialized into SQL.
	Content []byte `json:"-"`
}

// Revision fences stale upload/index workers. A stored upload is not ready.
// Source is attached before the first remote attempt and stays immutable;
// ready requires positive indexing and private-ACL verification by the host.
type ConversationAttachmentTransition struct {
	State     string
	ErrorCode string
}

type ConversationAttachmentRepository interface {
	ReserveAttachment(context.Context, ConversationAttachmentReserve, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	AttachmentRecord(context.Context, string, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	Attachments(context.Context, string, string, int, agentsdk.ConversationAuthority) (agentsdk.ConversationAttachmentPage, error)
	TransitionAttachment(context.Context, string, int64, ConversationAttachmentTransition, agentsdk.ConversationAuthority) (ConversationAttachmentRecord, error)
	AttachmentContent(context.Context, string, agentsdk.ConversationAuthority) ([]byte, error)
	DeleteAttachmentContent(context.Context, string, agentsdk.ConversationAuthority) error
}
