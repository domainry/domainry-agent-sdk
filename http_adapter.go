package agentsdk

import (
	"fmt"
	"strings"

	actioncontract "github.com/domainry/domainry-foundation/action"
)

const (
	AgentHTTPAdapterContractVersion = "domainry-agent-http-adapter-v1"
	AgentHTTPAdapterOwner           = "agent"
	AgentHTTPAdapterName            = "dialog_state"
	AgentAuthorizationOwner         = "module:agent"

	AgentCapabilityDialog        = "agent.dialog"
	AgentCapabilityProposals     = "agent.proposals"
	AgentCapabilityOperations    = "agent.operations"
	AgentCapabilityToolGateway   = "agent.tool_gateway"
	AgentCapabilityTaskExecution = "agent.task_execution"

	ActionAgentRunsExecute      = "agent.runs.execute"
	ActionAgentRunsStream       = "agent.runs.stream"
	ActionAgentRunsGet          = "agent.runs.get"
	ActionAgentSessionsList     = "agent.sessions.list"
	ActionAgentSessionsUpsert   = "agent.sessions.upsert"
	ActionAgentSessionsArchive  = "agent.sessions.archive"
	ActionAgentSessionsRestore  = "agent.sessions.restore"
	ActionAgentProposalsList    = "agent.proposals.list"
	ActionAgentProposalsGet     = "agent.proposals.get"
	ActionAgentProposalsCreate  = "agent.proposals.create"
	ActionAgentProposalsApprove = "agent.proposals.approve"
	ActionAgentProposalsReject  = "agent.proposals.reject"
	ActionAgentTaskRunsStart    = "agent.task_runs.start"
	ActionAgentTaskRunsGet      = "agent.task_runs.get"
	ActionAgentTaskToolsInvoke  = "agent.task_tools.invoke"
	ActionAgentAnalysisQuery    = "agent.analysis.query"
	ActionAgentDiagnosticsRead  = "agent.diagnostics.read"
	ActionAgentTasksList        = "agent.tasks.list"
	ActionAgentTasksGet         = "agent.tasks.get"
	ActionAgentTasksRetry       = "agent.tasks.retry"
	ActionAgentTasksCancel      = "agent.tasks.cancel"
	ActionAgentTasksResolve     = "agent.tasks.resolve"
	ActionAgentTasksReconcile   = "agent.tasks.reconcile"

	ActionAgentTaskExecutionStart  = "agent.task_execution.start"
	ActionAgentTaskExecutionPoll   = "agent.task_execution.poll"
	ActionAgentTaskExecutionCancel = "agent.task_execution.cancel"
)

// HTTPRouteContract is the source-owned Agent product-route manifest. Runtime
// may host these routes, but it must not recreate their authorization or
// governance semantics.
type HTTPRouteContract struct {
	Action actioncontract.ActionDefinition `json:"action"`
}

func (route HTTPRouteContract) Pattern() string {
	if route.Action.HTTP == nil {
		return ""
	}
	return route.Action.HTTP.Method + " " + route.Action.HTTP.RouteTemplate
}

type HTTPAdapterContract struct {
	ContractVersion string              `json:"contract_version"`
	Owner           string              `json:"owner"`
	Name            string              `json:"name"`
	Routes          []HTTPRouteContract `json:"routes"`
}

