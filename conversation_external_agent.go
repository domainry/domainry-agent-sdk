package agentsdk

import (
	"context"
	"encoding/json"
	"time"
)

const (
	CapabilityConversationExternalAgentV1 = "conversation.external_agent.v1"
	ConversationExternalAgentProtocolV1   = "domainry-peer"
)

// ConversationExternalAgentCapabilities is negotiated explicitly for an
// Agent implemented by another process or service. It describes protocol
// behavior only; it never grants access to tools, sources, or a user identity.
type ConversationExternalAgentCapabilities struct {
	Steering         bool `json:"steering"`
	Cancellation     bool `json:"cancellation"`
	Resume           bool `json:"resume"`
	StructuredOutput bool `json:"structured_output"`
	ExecutionDetails bool `json:"execution_details"`
}

// ConversationExternalAgentConfig registers a peer execution endpoint by
// protocol and capability. Credentials and network locations stay with the
// calling process; Domainry stores no reusable remote secret on the Agent.
type ConversationExternalAgentConfig struct {
	Protocol     string                                `json:"protocol"`
	Version      string                                `json:"version"`
	Capabilities ConversationExternalAgentCapabilities `json:"capabilities"`
}

type ConversationExternalAgentEvent struct {
	Attempt   int             `json:"attempt"`
	Seq       int64           `json:"seq"`
	Kind      string          `json:"kind"`
	Phase     string          `json:"phase,omitempty"`
	Summary   string          `json:"summary"`
	Detail    string          `json:"detail,omitempty"`
	Tool      string          `json:"tool,omitempty"`
	Progress  *float64        `json:"progress,omitempty"`
	Usage     json.RawMessage `json:"usage,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// ConversationExternalAgentExecution is the durable public projection of one
// external claim. Events are bounded; EventsComplete states whether the list
// contains the whole execution history. Final delivery remains the existing
// structured delegation delivery, never an inferred replacement for details.
type ConversationExternalAgentExecution struct {
	Protocol         string                                `json:"protocol"`
	Version          string                                `json:"version"`
	Status           string                                `json:"status"`
	SessionID        string                                `json:"session_id,omitempty"`
	ClaimClientID    string                                `json:"claim_client_id,omitempty"`
	Attempt          int                                   `json:"attempt"`
	Capabilities     ConversationExternalAgentCapabilities `json:"capabilities"`
	LastEventSeq     int64                                 `json:"last_event_seq"`
	Events           []ConversationExternalAgentEvent      `json:"events"`
	EventsComplete   bool                                  `json:"events_complete"`
	DetailsAvailable bool                                  `json:"details_available"`
	StopRequested    bool                                  `json:"stop_requested,omitempty"`
	StopAcknowledged bool                                  `json:"stop_acknowledged,omitempty"`
	EffectState      string                                `json:"effect_state,omitempty"`
	ClaimedAt        *time.Time                            `json:"claimed_at,omitempty"`
	UpdatedAt        time.Time                             `json:"updated_at"`
}

type ConversationExternalAgentAssignmentQuery struct {
	AgentID string `json:"agent_id"`
	Limit   int    `json:"limit,omitempty"`
}

type ConversationExternalAgentAssignment struct {
	Delegation ConversationDelegationDetail `json:"delegation"`
}

type ConversationExternalAgentAssignmentPage struct {
	Items    []ConversationExternalAgentAssignment `json:"items"`
	Complete bool                                  `json:"complete"`
}

type ConversationExternalAgentClaim struct {
	ClientID     string                                `json:"client_id"`
	AgentID      string                                `json:"agent_id"`
	Capabilities ConversationExternalAgentCapabilities `json:"capabilities"`
}

type ConversationExternalAgentClaimReceipt struct {
	Task   ConversationTaskDetail `json:"task"`
	Replay bool                   `json:"replay"`
}

type ConversationExternalAgentReport struct {
	ClientID             string                           `json:"client_id"`
	AgentID              string                           `json:"agent_id"`
	SessionID            string                           `json:"session_id"`
	ExpectedLastEventSeq int64                            `json:"expected_last_event_seq"`
	Events               []ConversationExternalAgentEvent `json:"events"`
	AcknowledgedMessages []string                         `json:"acknowledged_messages,omitempty"`
	EffectState          string                           `json:"effect_state,omitempty"`
}

type ConversationExternalAgentReportReceipt struct {
	Task   ConversationTaskDetail `json:"task"`
	Replay bool                   `json:"replay"`
}

// ConversationExternalAgentService is an optional ConversationService
// extension for peer Agents running in other processes or services.
type ConversationExternalAgentService interface {
	ConversationExternalAgentAssignments(context.Context, ConversationExternalAgentAssignmentQuery, ConversationAuthority) (ConversationExternalAgentAssignmentPage, error)
	ClaimConversationExternalAgentTask(context.Context, string, ConversationExternalAgentClaim, ConversationAuthority) (ConversationExternalAgentClaimReceipt, error)
	ReportConversationExternalAgentTask(context.Context, string, ConversationExternalAgentReport, ConversationAuthority) (ConversationExternalAgentReportReceipt, error)
}
