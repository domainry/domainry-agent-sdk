package agentsdk

import (
	"bytes"
	"testing"
)

func TestCollaborationToolsPreserveSourceReferencesAndOwnTheirSchemas(t *testing.T) {
	first := ConversationCollaborationTools()
	if len(first) != 7 {
		t.Fatal("incomplete peer protocol")
	}
	for _, tool := range first {
		if bytes.Contains(tool.InputSchema, []byte(`"client_id"`)) || bytes.Contains(tool.InputSchema, []byte(`"ToolRequest"`)) {
			t.Fatalf("server identity exposed: %s", tool.Key)
		}
		if tool.Key == "delegation_update" && !bytes.Contains(tool.InputSchema, []byte(`"conversation_id"`)) {
			t.Fatal("delivery evidence lost its conversation reference")
		}
	}
	original := append([]byte(nil), first[1].InputSchema...)
	first[1].InputSchema[0] = '!'
	first[1].Description = "changed"
	next := ConversationCollaborationTools()
	if !bytes.Equal(next[1].InputSchema, original) || next[1].Description == "changed" {
		t.Fatal("caller mutated the trusted cached protocol")
	}
	for _, call := range []ConversationToolCall{{Name: "agent_delegate"}, {Name: "delegation_update", Arguments: `{"update":{"action":"accept_delivery"}}`}, {Name: "delegation_update", Arguments: `{"update":{"action":"resume"}}`}} {
		if ConversationPeerCommunication(call) {
			t.Fatal("management action gained communication scope", call)
		}
	}
	if !ConversationPeerCommunication(ConversationToolCall{Name: "agent_message"}) || !ConversationPeerCommunication(ConversationToolCall{Name: "delegation_update", Arguments: `{"update":{"action":"deliver"}}`}) {
		t.Fatal("accepted communication requires unrelated management scope")
	}
}
