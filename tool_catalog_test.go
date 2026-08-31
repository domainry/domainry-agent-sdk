package agentsdk

import (
	"slices"
	"testing"
)

func TestAgentToolCatalogIsStableAndDefensive(t *testing.T) {
	keys := AgentToolKeys()
	want := []string{"callConnector", "createRecord", "deleteRecord", "get_record", "invoke_action", "query_records", "readRecord", "sendEmail", "sendMessage", "updateRecord"}
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
