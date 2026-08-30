package agentmodel

// AgentAuthorizationEvidence is host-issued proof of an authorization
// decision persisted by the Agent owner and never sent to a model provider.
type AgentAuthorizationEvidence struct {
	Decision              string   `json:"decision"`
	Code                  string   `json:"code"`
	PolicyRevision        string   `json:"policy_revision"`
	ContextRevision       string   `json:"context_revision,omitempty"`
	AuthorizationRevision string   `json:"authorization_revision"`
	AllowedObjects        []string `json:"allowed_objects"`
	AllowedActions        []string `json:"allowed_actions"`
	AllowedOutcomes       []string `json:"allowed_outcomes"`
	AllowedTools          []string `json:"allowed_tools"`
}
