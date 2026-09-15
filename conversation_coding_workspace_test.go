package agentsdk

import "testing"

func TestConversationCodingToolContractsSeparateReadsAndWrites(t *testing.T) {
	tools := ConversationCodingTools()
	if len(tools) != 12 {
		t.Fatalf("coding tools = %d", len(tools))
	}
	seen := map[string]bool{}
	for _, tool := range tools {
		if seen[tool.Key] || !IsConversationCodingTool(tool.Key) || tool.ActionKey != ConversationToolActionPrefix+tool.Key || tool.Version != "1" {
			t.Fatalf("invalid coding tool contract: %+v", tool)
		}
		seen[tool.Key] = true
		if tool.Effect == "write" && tool.Idempotency != "reconcile" {
			t.Fatalf("mutation %s lacks recovery contract", tool.Key)
		}
		if tool.Effect == "read" && tool.Idempotency != "natural" {
			t.Fatalf("read %s is not natural", tool.Key)
		}
	}
	for _, required := range []string{"coding_file_read", "coding_file_search", "coding_file_write", "coding_file_edit", "coding_terminal_open", "coding_terminal_send", "coding_terminal_read", "coding_terminal_close", "coding_process_start", "coding_process_read", "coding_process_kill", "coding_lsp"} {
		if !seen[required] {
			t.Fatal("missing coding capability", required)
		}
	}
}
