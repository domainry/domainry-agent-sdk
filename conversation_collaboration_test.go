package agentsdk

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDependencySourceReaderKeepsSavedVersionOneAndExposesOnlyCurrentSchema(t *testing.T) {
	for _, version := range []string{"1", "2"} {
		definition, found := ConversationDelegationSourceReadDefinition(version)
		if !found || definition.Version != version {
			t.Fatal("missing source reader definition", version)
		}
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
			t.Fatal(err)
		}
		if _, present := schema.Properties["dependency_id"]; present != (version == "2") {
			t.Fatal("saved and current argument schemas were confused", version)
		}
		for _, required := range schema.Required {
			if required == "dependency_id" {
				t.Fatal("direct admitted source reading requires an upstream dependency")
			}
		}
		definition.InputSchema[0] = '!'
		fresh, _ := ConversationDelegationSourceReadDefinition(version)
		if !json.Valid(fresh.InputSchema) {
			t.Fatal("source reader definition mutation reached shared catalog", version)
		}
	}
	if _, found := ConversationDelegationSourceReadDefinition("unknown"); found {
		t.Fatal("unknown source reader version became trusted provenance")
	}
	for _, definition := range ConversationCollaborationTools() {
		if definition.Key == "delegation_source_read" && definition.Version != "2" {
			t.Fatal("legacy provenance definition was exposed for new model invocation")
		}
	}
}

func TestCollaborationToolsPreserveSourceReferencesAndOwnTheirSchemas(t *testing.T) {
	first := ConversationCollaborationTools()
	if len(first) != 11 {
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
