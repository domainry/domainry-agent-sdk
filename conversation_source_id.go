package agentsdk

import (
	"fmt"
	"strings"
)

// ConversationRunSourceID is the canonical immutable identity of one Agent
// conversation turn. Other owners may persist this opaque identity without
// reading or depending on the Agent service.
func ConversationRunSourceID(reference ConversationRunReference) (string, error) {
	conversationID := strings.TrimSpace(reference.ConversationID)
	runID := strings.TrimSpace(reference.RunID)
	if !validConversationSourceComponent(conversationID) || !validConversationSourceComponent(runID) {
		return "", fmt.Errorf("conversation source requires safe conversation and run identities")
	}
	return "conversation://" + conversationID + "/turn/" + runID, nil
}

// ParseConversationRunSourceID accepts only the canonical form emitted by
// ConversationRunSourceID. It validates identity syntax, not owner access.
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
