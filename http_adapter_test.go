package agentsdk

import (
	"strings"
	"testing"

	actioncontract "github.com/domainry/domainry-foundation/action"
)

func TestAgentHTTPAdapterContractOwnsCompleteRouteCatalog(t *testing.T) {
	actions, err := AgentAuthorizationActions()
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 140+len(ConversationToolActions()) {
		t.Fatalf("Agent Action count=%d", len(actions))
	}
	roleActions, nonHTTPActions := 0, 0
	for _, action := range actions {
		if (action.Key == ConversationActionPrefix+"external_agent_assignments" || action.Key == ConversationActionPrefix+"result_read" || action.Key == ConversationActionPrefix+"trajectory_compare" || action.Key == ConversationActionPrefix+"delegations_result" || action.Key == ConversationActionPrefix+"delegations_publication" || action.Key == ConversationActionPrefix+"delegations_publications" || action.Key == ConversationActionPrefix+"delegations_contract_publication" || action.Key == ConversationActionPrefix+"delegations_contract_candidates" || action.Key == ConversationActionPrefix+"delegations_contract_publications" || action.Key == ConversationActionPrefix+"delegations_artifact" || action.Key == ConversationActionPrefix+"delegations_export") && action.EffectClass != actioncontract.EffectRead {
			t.Fatal("stored-result read was classified as a mutation")
		}
		if action.Permission != nil {
			roleActions++
			if action.Permission.Key != action.Key || action.Permission.Owner != action.Owner {
				t.Fatalf("Agent role Action %q permission=%+v", action.Key, action.Permission)
			}
		}
		if action.HTTP == nil {
			nonHTTPActions++
			strategy := actioncontract.AuthorizationSigned
			if strings.HasPrefix(action.Key, ConversationToolActionPrefix) || strings.HasPrefix(action.Key, ConversationCollaborationPermissionPrefix) {
				strategy = actioncontract.AuthorizationAuthenticated
			}
			if len(action.NonHTTP) != 1 || action.Authorization.Strategy != strategy {
				t.Fatalf("Agent non-HTTP Action %q binding=%v authorization=%+v", action.Key, action.NonHTTP, action.Authorization)
			}
		}
		separator := strings.LastIndex(action.Key, ".")
		if separator <= 0 || action.OperationKey != action.Key[separator+1:] {
			t.Fatalf("Agent Action %q operation=%q", action.Key, action.OperationKey)
		}
	}
	if roleActions != 41+len(ConversationToolActions()) || nonHTTPActions != 13+len(ConversationToolActions()) {
		t.Fatalf("Agent role Actions=%d non-HTTP Actions=%d tool Actions=%d", roleActions, nonHTTPActions, len(ConversationToolActions()))
	}

	contract, err := CompileAgentHTTPAdapterContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.ContractVersion != AgentHTTPAdapterContractVersion || contract.Owner != "agent" || len(contract.Routes) != 127 {
		t.Fatalf("Agent HTTP contract=%s owner=%s routes=%d", contract.ContractVersion, contract.Owner, len(contract.Routes))
	}
	seen := map[string]bool{}
	foundFork, foundExport, foundProvenance := false, false, false
	var toolAuthorization actioncontract.Authorization
	for _, route := range contract.Routes {
		pattern := route.Pattern()
		if pattern == "" || seen[pattern] {
			t.Fatalf("incomplete or duplicate Agent route %q", pattern)
		}
		seen[pattern] = true
		switch pattern {
		case "POST /agent/conversations/{conversationID}/runs/{runID}/forks":
			foundFork = true
		case "POST /agent/delegations/{delegationID}/delivery-export":
			foundExport = true
		case "POST /agent/conversation-provenance":
			foundProvenance = true
		case "POST /agent/task-tools/invoke":
			toolAuthorization = route.Action.Authorization
		}
	}
	if !foundFork || !foundExport || !foundProvenance {
		t.Fatalf("Agent typed route catalog is missing fork=%t export=%t provenance=%t", foundFork, foundExport, foundProvenance)
	}
	if toolAuthorization.Strategy != actioncontract.AuthorizationSigned || toolAuthorization.PolicyKey != "agent.task_tool_credential" {
		t.Fatalf("Agent tool callback authorization=%+v", toolAuthorization)
	}
}
