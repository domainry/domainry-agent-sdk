package persistence

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type ConversationDelegationAdmission struct {
	SourceAgent agentsdk.ConversationAgentSnapshot
	Request     agentsdk.ConversationDelegationCreate
	FromAgentID string
	SourceRunID string
	Agent       agentsdk.ConversationAgentSnapshot
	Task        agentsdk.ConversationTask
}

type ConversationDelegationTransferAdmission struct {
	Request agentsdk.ConversationDelegationUpdate
	Agent   agentsdk.ConversationAgentSnapshot
	Task    agentsdk.ConversationTask
	Handoff agentsdk.ConversationDelegationHandoff
}

type ConversationDeliveryVerificationRepository interface {
	ConversationDeliveryHistory(context.Context, string, int64, agentsdk.ConversationAuthority) (agentsdk.ConversationDeliveryHistory, error)
	PreviewConversationDeliveryVerification(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationDeliveryVerification, error)
}

type ConversationDelegationTransferRepository interface {
	ConversationDelegationTransferReceipt(context.Context, string, agentsdk.ConversationDelegationUpdate, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegation, bool, error)
	ConversationDelegationAssignments(context.Context, string, agentsdk.ConversationAuthority) ([]agentsdk.ConversationDelegationAssignment, error)
	PrepareConversationDelegationHandoff(context.Context, string, int64, string, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegationHandoff, error)
	TransferConversationDelegation(context.Context, string, ConversationDelegationTransferAdmission, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegation, error)
	ReuseConversationDelegationEffect(context.Context, ConversationClaim, int, agentsdk.ConversationToolCall, agentsdk.ConversationToolDefinition) (agentsdk.ConversationToolResult, bool, error)
}

// Collaboration is persisted by the Agent owner. Mutation IDs and revisions
// are checked in the same transaction as the resulting messages/task records.
type ConversationCollaborationRepository interface {
	ConversationAgreementHistory(context.Context, string, int64, agentsdk.ConversationAuthority) (agentsdk.ConversationAgreementHistory, error)
	ConversationPeerInbox(context.Context, string, agentsdk.ConversationAuthority) ([]agentsdk.ConversationAgentMessage, error)
	LaunchConversationPeerMessage(context.Context, string) (agentsdk.ConversationRun, bool, error)
	ConversationAgents(context.Context, agentsdk.ConversationAuthority) ([]agentsdk.ConversationAgent, error)
	ConversationAgent(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationAgent, error)
	WriteConversationAgent(context.Context, string, agentsdk.ConversationAgentWrite, agentsdk.ConversationAuthority) (agentsdk.ConversationAgent, error)
	CreateConversationDelegation(context.Context, ConversationDelegationAdmission, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegation, error)
	ConversationDelegations(context.Context, string, agentsdk.ConversationAuthority) ([]agentsdk.ConversationDelegation, error)
	ConversationDelegation(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegation, error)
	UpdateConversationDelegation(context.Context, string, agentsdk.ConversationDelegationUpdate, agentsdk.ConversationAuthority) (agentsdk.ConversationDelegation, error)
	ConversationAgentMessages(context.Context, string, agentsdk.ConversationAuthority) ([]agentsdk.ConversationAgentMessage, error)
	SendConversationAgentMessage(context.Context, string, agentsdk.ConversationAgentMessageSend, string, agentsdk.ConversationAuthority) (agentsdk.ConversationAgentMessage, error)
}

// Discovery reads only owner-owned execution facts. No tool contents or other
// owners' task identities cross this port. History is a bounded recent sample.
type ConversationAgentObservation struct {
	ConfigurationDigest string
	AgentID             string
	Revision            int64
	Model               agentsdk.ConversationModelIdentity
	TaskType            string
	Status              string
	Usage               map[string]any
	DurationMillis      int64
	ToolCalls           int
}
type ConversationAgentLoad struct{ Running, Queued, Waiting int }
type ConversationAgentObservations struct {
	Load            map[string]ConversationAgentLoad
	History         []ConversationAgentObservation
	HistoryComplete bool
}
type ConversationAgentDiscoveryRepository interface {
	ConversationAgentObservations(context.Context, agentsdk.ConversationAuthority) (ConversationAgentObservations, error)
}

// Immutable disagreement revisions are read independently from bounded summaries.
type ConversationDisagreementRepository interface {
	ConversationDisagreementHistory(context.Context, string, string, int64, agentsdk.ConversationAuthority) (agentsdk.ConversationDisagreementHistory, error)
}
