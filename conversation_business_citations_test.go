package agentsdk

import (
	"encoding/json"
	"strings"
	"testing"

	toolsdk "github.com/domainry/domainry-tools-sdk"
)

func TestConversationBusinessRecordCitationsBindEvidenceAndStayBounded(t *testing.T) {
	record := ConversationBusinessRecord{ID: "opportunity-1", Version: "v1", Data: map[string]json.RawMessage{
		"name": json.RawMessage(`"续约商机"`), "notes": json.RawMessage(`"` + strings.Repeat("证", 2048) + `"`),
	}}
	first := ConversationBusinessRecordCitations(" source ", "get_record", "opportunity", []ConversationBusinessRecord{record}, 1)
	if len(first) != 1 || first[0].ID == "" || first[0].Source != "source" || first[0].Title != "续约商机" || len(first[0].Excerpt) > conversationBusinessCitationExcerptBytes || !strings.HasPrefix(first[0].ID, "business_") {
		t.Fatalf("citation=%#v", first)
	}
	record.Version = "v2"
	second := ConversationBusinessRecordCitations("source", "get_record", "opportunity", []ConversationBusinessRecord{record}, 1)
	if second[0].ID == first[0].ID {
		t.Fatal("citation identity did not bind the source version")
	}
	if !ConversationCitationsEqual(nil, []toolsdk.Citation{}) || ConversationCitationsEqual(first, second) {
		t.Fatal("citation equality did not preserve exact ordered evidence")
	}
}
