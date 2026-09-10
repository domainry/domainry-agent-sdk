package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Optional local personal effects. Implementations read arguments from the
// frozen call ledger and atomically commit the effect, result and event. The
// same ledger is the idempotency receipt, including after interruption.
type ConversationPersonalMutationRepository interface {
	ApplyPersonalTool(context.Context, agentsdk.ConversationToolRequest) (agentsdk.ConversationToolResult, error)
}
