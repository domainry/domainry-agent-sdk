package agentsdk

import "testing"

func TestAgentHTTPSurfaceContractOwnsCompleteRouteCatalog(t *testing.T) {
	contract := AgentHTTPSurfaceContract()
	if contract.ContractVersion != AgentHTTPSurfaceContractVersion || contract.Owner != "agent" || len(contract.Routes) != 22 || len(contract.OpenAPI) != 22 {
		t.Fatalf("Agent HTTP contract=%s owner=%s routes=%d operations=%d", contract.ContractVersion, contract.Owner, len(contract.Routes), len(contract.OpenAPI))
	}
	seen := map[string]bool{}
	for _, route := range contract.Routes {
		if seen[route.Pattern] || contract.OpenAPI[route.Pattern]["operationId"] == nil || contract.OpenAPI[route.Pattern]["responses"] == nil {
			t.Fatalf("incomplete or duplicate Agent route %q", route.Pattern)
		}
		seen[route.Pattern] = true
	}
	stream := contract.OpenAPI["POST /agent-dialog/runs/stream"]
	if stream["requestBody"] == nil || stream["parameters"] == nil || stream["x-domainry-runtime-client-method"] != "runAgentStream" {
		t.Fatalf("Agent stream operation=%#v", stream)
	}
	tool := contract.OpenAPI["POST /agent-dialog/task-tools/invoke"]
	security, ok := tool["security"].([]any)
	if !ok || len(security) != 0 {
		t.Fatalf("Agent tool callback security=%#v", tool["security"])
	}
	components := HTTPSurfaceReferencedComponents(map[string]map[string]any{"POST /agent-dialog/runs": contract.OpenAPI["POST /agent-dialog/runs"]})
	if components["securitySchemes"]["BearerAuth"] == nil || components["schemas"]["AgentInteractiveRunRequest"] == nil || components["schemas"]["AgentInteractiveExecutionResult"] == nil || components["schemas"]["AgentInteractiveRun"] == nil {
		t.Fatalf("Agent referenced component closure=%v", components)
	}
}
