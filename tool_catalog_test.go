package agentsdk

import (
	"slices"
	"testing"
)

func TestAgentToolCatalogIsStableAndDefensive(t *testing.T) {
	keys := AgentToolKeys()
	want := []string{"callConnector", "createRecord", "deleteRecord", "get_record", "invoke_action", "knowledge_read", "knowledge_search", "query_records", "readRecord", "sendEmail", "sendMessage", "updateRecord"}
	if !slices.Equal(keys, want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	keys[0] = "mutated"
	if slices.Equal(AgentToolKeys(), keys) {
		t.Fatal("AgentToolKeys returned mutable catalog storage")
	}
	tool, found := LookupAgentTool(" invoke_action ")
	if !found || !tool.Writes || !tool.RequiresAllowedObjects || !tool.RequiresAllowedActions {
		t.Fatalf("invoke_action metadata = %+v, found=%v", tool, found)
	}
}

func TestKnowledgeTaskToolsAreExplicitReadCapabilitiesNotObjectOrActionGrants(t *testing.T) {
	for _, key := range []string{AgentToolKnowledgeSearch, AgentToolKnowledgeRead} {
		tool, ok := LookupAgentTool(key)
		if !ok || tool.Key != key || tool.Writes || tool.RequiresAllowedObjects || tool.RequiresAllowedActions {
			t.Fatalf("Knowledge read reused project write or Object authority: %+v", tool)
		}
	}
	if _, ok := LookupAgentTool("knowledge_delete"); ok {
		t.Fatal("read registration implicitly enabled a Knowledge write")
	}
}
