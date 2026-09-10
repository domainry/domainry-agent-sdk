package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// Optional, read-only lookup for immutable completed results, including older
// runs. The repository checks owner and digest; the application checks live
// source-tool authorization before data enters any new model request.
type ConversationResultRepository interface {
	ConversationResult(context.Context, agentsdk.ConversationResultReference, agentsdk.ConversationAuthority) (ConversationToolExecution, error)
}
