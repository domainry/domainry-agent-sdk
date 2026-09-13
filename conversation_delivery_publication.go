package agentsdk

import (
	"context"
	"time"
)

// Zero selects the current delivery. A positive revision selects one immutable
// delivery/assessment record, without replacing today's delivery or agreement.
type ConversationDeliveryPublicationRequest struct {
	DeliveryRevision int64 `json:"delivery_revision"`
}

type ConversationDeliveryPublication struct {
	DeliveryRevision int64  `json:"delivery_revision"`
	RecordDigest     string `json:"record_digest"`
}

type ConversationDeliveryPublicationPreview struct {
	Record           ConversationDeliveryRecord `json:"record"`
	RecordDigest     string                     `json:"record_digest"`
	ExpectedRevision int64                      `json:"expected_revision"`
	Publisher        ConversationAuthority      `json:"publisher"`
	RecipientUserID  string                     `json:"recipient_user_id"`
}

// A preparation index contains versions only. It never reveals the old
// delivery text, assessment basis, source references or execution details.
type ConversationDeliveryPublicationCandidate struct {
	Revision          int64  `json:"revision"`
	Kind              string `json:"kind"`
	BriefVersion      int64  `json:"brief_version"`
	AgreementRevision int64  `json:"agreement_revision"`
}

type ConversationDeliveryPublicationCandidates struct {
	Items            []ConversationDeliveryPublicationCandidate `json:"items"`
	CurrentAvailable bool                                       `json:"current_available"`
	Complete         bool                                       `json:"complete"`
	NextBefore       int64                                      `json:"next_before,omitempty"`
}

type ConversationDeliveryPublicationReceipt struct {
	ConversationDeliveryPublication
	Publisher       ConversationAuthority `json:"publisher"`
	RecipientUserID string                `json:"recipient_user_id"`
	PublishedAt     time.Time             `json:"published_at"`
}

// Optional preparation port. Preparation validates current source access but
// creates no grant; commit uses republish_delivery with this exact record digest.
type ConversationDeliveryPublicationReader interface {
	ConversationDeliveryPublicationCandidates(context.Context, string, int64, ConversationAuthority) (ConversationDeliveryPublicationCandidates, error)
	PreviewConversationDeliveryPublication(context.Context, string, ConversationDeliveryPublicationRequest, ConversationAuthority) (ConversationDeliveryPublicationPreview, error)
}
