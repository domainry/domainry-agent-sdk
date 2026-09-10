package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Optional owner-scoped ledger reads; no worker lease is needed. Application
// code must reauthorize each original tool before exposing any call metadata.
type ConversationExecutionReadRepository interface {
	ReadExecutionCalls(context.Context, agentsdk.ConversationExecutionRead, agentsdk.ConversationAuthority) (ConversationExecutionCallPage, error)
	ReadExecutionCall(context.Context, string, string, int, string, agentsdk.ConversationAuthority) (ConversationToolExecution, error)
}

type ConversationExecutionCallPage struct {
	RunStatus  string
	Items      []ConversationToolExecution
	NextCursor string
	Complete   bool
}
