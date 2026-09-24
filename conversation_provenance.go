package agentsdk

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const ConversationProvenanceContentMaxBytes = 1 << 20

// ConversationProvenancePublication imports one already-completed turn from a
// trusted local Agent runtime. Agent remains the durable Conversation/Run
// owner; the publishing runtime supplies content and idempotency identities,
// never server Conversation or Run IDs.
type ConversationProvenancePublication struct {
	ClientID             string `json:"client_id"`
	ConversationClientID string `json:"conversation_client_id"`
	ConversationTitle    string `json:"conversation_title,omitempty"`
	UserMessage          string `json:"user_message"`
	AssistantMessage     string `json:"assistant_message"`
}

func (publication ConversationProvenancePublication) Validate() error {
	if !validConversationProvenanceKey(publication.ClientID) || !validConversationProvenanceKey(publication.ConversationClientID) {
		return fmt.Errorf("conversation provenance client identities are invalid")
	}
	if !validConversationProvenanceText(publication.ConversationTitle, 512, false) ||
		!validConversationProvenanceText(publication.UserMessage, ConversationProvenanceContentMaxBytes, true) ||
		!validConversationProvenanceText(publication.AssistantMessage, ConversationProvenanceContentMaxBytes, true) {
		return fmt.Errorf("conversation provenance content is invalid")
	}
	return nil
}

type ConversationProvenanceReceipt struct {
	ConversationID string    `json:"conversation_id"`
	RunID          string    `json:"run_id"`
	SourceID       string    `json:"source_id"`
	PublishedAt    time.Time `json:"published_at"`
}

type ConversationProvenancePublisher interface {
	PublishConversationProvenance(context.Context, ConversationProvenancePublication, ConversationAuthority) (ConversationProvenanceReceipt, error)
}

func validConversationProvenanceKey(value string) bool {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 1 || len(trimmed) > 96 || value != trimmed {
		return false
	}
	value = trimmed
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-' || character == '.' || character == ':') {
			return false
		}
	}
	return true
}

func validConversationProvenanceText(value string, maximum int, required bool) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && len(value) <= maximum && (!required || strings.TrimSpace(value) != "")
}
