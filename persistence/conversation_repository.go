package persistence

import (
	"context"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type ConversationClaim struct {
	Authority agentsdk.ConversationAuthority
	Run       agentsdk.ConversationRun
	Owner     string
	Fence     int64
	ExpiresAt time.Time
}

// Repository methods are atomic owner operations, not a remote mutation bridge
// into Runtime business transactions. All rows include the runtime/workspace/user scope.
type ConversationRepository interface {
	Ready(context.Context) error
	Create(context.Context, agentsdk.ConversationCreate, agentsdk.ConversationAuthority) (agentsdk.Conversation, error)
	List(context.Context, agentsdk.ConversationQuery, agentsdk.ConversationAuthority) (agentsdk.ConversationPage, error)
	Get(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.Conversation, error)
	Update(context.Context, string, agentsdk.ConversationUpdate, agentsdk.ConversationAuthority) (agentsdk.Conversation, error)
	Delete(context.Context, string, int64, agentsdk.ConversationAuthority) error
	Enqueue(context.Context, string, agentsdk.ConversationSend, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
	Messages(context.Context, string, agentsdk.ConversationMessageQuery, agentsdk.ConversationAuthority) (agentsdk.ConversationMessagePage, error)
	History(context.Context, string, int64, int64, int, agentsdk.ConversationAuthority) ([]agentsdk.ConversationMessage, error)
	Run(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
	Events(context.Context, string, string, int64, int, agentsdk.ConversationAuthority) (agentsdk.ConversationEventPage, error)
	Cancel(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
	Resume(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationRun, error)
	Claim(context.Context, string, string, time.Duration) (ConversationClaim, bool, error)
	Heartbeat(context.Context, ConversationClaim, time.Duration) (bool, error)
	Finish(context.Context, ConversationClaim, agentsdk.ConversationModelResult, string) error
	AppendEvent(context.Context, ConversationClaim, string, map[string]any) error
	// Offset is a UTF-8 byte offset within the current attempt. A retry of an
	// already committed identical delta is idempotent; another attempt is fenced.
	AppendDelta(context.Context, ConversationClaim, int, string) error
	Summary(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationSummary, error)
	SaveSummary(context.Context, ConversationClaim, agentsdk.ConversationSummary) error
	ModelInput(context.Context, ConversationClaim, *agentsdk.ConversationModelRequest) (agentsdk.ConversationModelRequest, bool, error)
	Memories(context.Context, agentsdk.ConversationAuthority) ([]agentsdk.ConversationMemory, error)
	WriteMemory(context.Context, agentsdk.ConversationMemoryWrite, agentsdk.ConversationAuthority) (agentsdk.ConversationMemory, error)
	DeleteMemory(context.Context, string, int64, agentsdk.ConversationAuthority) error
}
