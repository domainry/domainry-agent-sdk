package agentsdk

import "testing"

func TestExternalAgentProtocolRoutesAreExplicitAndSeparateFromDelivery(t *testing.T) {
	routes := map[string]ConversationHTTPDefinition{}
	for _, route := range ConversationHTTPDefinitions() {
		routes[route.Operation] = route
	}
	for operation, pattern := range map[string]string{
		"external_agent_assignments": "POST /agent/external-agents/assignments/query",
		"external_agent_claim":       "POST /agent/external-agents/tasks/{taskID}/claim",
		"external_agent_report":      "POST /agent/external-agents/tasks/{taskID}/reports",
	} {
		if routes[operation].Pattern != pattern {
			t.Fatalf("%s route=%q", operation, routes[operation].Pattern)
		}
	}
	if _, ok := routes["delegations_update"]; !ok {
		t.Fatal("external protocol removed the canonical delivery and review route")
	}
	config := ConversationExternalAgentConfig{Protocol: ConversationExternalAgentProtocolV1, Version: "1", Capabilities: ConversationExternalAgentCapabilities{Steering: true, Cancellation: true, Resume: true, StructuredOutput: true, ExecutionDetails: false}}
	execution := ConversationExternalAgentExecution{Protocol: config.Protocol, Version: config.Version, Status: "running", Capabilities: config.Capabilities, DetailsAvailable: false, Events: []ConversationExternalAgentEvent{}, EventsComplete: true}
	if execution.DetailsAvailable || execution.Capabilities.ExecutionDetails {
		t.Fatal("missing detail capability was promoted to execution evidence")
	}
}
