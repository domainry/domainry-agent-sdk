package agentsdk

import (
	"encoding/json"
	"fmt"
	"strings"

	actioncontract "github.com/domainry/domainry-foundation/action"
	"github.com/domainry/domainry-foundation/modulecapability"
)

const (
	AgentHTTPSurfaceContractVersion = "domainry-agent-http-surface-v2"
	AgentHTTPSurfaceOwner           = "agent"
	AgentHTTPSurfaceName            = "dialog_state"
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

type HTTPSurfaceContract struct {
	ContractVersion string                                `json:"contract_version"`
	Owner           string                                `json:"owner"`
	Name            string                                `json:"name"`
	Routes          []HTTPRouteContract                   `json:"routes"`
	OpenAPI         map[string]map[string]any             `json:"openapi_operations"`
	Components      map[string]map[string]json.RawMessage `json:"components"`
}

// AgentAuthorizationActions is the complete source-owned Agent product and
// host-capability manifest. HTTP routes, direct TaskRunner calls, permission
// reconcile and configuration projections all consume this same batch.
func AgentAuthorizationActions() ([]actioncontract.ActionDefinition, error) {
	definitions := []actioncontract.ActionDefinition{
		agentPrincipalAction(ActionAgentRunsExecute, AgentCapabilityDialog, "Agent dialog and analysis", "Run conversation", "POST /agent-dialog/runs", actioncontract.EffectWrite, "caller_key_required", "agent_interactive_execution"),
		agentPrincipalAction(ActionAgentRunsStream, AgentCapabilityDialog, "Agent dialog and analysis", "Stream conversation", "POST /agent-dialog/runs/stream", actioncontract.EffectWrite, "caller_key_required", "agent_interactive_execution"),
		agentPrincipalAction(ActionAgentSessionsList, AgentCapabilityDialog, "Agent dialog and analysis", "List sessions", "GET /agent-dialog/sessions", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentSessionsUpsert, AgentCapabilityDialog, "Agent dialog and analysis", "Save session", "POST /agent-dialog/sessions", actioncontract.EffectWrite, "not_supported", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentSessionsArchive, AgentCapabilityDialog, "Agent dialog and analysis", "Archive session", "POST /agent-dialog/sessions/{externalSessionID}/archive", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentSessionsRestore, AgentCapabilityDialog, "Agent dialog and analysis", "Restore session", "POST /agent-dialog/sessions/{externalSessionID}/restore", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsList, AgentCapabilityProposals, "Agent proposals", "List proposals", "GET /agent-dialog/proposals", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentProposalsGet, AgentCapabilityProposals, "Agent proposals", "Read proposal", "GET /agent-dialog/proposals/{proposalID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentProposalsCreate, AgentCapabilityProposals, "Agent proposals", "Create proposal", "POST /agent-dialog/proposals", actioncontract.EffectWrite, "not_supported", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsApprove, AgentCapabilityProposals, "Agent proposals", "Approve proposal", "POST /agent-dialog/proposals/{proposalID}/approve", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentProposalsReject, AgentCapabilityProposals, "Agent proposals", "Reject proposal", "POST /agent-dialog/proposals/{proposalID}/reject", actioncontract.EffectWrite, "natural", "mutation_audit_required"),
		agentPrincipalAction(ActionAgentRunsGet, AgentCapabilityDialog, "Agent dialog and analysis", "Read conversation run", "GET /agent-dialog/runs/{runID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentPrincipalAction(ActionAgentTaskRunsGet, AgentCapabilityDialog, "Agent dialog and analysis", "Read own task run", "GET /agent-dialog/task-runs/{taskRunID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentDelegatedCredentialAction(ActionAgentTaskToolsInvoke, AgentCapabilityToolGateway, "Agent task tool gateway", "Invoke task tool", "POST /agent-dialog/task-tools/invoke", actioncontract.EffectWrite, "credential_payload_key_required", "credential_scoped_tool_audit"),
		agentPrincipalAction(ActionAgentAnalysisQuery, AgentCapabilityDialog, "Agent dialog and analysis", "Query analysis", "POST /agent-dialog/analysis/query", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentDiagnosticsRead, AgentCapabilityDialog, "Agent dialog and analysis", "Read diagnostics", "GET /agent-dialog/diagnostics", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksList, AgentCapabilityOperations, "Agent task operations", "List task runs", "GET /operations/agent/tasks", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksGet, AgentCapabilityOperations, "Agent task operations", "Read task run", "GET /operations/agent/tasks/{taskRunID}", actioncontract.EffectRead, "not_applicable", "owner_read_audit_policy"),
		agentRoleAction(ActionAgentTasksRetry, AgentCapabilityOperations, "Agent task operations", "Retry task run", "POST /operations/agent/tasks/{taskRunID}/retry", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksCancel, AgentCapabilityOperations, "Agent task operations", "Cancel task run", "POST /operations/agent/tasks/{taskRunID}/cancel", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksResolve, AgentCapabilityOperations, "Agent task operations", "Resolve task run", "POST /operations/agent/tasks/{taskRunID}/resolve", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentRoleAction(ActionAgentTasksReconcile, AgentCapabilityOperations, "Agent task operations", "Reconcile task run", "POST /operations/agent/tasks/{taskRunID}/reconcile", actioncontract.EffectWrite, "caller_key_required", "mutation_audit_required"),
		agentTaskExecutionAction(ActionAgentTaskExecutionStart, "Start task execution", actioncontract.EffectWrite, "request_idempotency_key"),
		agentTaskExecutionAction(ActionAgentTaskExecutionPoll, "Poll task execution", actioncontract.EffectRead, "not_applicable"),
		agentTaskExecutionAction(ActionAgentTaskExecutionCancel, "Cancel task execution", actioncontract.EffectWrite, "request_idempotency_key"),
	}
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

// AgentHTTPSurfaceContract returns the statically compiled HTTP projection.
// Invalid source-owned definitions are programmer errors and fail immediately;
// host readiness paths should use CompileAgentHTTPSurfaceContract so they can
// return a contextual assembly error instead.
func AgentHTTPSurfaceContract() HTTPSurfaceContract {
	contract, err := CompileAgentHTTPSurfaceContract()
	if err != nil {
		panic("compile Agent HTTP Surface contract: " + err.Error())
	}
	return contract
}

// CompileAgentHTTPSurfaceContract validates and projects the source-owned
// Action manifest for host assembly and readiness checks.
func CompileAgentHTTPSurfaceContract() (HTTPSurfaceContract, error) {
	definitions, err := AgentAuthorizationActions()
	if err != nil {
		return HTTPSurfaceContract{}, err
	}
	routes := make([]HTTPRouteContract, 0, 22)
	patterns := map[string]bool{}
	for _, definition := range definitions {
		if definition.HTTP == nil {
			continue
		}
		route := HTTPRouteContract{Action: definition}
		pattern := route.Pattern()
		if patterns[pattern] {
			return HTTPSurfaceContract{}, fmt.Errorf("Agent HTTP manifest repeats %q", pattern)
		}
		patterns[pattern] = true
		routes = append(routes, route)
	}
	operations := agentHTTPSurfaceOperations()
	for pattern := range patterns {
		if len(operations[pattern]) == 0 {
			return HTTPSurfaceContract{}, fmt.Errorf("Agent Action route %q has no OpenAPI operation", pattern)
		}
	}
	for pattern := range operations {
		if !patterns[pattern] {
			return HTTPSurfaceContract{}, fmt.Errorf("Agent OpenAPI operation %q has no Action route", pattern)
		}
	}
	return HTTPSurfaceContract{
		ContractVersion: AgentHTTPSurfaceContractVersion,
		Owner:           AgentHTTPSurfaceOwner,
		Name:            AgentHTTPSurfaceName,
		Routes:          routes,
		OpenAPI:         operations,
		Components:      agentHTTPComponents(),
	}, nil
}

func HTTPSurfaceOpenAPIOperations() map[string]map[string]any {
	return AgentHTTPSurfaceContract().OpenAPI
}

// HTTPSurfaceReferencedComponents returns the exact transitive component
// closure used by the selected operations. Capability categories use it to
// avoid returning unrelated Agent schemas in every bounded batch.
func HTTPSurfaceReferencedComponents(operations map[string]map[string]any) map[string]map[string]json.RawMessage {
	all := agentHTTPComponents()
	wanted := map[string]bool{}
	for _, operation := range operations {
		collectAgentHTTPReferences(operation, wanted)
	}
	for changed := true; changed; {
		changed = false
		for reference := range wanted {
			group, key, ok := agentHTTPReferenceParts(reference)
			if !ok {
				continue
			}
			raw, found := all[group][key]
			if !found {
				continue
			}
			var value any
			if json.Unmarshal(raw, &value) != nil {
				continue
			}
			before := len(wanted)
			collectAgentHTTPReferences(value, wanted)
			changed = changed || len(wanted) != before
		}
	}
	result := map[string]map[string]json.RawMessage{}
	for reference := range wanted {
		group, key, ok := agentHTTPReferenceParts(reference)
		if !ok {
			continue
		}
		if raw, found := all[group][key]; found {
			if result[group] == nil {
				result[group] = map[string]json.RawMessage{}
			}
			result[group][key] = append(json.RawMessage(nil), raw...)
		}
	}
	return result
}

func agentPrincipalAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposurePublic}, effect, idempotency, audit, actioncontract.AuthorizationAuthenticated, false)
}

func agentDelegatedCredentialAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposurePublic}, effect, idempotency, audit, actioncontract.AuthorizationSigned, false)
}

func agentRoleAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, effect actioncontract.EffectClass, idempotency, audit string) actioncontract.ActionDefinition {
	return agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern, []actioncontract.Exposure{actioncontract.ExposureTenantAdmin, actioncontract.ExposureOps}, effect, idempotency, audit, actioncontract.AuthorizationAuthenticated, true)
}

func agentHTTPAction(key, capabilityKey, capabilityLabel, operationLabel, pattern string, exposures []actioncontract.Exposure, effect actioncontract.EffectClass, idempotency, audit string, strategy actioncontract.AuthorizationStrategy, requirePermission bool) actioncontract.ActionDefinition {
	method, path, _ := strings.Cut(pattern, " ")
	separator := strings.LastIndex(key, ".")
	risk := actioncontract.RiskMedium
	if effect == actioncontract.EffectRead {
		risk = actioncontract.RiskLow
	}
	definition := actioncontract.ActionDefinition{
		Key: key, Owner: AgentAuthorizationOwner, SourceKind: "module_surface", CapabilityKey: capabilityKey, CapabilityLabel: capabilityLabel,
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
		Key: key, Owner: AgentAuthorizationOwner, SourceKind: "module_surface",
		CapabilityKey: AgentCapabilityTaskExecution, CapabilityLabel: "Agent task execution",
		OperationKey: key[separator+1:], OperationLabel: operationLabel, Label: operationLabel,
		Exposures:     []actioncontract.Exposure{actioncontract.ExposureOps},
		Authorization: actioncontract.Authorization{Strategy: actioncontract.AuthorizationSigned, PolicyKey: "agent.runtime_host", Audiences: []string{AgentRuntimeServiceAudience}},
		NonHTTP:       []actioncontract.NonHTTPBinding{{Kind: "sdk", InvocationKey: key}},
		EffectClass:   effect, RiskLevel: risk, IdempotencyDecision: idempotency,
		AuditClass: "agent_task_execution", LifecycleStatus: actioncontract.LifecycleActive,
	}
}

