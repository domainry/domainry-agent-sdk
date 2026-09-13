package agentsdk

import (
	"context"

	action "github.com/domainry/domainry-foundation/action"
)

const ConversationCollaborationPermissionPrefix = "agent.collaboration."

// Collaboration authorization is independent of tool selection and of the
// underlying source owner's data policy. Resource facts are loaded by the
// service; neither a model nor a browser may supply an owner or acting Agent.
type ConversationCollaborationAuthorizationRequest struct {
	Authority    ConversationAuthority
	Operations   []string
	OwnerUserID  string
	DelegationID string
	AgentID      string
	FromAgentID  string
	ToAgentID    string
}

type ConversationCollaborationAuthorization struct {
	Allowed  []string `json:"allowed"`
	Revision string   `json:"revision"`
}

type ConversationCollaborationAuthorizer interface {
	AuthorizeConversationCollaboration(context.Context, ConversationCollaborationAuthorizationRequest) (ConversationCollaborationAuthorization, error)
}

type ConversationCollaborationAccess struct {
	View          bool `json:"view"`
	Receive       bool `json:"receive"`
	Manage        bool `json:"manage"`
	Communicate   bool `json:"communicate"`
	ExecutionRead bool `json:"execution_read"`
	DeliveryRead  bool `json:"delivery_read"`
	Share         bool `json:"share"`
}

func ConversationCollaborationOperations() []string {
	return []string{"discover", "configure", "initiate", "view", "receive", "manage", "communicate", "execution_read", "delivery_read", "share"}
}

func ConversationCollaborationPermission(operation string) *action.PermissionDefinition {
	labels := map[string]string{"discover": "发现 Agent", "configure": "配置 Agent", "initiate": "发起委派", "view": "查看委派约定与状态", "receive": "接收与交付委派", "manage": "管理与验收委派", "communicate": "委派沟通", "execution_read": "查看委派执行", "delivery_read": "读取委派交付", "share": "共享委派资料"}
	label, ok := labels[operation]
	if !ok {
		return nil
	}
	return &action.PermissionDefinition{Key: ConversationCollaborationPermissionPrefix + operation, Owner: AgentAuthorizationOwner, ResourceKey: "agent.collaboration", OperationKey: operation, Label: label, Category: "Agent collaboration", LifecycleStatus: action.LifecycleActive}
}

func ConversationCollaborationActions() []action.ActionDefinition {
	out := []action.ActionDefinition{}
	for _, operation := range ConversationCollaborationOperations() {
		p := ConversationCollaborationPermission(operation)
		effect := action.EffectWrite
		if operation == "discover" || operation == "view" || operation == "execution_read" || operation == "delivery_read" {
			effect = action.EffectRead
		}
		out = append(out, action.ActionDefinition{Key: p.Key, Owner: AgentAuthorizationOwner, SourceKind: "agent_collaboration", CapabilityKey: AgentCapabilityCollaboration, CapabilityLabel: "Peer Agent collaboration", OperationKey: operation, OperationLabel: p.Label, Label: p.Label, Exposures: []action.Exposure{action.ExposurePublic}, Authorization: action.Authorization{Strategy: action.AuthorizationAuthenticated}, NonHTTP: []action.NonHTTPBinding{{Kind: "sdk", InvocationKey: p.Key}}, EffectClass: effect, RiskLevel: action.RiskLow, IdempotencyDecision: "request_contract", AuditClass: "agent_collaboration_policy", LifecycleStatus: action.LifecycleActive, Permission: p})
	}
	return out
}
