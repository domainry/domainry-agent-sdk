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
	References  []ConversationRunReference `json:"references"`
	SourceIDs   []string                   `json:"source_ids"`
	DecisionIDs []string                   `json:"decision_ids"`
	Reader      ConversationAuthority      `json:"reader"`
}

// ConversationSourceVerificationReceipt contains only immutable identities.
// It deliberately contains no message, attachment, artifact, or result body.
type ConversationSourceVerificationReceipt struct {
	WorkspaceID string                     `json:"workspace_id"`
	References  []ConversationRunReference `json:"references"`
	SourceIDs   []string                   `json:"source_ids"`
	DecisionIDs []string                   `json:"decision_ids"`
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
	for _, values := range [][]string{request.SourceIDs, request.DecisionIDs} {
		seen = map[string]bool{}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" || seen[value] {
				return fmt.Errorf("source verification contains an empty or duplicate identity")
			}
			seen[value] = true
		}
	}
	return nil
}

// ConversationSourceVerifier is a read-only owner port. Implementations must
// fail closed unless every conversation/run exists in Reader's workspace,
// Reader can currently read it, and every BeforeStep is within the owned run.
type ConversationSourceVerifier interface {
	VerifyConversationSources(context.Context, ConversationSourceVerificationRequest) (ConversationSourceVerificationReceipt, error)
}