func agentHTTPSurfaceOperations() map[string]map[string]any {
	operations := map[string]map[string]any{}
	operations["POST /agent-dialog/runs"] = agentOperation("runAgent", "Run one authenticated Agent conversation turn", bearerSecurity(), []any{headerParameter("Idempotency-Key", false)}, schemaReference("AgentInteractiveRunRequest"), true, "200", schemaReference("AgentInteractiveExecutionResult"))
	operations["POST /agent-dialog/runs/stream"] = agentOperation("runAgentStream", "Stream one authenticated Agent conversation turn", bearerSecurity(), []any{headerParameter("Idempotency-Key", false), headerParameter("Last-Event-ID", false)}, schemaReference("AgentInteractiveRunRequest"), true, "200", map[string]any{"type": "string"})
	operations["POST /agent-dialog/runs/stream"]["responses"] = map[string]any{
		"200": map[string]any{"description": "Agent server-sent event stream", "content": map[string]any{"text/event-stream": map[string]any{"schema": map[string]any{"type": "string"}}}},
		"400": errorResponse("Invalid Agent request"), "403": errorResponse("Agent access denied"), "409": errorResponse("Invalid stream cursor or conflicting run"),
	}
	operations["POST /agent-dialog/runs/stream"]["x-domainry-runtime-client-method"] = "runAgentStream"

	operations["GET /agent-dialog/sessions"] = agentOperation("listAgentSessions", "List the current principal's Agent sessions", bearerSecurity(), []any{
		queryParameter("q", map[string]any{"type": "string"}), queryParameter("object_key", map[string]any{"type": "string"}), queryParameter("record_id", map[string]any{"type": "string"}), queryParameter("archived", map[string]any{"type": "boolean"}), queryParameter("limit", map[string]any{"type": "integer", "minimum": 1, "maximum": 100}),
	}, nil, false, "200", schemaReference("AgentSessionList"))
	operations["POST /agent-dialog/sessions"] = agentOperation("upsertAgentSession", "Create or update an Agent session", bearerSecurity(), nil, schemaReference("AgentSessionUpsertRequest"), true, "200", schemaReference("AgentSession"))
	operations["POST /agent-dialog/sessions/{externalSessionID}/archive"] = agentOperation("archiveAgentSession", "Archive an Agent session", bearerSecurity(), []any{pathParameter("externalSessionID")}, nil, false, "200", schemaReference("AgentSession"))
	operations["POST /agent-dialog/sessions/{externalSessionID}/restore"] = agentOperation("restoreAgentSession", "Restore an Agent session", bearerSecurity(), []any{pathParameter("externalSessionID")}, nil, false, "200", schemaReference("AgentSession"))

	operations["GET /agent-dialog/proposals"] = agentOperation("listAgentProposals", "List the current principal's Agent proposals", bearerSecurity(), []any{queryParameter("status", map[string]any{"type": "string"})}, nil, false, "200", schemaReference("AgentProposalList"))
	operations["GET /agent-dialog/proposals/{proposalID}"] = agentOperation("getAgentProposal", "Read one Agent proposal", bearerSecurity(), []any{pathParameter("proposalID")}, nil, false, "200", schemaReference("AgentProposal"))
	operations["POST /agent-dialog/proposals"] = agentOperation("createAgentProposal", "Create an Agent proposal under the current run policy", bearerSecurity(), nil, schemaReference("AgentProposalCreateRequest"), true, "201", schemaReference("AgentProposal"))
	operations["POST /agent-dialog/proposals/{proposalID}/approve"] = agentOperation("approveAgentProposal", "Approve and execute an Agent proposal", bearerSecurity(), []any{pathParameter("proposalID")}, schemaReference("AgentProposalDecisionRequest"), false, "200", schemaReference("AgentProposal"))
	operations["POST /agent-dialog/proposals/{proposalID}/reject"] = agentOperation("rejectAgentProposal", "Reject an Agent proposal", bearerSecurity(), []any{pathParameter("proposalID")}, schemaReference("AgentProposalDecisionRequest"), false, "200", schemaReference("AgentProposal"))

	operations["GET /agent-dialog/runs/{runID}"] = agentOperation("getAgentRun", "Read one principal-owned Agent conversation run", bearerSecurity(), []any{pathParameter("runID")}, nil, false, "200", schemaReference("AgentInteractiveRun"))
	operations["GET /agent-dialog/task-runs/{taskRunID}"] = agentOperation("getAgentTaskRun", "Read one principal-owned Agent task run", bearerSecurity(), []any{pathParameter("taskRunID")}, nil, false, "200", schemaReference("AgentTaskRunView"))
	operations["POST /agent-dialog/task-tools/invoke"] = agentOperation("invokeAgentTaskTool", "Invoke a credential-scoped Agent task tool", []any{}, nil, schemaReference("AgentTaskToolInvokeRequest"), true, "200", schemaReference("AgentTaskToolResult"))
	operations["POST /agent-dialog/analysis/query"] = agentOperation("queryAgentAnalysis", "Run a principal-scoped Agent analysis query", bearerSecurity(), nil, schemaReference("AgentAnalysisQueryRequest"), true, "200", schemaReference("AgentAnalysisResult"))
	operations["POST /agent-dialog/analysis/query"]["x-domainry-runtime-client-method"] = "queryAgentAnalysis"
	operations["GET /agent-dialog/diagnostics"] = agentOperation("getAgentDiagnostics", "Inspect Agent configuration, policies, proposal state, and execution evidence", bearerSecurity(), diagnosticParameters(), nil, false, "200", schemaReference("AgentDiagnosticsResult"))

	operations["GET /operations/agent/tasks"] = agentOperation("listAgentTasks", "List Agent task runs for operators", bearerSecurity(), []any{
		queryArrayParameter("status"), queryParameter("process_id", map[string]any{"type": "string"}), queryParameter("task_key", map[string]any{"type": "string"}), queryParameter("limit", map[string]any{"type": "integer", "minimum": 1}),
	}, nil, false, "200", schemaReference("AgentTaskRunList"))
	operations["GET /operations/agent/tasks/{taskRunID}"] = agentOperation("getAgentOperationalTask", "Inspect one Agent task run", bearerSecurity(), []any{pathParameter("taskRunID")}, nil, false, "200", schemaReference("AgentTaskRunView"))
	for _, operation := range []struct{ suffix, id, summary string }{
		{"retry", "retryAgentTask", "Retry one Agent task run"}, {"cancel", "cancelAgentTask", "Cancel one Agent task run"}, {"resolve", "resolveAgentTask", "Resolve one Agent task run"}, {"reconcile", "reconcileAgentTask", "Reconcile one Agent task run with its workflow callback"},
	} {
		pattern := "POST /operations/agent/tasks/{taskRunID}/" + operation.suffix
		operations[pattern] = agentOperation(operation.id, operation.summary, bearerSecurity(), []any{pathParameter("taskRunID"), headerParameter("Idempotency-Key", true)}, schemaReference("AgentTaskOperationRequest"), true, "200", schemaReference("AgentTaskOperationResult"))
	}
	return operations
}

