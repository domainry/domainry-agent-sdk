package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type ConversationHistoryRepository interface {
	SearchHistory(context.Context, agentsdk.ConversationHistorySearch, agentsdk.ConversationAuthority) (agentsdk.ConversationHistorySearchResult, error)
	HistoryMessage(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationMessage, error)
}
