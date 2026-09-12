package agentsdk

import (
	"context"
	toolsdk "github.com/domainry/domainry-tools-sdk"
	"time"
)

// A question/confirmation belongs to one frozen logical call, never to an
// ambient browser session. Parameters shown to the user are the parameters
// that resume will execute; a new payload requires a new approval.
type ConversationInteraction struct {
	// Operations is the server-built list of concrete remaining writes in the
	// frozen step. Only an explicit listed_operations response approves it.
	Operations      []ConversationInteractionOperation `json:"operations,omitempty"`
	AuthorizationID string                             `json:"authorization_id,omitempty"` // original grouped approval, if any
	ApprovedScope   string                             `json:"approved_scope,omitempty"`   // empty: this call only; listed_operations: the displayed list
	ID              string                             `json:"id"`
	ConversationID  string                             `json:"conversation_id"`
	RunID           string                             `json:"run_id"`
	Step            int                                `json:"step"`
	CallID          string                             `json:"call_id"`
	Kind            string                             `json:"kind"`   // input, confirmation or reconciliation
	Status          string                             `json:"status"` // pending, answered, approved, rejected, cancelled, expired, resolved
	Question        string                             `json:"question"`
	Choices         []string                           `json:"choices,omitempty"`
	Tool            string                             `json:"tool"`
	ToolVersion     string                             `json:"tool_version"`
	ActionKey       string                             `json:"action_key"`
	Arguments       string                             `json:"arguments"`
	ArgumentsHash   string                             `json:"arguments_hash"`
	DefinitionHash  string                             `json:"definition_hash"`
	Revision        int64                              `json:"revision"`
	Answer          string                             `json:"answer,omitempty"`
	RespondedBy     string                             `json:"responded_by,omitempty"`
	CreatedAt       time.Time                          `json:"created_at"`
	ExpiresAt       time.Time                          `json:"expires_at"`
	RespondedAt     *time.Time                         `json:"responded_at,omitempty"`
}

// A listed operation is one logical invocation, not a wildcard permission for
// every future use of the tool. The normal call ledger provides at-most-once
// identity across retries; changed calls need their own user approval.
type ConversationInteractionOperation struct {
	CallID         string `json:"call_id"`
	Tool           string `json:"tool"`
	ToolVersion    string `json:"tool_version"`
	ActionKey      string `json:"action_key"`
	Arguments      string `json:"arguments"`
	ArgumentsHash  string `json:"arguments_hash"`
	DefinitionHash string `json:"definition_hash"`
}

type ConversationInteractionResponse struct {
	Scope            string `json:"scope,omitempty"` // empty: current call; listed_operations: explicit concrete list
	InteractionID    string `json:"interaction_id"`
	ClientID         string `json:"client_id"`
	ExpectedRevision int64  `json:"expected_revision"`
	Decision         string `json:"decision"` // answer, approve or reject
	Answer           string `json:"answer,omitempty"`
}

// Optional extension so existing pure-text service implementations stay valid.
type ConversationInteractionService interface {
	Respond(context.Context, string, string, ConversationInteractionResponse, ConversationAuthority) (ConversationRun, error)
}

type ConversationInteractionAuthorizer interface {
	AuthorizeConversationInteraction(context.Context, ConversationAuthority, ConversationInteraction) (ConversationToolAuthorization, error)
}

// This is a trusted server-side receipt, never accepted as a browser or model
// input. The executor loads it from an owner-scoped persisted interaction.
type ConversationConfirmation = toolsdk.Confirmation

type ConversationConfirmationVerifier = toolsdk.ConfirmationVerifier
