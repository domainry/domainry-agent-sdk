package agentsdk

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestHistoricalKnowledgeDefinitionsMatchPreRemovalReceipts(t *testing.T) {
	// Captured definitions only, from the pre-removal synthetic live acceptance DB.
	// No conversation content, resource identifiers, or credentials are fixtures.
	data, err := os.ReadFile("testdata/knowledge_history_v3_v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var previous []ConversationToolDefinition
	if err = json.Unmarshal(data, &previous); err != nil {
		t.Fatal(err)
	}
	if len(previous) != 2 {
		t.Fatal("incomplete pre-removal receipt definitions")
	}
	for _, saved := range previous {
		found := false
		for _, current := range HistoricalRemoteKnowledgeTools() {
			if saved.Key != current.Key || saved.Version != current.Version {
				continue
			}
			found = true
			before, _ := json.Marshal(saved)
			after, _ := json.Marshal(current)
			if !bytes.Equal(before, after) {
				t.Errorf("historical %s v%s definition differs from persisted pre-removal bytes", saved.Key, saved.Version)
			}
		}
		if !found {
			t.Fatal("historical definition lost", saved.Key, saved.Version)
		}
	}
}
