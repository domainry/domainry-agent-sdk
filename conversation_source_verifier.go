package agentsdk

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ConversationSourceVerificationRequest is the narrow owner contract used by
// modules that freeze Agent-owned provenance. The caller supplies the exact
// reader authority and cannot replace it with a model-authored identity.
type ConversationSourceVerificationRequest struct {
	References []ConversationRunReference `json:"references"`
	SourceIDs  []string                   `json:"source_ids"`
	Reader     ConversationAuthority      `json:"reader"`
}

// ConversationSourceVerificationReceipt contains only immutable identities.
// It deliberately contains no message, attachment, artifact, or result body.
type ConversationSourceVerificationReceipt struct {
	WorkspaceID string                     `json:"workspace_id"`
	References  []ConversationRunReference `json:"references"`
	SourceIDs   []string                   `json:"source_ids"`
	VerifiedAt  time.Time                  `json:"verified_at"`
}

func (request ConversationSourceVerificationRequest) Validate() error {
	if !request.Reader.Known || strings.TrimSpace(request.Reader.RuntimeID) == "" || strings.TrimSpace(request.Reader.WorkspaceID) == "" || strings.TrimSpace(request.Reader.UserID) == "" {
		return fmt.Errorf("source verification requires an authenticated reader authority")
	}
	if len(request.References) == 0 || len(request.SourceIDs) == 0 {
		return fmt.Errorf("source verification requires run references and source identities")
	}
	seen := map[string]bool{}
	for _, reference := range request.References {
		key := strings.TrimSpace(reference.ConversationID) + "\x00" + strings.TrimSpace(reference.RunID)
		if strings.TrimSpace(reference.ConversationID) == "" || strings.TrimSpace(reference.RunID) == "" || reference.BeforeStep < 0 || seen[key] {
			return fmt.Errorf("source verification contains an invalid or duplicate run reference")
		}
		seen[key] = true
	}
	seen = map[string]bool{}
	for _, value := range request.SourceIDs {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return fmt.Errorf("source verification contains an empty or duplicate identity")
		}
		if _, err := ParseConversationRunSourceID(value); err != nil {
			return err
		}
		seen[value] = true
	}
	return nil
}

// ConversationRunSourceID is the canonical immutable identity of one Agent
// conversation turn. Callers may store the string, but only Agent can prove
// that the referenced run exists and is readable in the current Workspace.
func ConversationRunSourceID(reference ConversationRunReference) (string, error) {
	conversationID := strings.TrimSpace(reference.ConversationID)
	runID := strings.TrimSpace(reference.RunID)
	if !validConversationSourceComponent(conversationID) || !validConversationSourceComponent(runID) {
		return "", fmt.Errorf("conversation source requires safe conversation and run identities")
	}
	return "conversation://" + conversationID + "/turn/" + runID, nil
}

// ParseConversationRunSourceID accepts only the canonical form emitted by
// ConversationRunSourceID. It does not establish ownership or access; that is
// the ConversationSourceVerifier implementation's responsibility.
func ParseConversationRunSourceID(value string) (ConversationRunReference, error) {
	value = strings.TrimSpace(value)
	const prefix = "conversation://"
	if !strings.HasPrefix(value, prefix) {
		return ConversationRunReference{}, fmt.Errorf("source identity is not an Agent conversation turn")
	}
	conversationID, runID, found := strings.Cut(strings.TrimPrefix(value, prefix), "/turn/")
	if !found || !validConversationSourceComponent(conversationID) || !validConversationSourceComponent(runID) {
		return ConversationRunReference{}, fmt.Errorf("conversation source identity is invalid")
	}
	reference := ConversationRunReference{ConversationID: conversationID, RunID: runID}
	canonical, err := ConversationRunSourceID(reference)
	if err != nil || canonical != value {
		return ConversationRunReference{}, fmt.Errorf("conversation source identity is not canonical")
	}
	return reference, nil
}

func validConversationSourceComponent(value string) bool {
	if value == "" || len(value) > 255 {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || strings.ContainsRune("_.:-", character)) {
			return false
		}
	}
	return true
}

// ConversationSourceVerifier is a read-only owner port. Implementations must
// fail closed unless every conversation/run exists in Reader's workspace,
// Reader can currently read it, and every BeforeStep is within the owned run.
type ConversationSourceVerifier interface {
	VerifyConversationSources(context.Context, ConversationSourceVerificationRequest) (ConversationSourceVerificationReceipt, error)
}