func agentOperation(operationID, summary string, security []any, parameters []any, requestSchema map[string]any, requestRequired bool, successStatus string, responseSchema map[string]any) map[string]any {
	operation := map[string]any{
		"operationId": operationID, "tags": []string{"Agent"}, "summary": summary, "security": security,
		"responses": map[string]any{
			successStatus: map[string]any{"description": "Agent response", "content": map[string]any{"application/json": map[string]any{"schema": responseSchema}}},
			"400":         errorResponse("Invalid Agent request"), "403": errorResponse("Agent access denied"), "404": errorResponse("Agent resource not found"), "409": errorResponse("Agent state conflict"),
		},
	}
	if len(parameters) != 0 {
		operation["parameters"] = parameters
	}
	if requestSchema != nil {
		operation["requestBody"] = map[string]any{"required": requestRequired, "content": map[string]any{"application/json": map[string]any{"schema": requestSchema}}}
	}
	return operation
}

func agentHTTPComponents() map[string]map[string]json.RawMessage {
	components := map[string]map[string]json.RawMessage{
		"securitySchemes": {"BearerAuth": rawSchema(map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT"})},
		"schemas":         {},
	}
	schemas := components["schemas"]
	schemas["Error"] = rawSchema(objectSchema(map[string]any{"code": map[string]any{"type": "string"}}, "code"))
	schemas["AgentSession"] = rawSchema(modulecapability.JSONSchemaForGoValue(AgentSession{}))
	schemas["AgentSessionUpsertRequest"] = rawSchema(modulecapability.JSONSchemaForGoValue(AgentSessionUpsertRequest{}))
	schemas["AgentSessionList"] = rawSchema(objectSchema(map[string]any{"sessions": arraySchema(schemaReference("AgentSession"))}, "sessions"))
	schemas["AgentProposal"] = rawSchema(modulecapability.JSONSchemaForGoValue(AgentProposal{}))
	schemas["AgentProposalList"] = rawSchema(objectSchema(map[string]any{"proposals": arraySchema(schemaReference("AgentProposal"))}, "proposals"))
	schemas["AgentProposalCreateRequest"] = rawSchema(objectSchema(map[string]any{
		"title": stringSchema(), "summary": stringSchema(), "source": stringSchema(), "reference": stringSchema(), "proposed": freeObjectSchema(), "metadata": freeObjectSchema(),
	}, "metadata"))
	schemas["AgentProposalDecisionRequest"] = rawSchema(objectSchema(map[string]any{"reason": stringSchema(), "metadata": freeObjectSchema()}))
	schemas["AgentInteractiveRunRequest"] = rawSchema(objectSchema(map[string]any{
		"message": stringSchema(), "timeout_seconds": map[string]any{"type": "integer", "minimum": 0}, "new_session": map[string]any{"type": "boolean"},
		"external_session_id": stringSchema(), "context": freeObjectSchema(), "metadata": freeObjectSchema(), "idempotency_key": stringSchema(),
	}, "message"))
	schemas["AgentInteractiveResult"] = rawSchema(modulecapability.JSONSchemaForGoValue(InteractiveResult{}))
	schemas["AgentInteractiveRun"] = rawSchema(agentInteractiveRunSchema())
	schemas["AgentInteractiveExecutionResult"] = rawSchema(objectSchema(map[string]any{"run": schemaReference("AgentInteractiveRun"), "result": schemaReference("AgentInteractiveResult")}, "run", "result"))
	schemas["AgentTaskRunView"] = rawSchema(agentTaskRunViewSchema())
	schemas["AgentTaskRunList"] = rawSchema(objectSchema(map[string]any{"items": arraySchema(schemaReference("AgentTaskRunView")), "count": map[string]any{"type": "integer"}}, "items", "count"))
	schemas["AgentTaskOperationRequest"] = rawSchema(objectSchema(map[string]any{"reason": map[string]any{"type": "string", "minLength": 1}}, "reason"))
	schemas["AgentTaskOperationResult"] = rawSchema(objectSchema(map[string]any{"task": schemaReference("AgentTaskRunView"), "replayed": map[string]any{"type": "boolean"}, "idempotency_key": stringSchema()}, "task", "replayed"))
	schemas["AgentTaskToolInvokeRequest"] = rawSchema(objectSchema(map[string]any{
		"credential": nonEmptyStringSchema(), "workspace_id": nonEmptyStringSchema(), "task_run_id": nonEmptyStringSchema(), "tool": nonEmptyStringSchema(), "input": freeObjectSchema(), "idempotency_key": nonEmptyStringSchema(),
	}, "credential", "workspace_id", "task_run_id", "tool", "input", "idempotency_key"))
	schemas["AgentTaskToolResult"] = rawSchema(objectSchema(map[string]any{
		"status": stringSchema(), "tool": stringSchema(), "call_ref": stringSchema(), "output": map[string]any{}, "proposal": map[string]any{}, "authorization": freeObjectSchema(),
	}, "status", "tool", "call_ref", "authorization"))
	schemas["AgentAnalysisQueryRequest"] = rawSchema(objectSchema(map[string]any{
		"intent": nonEmptyStringSchema(), "sql": stringSchema(), "metric_spec": freeObjectSchema(), "dry_run": map[string]any{"type": "boolean"}, "max_rows": map[string]any{"type": "integer", "minimum": 0},
	}, "intent"))
	schemas["AgentAnalysisResult"] = rawSchema(agentAnalysisResultSchema())
	schemas["AgentDiagnosticsResult"] = rawSchema(agentDiagnosticsResultSchema())
	return components
}

func agentInteractiveRunSchema() map[string]any {
	properties := map[string]any{
		"id": stringSchema(), "external_run_id": stringSchema(), "session_id": stringSchema(), "workspace_id": stringSchema(), "user_id": stringSchema(), "role_key": stringSchema(), "route_key": stringSchema(),
		"agent_key": stringSchema(), "entrypoint_key": stringSchema(), "context_revision": stringSchema(), "context": freeObjectSchema(), "authorization": freeObjectSchema(), "status": stringSchema(),
		"route_type": stringSchema(), "routed_target_key": stringSchema(), "routed_target_version": stringSchema(), "process_id": stringSchema(), "task_run_id": stringSchema(), "proposal_id": stringSchema(),
		"structured_result": freeObjectSchema(), "model": stringSchema(), "usage": freeObjectSchema(), "tool_call_count": map[string]any{"type": "integer"}, "tool_invocations": arraySchema(freeObjectSchema()),
		"error_code": stringSchema(), "correlation_id": stringSchema(), "idempotency_key": stringSchema(), "created_at": dateTimeSchema(), "updated_at": dateTimeSchema(), "completed_at": dateTimeSchema(), "revision": map[string]any{"type": "integer"},
	}
	return objectSchema(properties, "id", "session_id", "workspace_id", "user_id", "role_key", "context", "authorization", "status", "tool_call_count", "idempotency_key", "created_at", "updated_at", "revision")
}

func agentTaskRunViewSchema() map[string]any {
	properties := map[string]any{}
	for _, name := range []string{"id", "process_id", "node_instance_id", "interactive_run_id", "task_key", "task_version", "status", "outcome", "proposal_id", "error_code", "reconciliation_state", "created_at", "updated_at"} {
		properties[name] = stringSchema()
	}
	properties["reconciliation_required"] = map[string]any{"type": "boolean"}
	properties["identity"] = objectSchema(map[string]any{"mode": stringSchema(), "execution_user_id": stringSchema(), "execution_role_key": stringSchema(), "service_principal_key": stringSchema()}, "mode", "execution_user_id", "execution_role_key")
	properties["output"] = freeObjectSchema()
	properties["action_receipts"] = arraySchema(stringSchema())
	properties["audit_refs"] = arraySchema(stringSchema())
	for _, name := range []string{"attempt", "max_attempts", "revision"} {
		properties[name] = map[string]any{"type": "integer"}
	}
	return objectSchema(properties, "id", "task_key", "task_version", "status", "reconciliation_required", "identity", "attempt", "max_attempts", "revision", "created_at", "updated_at")
}

func agentAnalysisResultSchema() map[string]any {
	return objectSchema(map[string]any{
		"query_ref": stringSchema(), "status": stringSchema(), "execution_mode": stringSchema(), "html_fragment": stringSchema(), "rendered_report": freeObjectSchema(), "report_provenance": freeObjectSchema(),
		"scope_note": stringSchema(), "report_center_ref": stringSchema(), "proposal_suggestion": freeObjectSchema(), "workspace_id": stringSchema(), "role": stringSchema(), "object_key": stringSchema(), "report_key": stringSchema(),
		"masked_fields": arraySchema(stringSchema()), "truncated": map[string]any{"type": "boolean"}, "audit_event_key": stringSchema(), "rows": arraySchema(freeObjectSchema()), "row_count": map[string]any{"type": "integer"}, "total": map[string]any{"type": "integer"},
	}, "query_ref", "status", "execution_mode", "html_fragment", "rendered_report", "report_provenance", "scope_note", "workspace_id", "role", "masked_fields", "truncated", "audit_event_key")
}

func agentDiagnosticsResultSchema() map[string]any {
	return objectSchema(map[string]any{
		"runtime_context_preview": freeObjectSchema(), "agent_runner": freeObjectSchema(), "skill_bindings": arraySchema(freeObjectSchema()), "effective_tools": arraySchema(freeObjectSchema()),
		"proposal_queue": arraySchema(schemaReference("AgentProposal")), "execution_logs": arraySchema(freeObjectSchema()), "execution_log_error": stringSchema(), "guardrails": arraySchema(stringSchema()),
	}, "runtime_context_preview", "agent_runner", "skill_bindings", "effective_tools", "proposal_queue", "execution_logs", "execution_log_error", "guardrails")
}

func diagnosticParameters() []any {
	parameters := []any{}
	for _, name := range []string{"workspace", "module_key", "view_key", "object_key", "record_id", "data_scope_note", "agent_mode", "run_mode", "risk_level"} {
		parameters = append(parameters, queryParameter(name, stringSchema()))
	}
	return parameters
}

func bearerSecurity() []any { return []any{map[string]any{"BearerAuth": []any{}}} }

func errorResponse(description string) map[string]any {
	return map[string]any{"description": description, "content": map[string]any{"application/json": map[string]any{"schema": schemaReference("Error")}}}
}

func pathParameter(name string) map[string]any {
	return map[string]any{"name": name, "in": "path", "required": true, "schema": nonEmptyStringSchema()}
}

func headerParameter(name string, required bool) map[string]any {
	return map[string]any{"name": name, "in": "header", "required": required, "schema": nonEmptyStringSchema()}
}

func queryParameter(name string, schema map[string]any) map[string]any {
	return map[string]any{"name": name, "in": "query", "required": false, "schema": schema}
}

func queryArrayParameter(name string) map[string]any {
	return map[string]any{"name": name, "in": "query", "required": false, "schema": arraySchema(stringSchema()), "style": "form", "explode": true}
}

func schemaReference(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}
func stringSchema() map[string]any         { return map[string]any{"type": "string"} }
func nonEmptyStringSchema() map[string]any { return map[string]any{"type": "string", "minLength": 1} }
func enumStringSchema(values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values}
}
func dateTimeSchema() map[string]any { return map[string]any{"type": "string", "format": "date-time"} }
func freeObjectSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true}
}
func arraySchema(items map[string]any) map[string]any {
	return map[string]any{"type": "array", "items": items}
}
func objectSchema(properties map[string]any, required ...string) map[string]any {
	value := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) != 0 {
		value["required"] = required
	}
	return value
}

func rawSchema(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func collectAgentHTTPReferences(value any, references map[string]bool) {
	switch typed := value.(type) {
	case map[string]any:
		if reference, ok := typed["$ref"].(string); ok {
			references[reference] = true
		}
		if security, ok := typed["security"].([]any); ok {
			for _, item := range security {
				if requirement, ok := item.(map[string]any); ok {
					for name := range requirement {
						references["#/components/securitySchemes/"+name] = true
					}
				}
			}
		}
		for _, child := range typed {
			collectAgentHTTPReferences(child, references)
		}
	case []any:
		for _, child := range typed {
			collectAgentHTTPReferences(child, references)
		}
	}
}

func agentHTTPReferenceParts(reference string) (string, string, bool) {
	const prefix = "#/components/"
	if !strings.HasPrefix(reference, prefix) {
		return "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(reference, prefix), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}
