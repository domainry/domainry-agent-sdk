package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type ConversationInteractionRecord struct {
	Interaction agentsdk.ConversationInteraction          `json:"interaction"`
	Response    *agentsdk.ConversationInteractionResponse `json:"response,omitempty"`
}

type ConversationWait struct {
	Step     int
	CallID   string
	Kind     string
	Question string
	Choices  []string
	TTL      time.Duration
}

type ConversationInteractionRepository interface {
	ExecutionInteraction(context.Context, ConversationClaim, int, string, string) (ConversationInteractionRecord, bool, error)
	WaitExecution(context.Context, ConversationClaim, ConversationWait) (agentsdk.ConversationInteraction, error)
	RespondExecution(context.Context, string, string, agentsdk.ConversationInteractionResponse, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
	ExpireInteractions(context.Context, string, int) (int, error)
}
