package agentsdk

import (
	"context"
	"time"
)

// A conversation attachment is private to its authenticated owner and source
// conversation. Uploading bytes does not make them indexed or model-readable.
type ConversationAttachment struct {
	Indexing       *ConversationAttachmentIndexView `json:"indexing,omitempty"`
	ID             string                           `json:"id"`
	ConversationID string                           `json:"conversation_id"`
	Filename       string                           `json:"filename"`
	ContentType    string                           `json:"content_type"`
	Bytes          int64                            `json:"bytes"`
	SHA256         string                           `json:"sha256"`
	Visibility     string                           `json:"visibility"` // conversation_private
	State          string                           `json:"state"`      // uploading, stored, indexing, ready, failed, needs_reconcile, deleting, deleted
	IndexStatus    string                           `json:"index_status,omitempty"`
	Revision       int64                            `json:"revision"`
	ErrorCode      string                           `json:"error_code,omitempty"`
	CreatedAt      time.Time                        `json:"created_at"`
	UpdatedAt      time.Time                        `json:"updated_at"`
}

// Current host capabilities, computed on reads and responses, never an ACL
// grant or a persisted worker input. No remote identifiers are exposed.
type ConversationAttachmentIndexView struct {
	LastCheckRevision int64  `json:"last_check_revision,omitempty"`
	Requested         bool   `json:"requested"`
	CanStart          bool   `json:"can_start"`
	CanCheck          bool   `json:"can_check"`
	MaxBytes          int64  `json:"max_bytes,omitempty"`
	Reason            string `json:"reason,omitempty"`
}

type ConversationAttachmentPage struct {
	Items     []ConversationAttachment `json:"items"`
	NextAfter string                   `json:"next_after,omitempty"`
	Complete  bool                     `json:"complete"`
}

const ConversationAttachmentMaxBytes int64 = 16 << 20

// Data is transported as binary by the browser route and as base64 by the
// trusted service RPC. File paths, owner IDs and indexing permissions are not input.
type ConversationAttachmentUpload struct {
	ClientID string `json:"client_id"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

type ConversationAttachmentDownload struct {
	Attachment ConversationAttachment `json:"attachment"`
	Data       []byte                 `json:"data"`
}

type ConversationAttachmentService interface {
	UploadAttachment(context.Context, string, ConversationAttachmentUpload, ConversationAuthority) (ConversationAttachment, error)
	Attachments(context.Context, string, string, int, ConversationAuthority) (ConversationAttachmentPage, error)
	Attachment(context.Context, string, string, ConversationAuthority) (ConversationAttachment, error)
	DownloadAttachment(context.Context, string, string, ConversationAuthority) (ConversationAttachmentDownload, error)
	DeleteAttachment(context.Context, string, string, int64, ConversationAuthority) (ConversationAttachment, error)
}

// Authorization is separate from model tools: these operations are explicit
// user file management, never automatically exposed to the model's tool catalog.
type ConversationAttachmentAuthorizer interface {
	AuthorizeConversationAttachment(context.Context, string, ConversationAuthority) error
}
