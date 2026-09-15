package agentsdk

import (
	"context"
	"time"
)

type ConversationContractPublicationRequest struct {
	AgreementRevision int64 `json:"agreement_revision"`
}

type ConversationContractPublication struct {
	AgreementRevision int64  `json:"agreement_revision"`
	RecordDigest      string `json:"record_digest"`
}

type ConversationContractPublicationPreview struct {
	Agreement        ConversationAgreementRevision `json:"agreement"`
	Requirements     ConversationAgentRequirements `json:"requirements"`
	Sources          []ConversationRunReference    `json:"sources"`
	RecordDigest     string                        `json:"record_digest"`
	ExpectedRevision int64                         `json:"expected_revision"`
	Publisher        ConversationAuthority         `json:"publisher"`
	RecipientUserID  string                        `json:"recipient_user_id"`
}

type ConversationContractPublicationCandidate struct {
	Revision     int64 `json:"revision"`
	BriefVersion int64 `json:"brief_version"`
}

type ConversationContractPublicationCandidates struct {
	Items                    []ConversationContractPublicationCandidate `json:"items"`
	CurrentAgreementRevision int64                                      `json:"current_agreement_revision"`
	Complete                 bool                                       `json:"complete"`
	NextBefore               int64                                      `json:"next_before,omitempty"`
}

type ConversationContractPublicationReceipt struct {
	ConversationContractPublication
	Revision        int64                 `json:"revision"`
	Publisher       ConversationAuthority `json:"publisher"`
	RecipientUserID string                `json:"recipient_user_id"`
	PublishedAt     time.Time             `json:"published_at"`
	Reason          string                `json:"reason"`
}

type ConversationContractPublicationHistory struct {
	Items      []ConversationContractPublicationReceipt `json:"items"`
	Complete   bool                                     `json:"complete"`
	NextBefore int64                                    `json:"next_before,omitempty"`
}

// Preparation never changes an agreement or grants access. A deliberate commit
// binds this exact server-owned snapshot and the relationship's CAS revision.
type ConversationContractPublicationReader interface {
	ConversationContractPublicationCandidates(context.Context, string, int64, ConversationAuthority) (ConversationContractPublicationCandidates, error)
	PreviewConversationContractPublication(context.Context, string, ConversationContractPublicationRequest, ConversationAuthority) (ConversationContractPublicationPreview, error)
	ConversationContractPublicationHistory(context.Context, string, int64, ConversationAuthority) (ConversationContractPublicationHistory, error)
}
