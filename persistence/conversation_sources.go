package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// A consistent owner-scoped snapshot for source access checks, without a
// worker lease. Never expose the input or ledger directly to a browser.
type ConversationSourceRepository interface {
	ConversationSourceSnapshot(context.Context, agentsdk.ConversationRunReference, agentsdk.ConversationAuthority) (ConversationSourceSnapshot, error)
}

type ConversationSourceSnapshot struct {
	Run   agentsdk.ConversationRun
	Input *agentsdk.ConversationModelRequest
	Calls []ConversationToolExecution
}
