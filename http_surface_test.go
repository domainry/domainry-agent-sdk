package agentsdk

import (
	"strings"
	"testing"

	actioncontract "github.com/domainry/domainry-foundation/action"
)

func TestAgentHTTPSurfaceContractOwnsCompleteRouteCatalog(t *testing.T) {
	actions, err := AgentAuthorizationActions()
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 25 {
		t.Fatalf("Agent Action count=%d want=25", len(actions))
	}
	roleActions, nonHTTPActions := 0, 0
	for _, action := range actions {
		if action.Permission != nil {
			roleActions++
			if action.Permission == nil || action.Permission.Key != action.Key || action.Permission.Owner != action.Owner {
				t.Fatalf("Agent role Action %q permission=%+v", action.Key, action.Permission)
			}
		}
		if action.HTTP == nil {
			nonHTTPActions++
			if len(action.NonHTTP) != 1 || action.Authorization.Strategy != actioncontract.AuthorizationSigned {
				t.Fatalf("Agent non-HTTP Action %q binding=%v authorization=%+v", action.Key, action.NonHTTP, action.Authorization)
			}
		}
		separator := strings.LastIndex(action.Key, ".")
		if separator <= 0 || action.OperationKey != action.Key[separator+1:] {
			t.Fatalf("Agent Action %q operation=%q", action.Key, action.OperationKey)
		}
	}
	if roleActions != 7 || nonHTTPActions != 3 {
		t.Fatalf("Agent role Actions=%d non-HTTP Actions=%d", roleActions, nonHTTPActions)
	}

	contract, err := CompileAgentHTTPSurfaceContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ContractVersion != AgentHTTPSurfaceContractVersion || contract.Owner != "agent" || len(contract.Routes) != 22 || len(contract.OpenAPI) != 22 {
		t.Fatalf("Agent HTTP contract=%s owner=%s routes=%d operations=%d", contract.ContractVersion, contract.Owner, len(contract.Routes), len(contract.OpenAPI))
	}
	seen := map[string]bool{}
	for _, route := range contract.Routes {
		pattern := route.Pattern()
		if seen[pattern] || contract.OpenAPI[pattern]["operationId"] == nil || contract.OpenAPI[pattern]["responses"] == nil {
			t.Fatalf("incomplete or duplicate Agent route %q", pattern)
		}
		seen[pattern] = true
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
	var toolAuthorization actioncontract.Authorization
	for _, route := range contract.Routes {
		if route.Pattern() == "POST /agent-dialog/task-tools/invoke" {
			toolAuthorization = route.Action.Authorization
			break
		}
	}
	if toolAuthorization.Strategy != actioncontract.AuthorizationSigned || toolAuthorization.PolicyKey != "agent.task_tool_credential" {
		t.Fatalf("Agent tool callback authorization=%+v", toolAuthorization)
	}
	components := HTTPSurfaceReferencedComponents(map[string]map[string]any{"POST /agent-dialog/runs": contract.OpenAPI["POST /agent-dialog/runs"]})
	if components["securitySchemes"]["BearerAuth"] == nil || components["schemas"]["AgentInteractiveRunRequest"] == nil || components["schemas"]["AgentInteractiveExecutionResult"] == nil || components["schemas"]["AgentInteractiveRun"] == nil {
		t.Fatalf("Agent referenced component closure=%v", components)
	}
}
