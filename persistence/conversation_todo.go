package persistence

import agentsdk "github.com/domainry/domainry-agent-sdk"

// Direct user mutations persist a client-key receipt with the effect. Model
// tools share the same mutation rules through ApplyPersonalTool, whose receipt
// is the existing execution ledger. All operations are owner-scoped.
type ConversationTodoRepository interface {
	agentsdk.ConversationTodoService
}