// AgentAuthorizationActions is the complete source-owned Agent product and
// host-capability manifest. HTTP routes, direct TaskRunner calls, permission
// reconcile and configuration projections all consume this same batch.
func AgentAuthorizationActions() ([]actioncontract.ActionDefinition, error) {
	definitions := []actioncontract.ActionDefinition{
		agentPrincipalAction(ActionAgentRunsExecute, AgentCapabilityDialog, "Agent dialog and analysis", "Run conversation", "POST /agent/runs", actioncontract.EffectWrite, "caller_key_required", "agent_interactive_execution"),
		agentPrincipalAction(ActionAgentRunsStream, AgentCapabilityDialog, "Agent dialog and analysis", "Stream conversation", "POST /agent/runs/stream", actioncontract.EffectWrite, "caller_key_required", "agent_interactive_execution"),
		agentPrincipalAction(ActionAgentSessionsList, AgentCapabilityDialog, "Agent dialog and analysis", "List sessions", "GET /agent/sessions", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentSessionsUpsert, AgentCapabilityDialog, "Agent dialog and analysis", "Save session", "POST /agent/sessions", actioncontract.EffectWrite, "not_supported", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentSessionsArchive, AgentCapabilityDialog, "Agent dialog and analysis", "Archive session", "POST /agent/sessions/{externalSessionID}/archive", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentSessionsRestore, AgentCapabilityDialog, "Agent dialog and analysis", "Restore session", "POST /agent/sessions/{externalSessionID}/restore", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsList, AgentCapabilityProposals, "Agent proposals", "List proposals", "GET /agent/proposals", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentProposalsGet, AgentCapabilityProposals, "Agent proposals", "Read proposal", "GET /agent/proposals/{proposalID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentProposalsCreate, AgentCapabilityProposals, "Agent proposals", "Create proposal", "POST /agent/proposals", actioncontract.EffectWrite, "not_supported", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsApprove, AgentCapabilityProposals, "Agent proposals", "Approve proposal", "POST /agent/proposals/{proposalID}/approve", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsReject, AgentCapabilityProposals, "Agent proposals", "Reject proposal", "POST /agent/proposals/{proposalID}/reject", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentRunsGet, AgentCapabilityDialog, "Agent dialog and analysis", "Read conversation run", "GET /agent/runs/{runID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentTaskRunsStart, AgentCapabilityTaskExecution, "Agent task execution", "Start task run with private attachments", "POST /agent/task-runs", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentTaskRunsGet, AgentCapabilityDialog, "Agent dialog and analysis", "Read own task run", "GET /agent/task-runs/{taskRunID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentDelegatedCredentialAction(ActionAgentTaskToolsInvoke, AgentCapabilityToolGateway, "Agent task tool gateway", "Invoke task tool", "POST /agent/task-tools/invoke", actioncontract.EffectWrite, "credential_payload_key_required", "credential_scoped_tool_audit"),
		agentPrincipalAction(ActionAgentAnalysisQuery, AgentCapabilityDialog, "Agent dialog and analysis", "Query analysis", "POST /agent/analysis/query", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentDiagnosticsRead, AgentCapabilityDialog, "Agent dialog and analysis", "Read diagnostics", "GET /agent/diagnostics", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksList, AgentCapabilityOperations, "Agent task operations", "List task runs", "GET /agent/tasks", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksGet, AgentCapabilityOperations, "Agent task operations", "Read task run", "GET /agent/tasks/{taskRunID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksRetry, AgentCapabilityOperations, "Agent task operations", "Retry task run", "POST /agent/tasks/{taskRunID}/retry", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksCancel, AgentCapabilityOperations, "Agent task operations", "Cancel task run", "POST /agent/tasks/{taskRunID}/cancel", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksResolve, AgentCapabilityOperations, "Agent task operations", "Resolve task run", "POST /agent/tasks/{taskRunID}/resolve", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksReconcile, AgentCapabilityOperations, "Agent task operations", "Reconcile task run", "POST /agent/tasks/{taskRunID}/reconcile", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentTaskExecutionAction(ActionAgentTaskExecutionStart, "Start task execution", actioncontract.EffectWrite, "request_idempotency_key"),
		agentTaskExecutionAction(ActionAgentTaskExecutionPoll, "Poll task execution", actioncontract.EffectRead, "not_applicable"),
		agentTaskExecutionAction(ActionAgentTaskExecutionCancel, "Cancel task execution", actioncontract.EffectWrite, "request_idempotency_key"),
	}
	definitions = append(definitions, conversationActions()...)
	definitions = append(definitions, ConversationToolActions()...)
	definitions = append(definitions, ConversationCollaborationActions()...)
	result := make([]actioncontract.ActionDefinition, 0, len(definitions))
	for _, definition := range definitions {
		normalized, err := actioncontract.NormalizeDefinition(definition)
		if err != nil {
			return nil, fmt.Errorf("normalize Agent Action %q: %w", definition.Key, err)
		}
		result = append(result, normalized)
	}
	return result, nil
}

// AgentHTTPAdapterContract returns the statically compiled HTTP projection.
// Invalid source-owned definitions are programmer errors and fail immediately;
// host readiness paths should use CompileAgentHTTPAdapterContract so they can
// return a contextual assembly error instead.
func AgentHTTPAdapterContract() HTTPAdapterContract {
	contract, err := CompileAgentHTTPAdapterContract()
	if err != nil {
		panic("compile Agent HTTP Adapter contract: " + err.Error())
	}
	return contract
}

// CompileAgentHTTPAdapterContract validates and projects the source-owned
// Action manifest for host assembly and readiness checks.
func CompileAgentHTTPAdapterContract() (HTTPAdapterContract, error) {
	definitions, err := AgentAuthorizationActions()
	if err != nil {
		return HTTPAdapterContract{}, err
	}
	routes := make([]HTTPRouteContract, 0, len(definitions))
	patterns := map[string]bool{}
	for _, definition := range definitions {
		if definition.HTTP == nil {
			continue
		}
		route := HTTPRouteContract{Action: definition}
		pattern := route.Pattern()
		if patterns[pattern] {
			return HTTPAdapterContract{}, fmt.Errorf("Agent HTTP manifest repeats %q", pattern)
		}
		patterns[pattern] = true
		routes = append(routes, route)
	}
	return HTTPAdapterContract{
		ContractVersion: AgentHTTPAdapterContractVersion,
		Owner:           AgentHTTPAdapterOwner,
		Name:            AgentHTTPAdapterName,
		Routes:          routes,
	}, nil
}

func agentPrincipalAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposurePublic}, effect, idempotency, audit, actioncontract.AuthorizationAuthenticated, false)
}

func agentDelegatedCredentialAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposurePublic}, effect, idempotency, audit, actioncontract.AuthorizationSigned, false)
}

func agentRoleAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposureManagement, actioncontract.ExposureOps}, effect, idempotency, audit, actioncontract.AuthorizationAuthenticated, true)
}

func agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, exposures []actioncontract.Exposure, effect actioncontract.EffectClass, idempotency, audit string, strategy actioncontract.AuthorizationStrategy, requirePermission bool) actioncontract.ActionDefinition {
	method, path, _ := strings.Cut(pattern, " ")
	separator := strings.LastIndex(key, ".")
	risk := actioncontract.RiskMedium
	if effect == actioncontract.EffectRead {
		risk = actioncontract.RiskLow
	}
	definition := actioncontract.ActionDefinition{
		Key: key, Owner: AgentAuthorizationOwner, SourceKind: "module_http", CapabilityKey: capabilityKey, CapabilityLabel: capabilityLabel,
		OperationKey: key[separator+1:], OperationLabel: operationLabel, Label: operationLabel, Exposures: exposures,
		HTTP: &actioncontract.HTTPBinding{Method: method, RouteTemplate: path}, EffectClass: effect, RiskLevel: risk,
		IdempotencyDecision: idempotency, AuditClass: audit, LifecycleStatus: actioncontract.LifecycleActive,
	}
	switch strategy {
	case actioncontract.AuthorizationAuthenticated:
		definition.Authorization = actioncontract.Authorization{Strategy: actioncontract.AuthorizationAuthenticated}
	case actioncontract.AuthorizationSigned:
		definition.Authorization = actioncontract.Authorization{Strategy: actioncontract.AuthorizationSigned, PolicyKey: "agent.task_tool_credential"}
	}
	if requirePermission {
		definition.Permission = &actioncontract.PermissionDefinition{Key: key, Owner: definition.Owner, ResourceKey: key[:separator], OperationKey: key[separator+1:], Label: operationLabel, Category: capabilityLabel, LifecycleStatus: actioncontract.LifecycleActive}
	}
	return definition
}

func agentTaskExecutionAction(key, operationLabel string, effect actioncontract.EffectClass, idempotency string) actioncontract.ActionDefinition {
	separator := strings.LastIndex(key, ".")
	risk := actioncontract.RiskMedium
	if effect == actioncontract.EffectRead {
		risk = actioncontract.RiskLow
	}
	return actioncontract.ActionDefinition{
		Key: key, Owner: AgentAuthorizationOwner, SourceKind: "module_http",
		CapabilityKey: AgentCapabilityTaskExecution, CapabilityLabel: "Agent task execution",
		OperationKey: key[separator+1:], OperationLabel: operationLabel, Label: operationLabel,
		Exposures:     []actioncontract.Exposure{actioncontract.ExposureOps},
		Authorization: actioncontract.Authorization{Strategy: actioncontract.AuthorizationSigned, PolicyKey: "agent.runtime_host", Audiences: []string{AgentRuntimeServiceAudience}},
		NonHTTP:       []actioncontract.NonHTTPBinding{{Kind: "sdk", InvocationKey: key}},
		EffectClass:   effect, RiskLevel: risk, IdempotencyDecision: idempotency,
		AuditClass: "agent_task_execution", LifecycleStatus: actioncontract.LifecycleActive,
	}
}
