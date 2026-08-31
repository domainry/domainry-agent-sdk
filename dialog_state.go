package agentsdk

import "context"

type AgentAuthority struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id,omitempty"`
	RoleKey     string `json:"role_key,omitempty"`
}

type AgentSession struct {
	ExternalSessionID string         `json:"external_session_id"`
	AgentSessionID    string         `json:"agent_session_id,omitempty"`
	Title             string         `json:"title"`
	LastSummary       string         `json:"last_summary,omitempty"`
	Mode              string         `json:"mode,omitempty"`
	WorkspaceID       string         `json:"workspace_id,omitempty"`
	ObjectKey         string         `json:"object_key,omitempty"`
	RecordID          string         `json:"record_id,omitempty"`
	UserID            string         `json:"user_id,omitempty"`
	Role              string         `json:"role,omitempty"`
	Archived          bool           `json:"archived"`
	Context           map[string]any `json:"context,omitempty"`
	CreatedAt         int64          `json:"created_at"`
	UpdatedAt         int64          `json:"updated_at"`
}

type AgentSessionUpsertRequest struct {
	ExternalSessionID string         `json:"external_session_id,omitempty"`
	AgentSessionID    string         `json:"agent_session_id,omitempty"`
	Title             string         `json:"title,omitempty"`
	LastSummary       string         `json:"last_summary,omitempty"`
	Mode              string         `json:"mode,omitempty"`
	Context           map[string]any `json:"context,omitempty"`
	Archived          bool           `json:"archived,omitempty"`
}

type AgentSessionQuery struct {
	Search          string `json:"search,omitempty"`
	ObjectKey       string `json:"object_key,omitempty"`
	RecordID        string `json:"record_id,omitempty"`
	IncludeArchived bool   `json:"include_archived,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type AgentProposal struct {
	ProposalID     string         `json:"proposal_id"`
	Status         string         `json:"status"`
	Title          string         `json:"title,omitempty"`
	Summary        string         `json:"summary,omitempty"`
	Source         string         `json:"source,omitempty"`
	Reference      string         `json:"reference,omitempty"`
	Actor          string         `json:"actor,omitempty"`
	WorkspaceID    string         `json:"workspace_id,omitempty"`
	UserID         string         `json:"user_id,omitempty"`
	Role           string         `json:"role,omitempty"`
	Proposed       map[string]any `json:"proposed,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	Execution      map[string]any `json:"execution,omitempty"`
	DecisionReason string         `json:"decision_reason,omitempty"`
	DecisionActor  string         `json:"decision_actor,omitempty"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
	DecidedAt      int64          `json:"decided_at,omitempty"`
	Audited        bool           `json:"audited"`
}

type AgentProposalDecision struct {
	ProposalID        string         `json:"proposal_id"`
	Decision          string         `json:"decision"`
	Reason            string         `json:"reason,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	Execution         map[string]any `json:"execution,omitempty"`
	ExpectedUpdatedAt int64          `json:"expected_updated_at,omitempty"`
}

// AgentDialogStateService owns Agent dialog/session and proposal state. The
// Runtime supplies authenticated authority and retains approval execution and
// other host business effects.
type AgentDialogStateService interface {
	ListSessions(context.Context, AgentSessionQuery, AgentAuthority) ([]AgentSession, error)
	UpsertSession(context.Context, AgentSessionUpsertRequest, AgentAuthority) (AgentSession, error)
	SetSessionArchived(context.Context, string, bool, AgentAuthority) (AgentSession, error)
	ListProposals(context.Context, string, AgentAuthority) ([]AgentProposal, error)
	GetProposal(context.Context, string, AgentAuthority) (AgentProposal, error)
	StoreProposal(context.Context, AgentProposal) (AgentProposal, error)
	DecideProposal(context.Context, AgentProposalDecision, AgentAuthority) (AgentProposal, error)
}

// AgentDialogStateBinding is optional so runner-only SDK implementations stay
// source compatible while product HTTP and Runtime composition can fail closed
// when dialog state is required.
type AgentDialogStateBinding interface {
	DialogState() AgentDialogStateService
}
