package agentmodel

import (
	"encoding/json"
	"strings"
	"testing"
)

func knowledgeReceiptFixture() AgentTaskKnowledgeEvidence {
	return AgentTaskKnowledgeEvidence{LibraryID: "library-a", Operation: "search", ScopeSHA256: strings.Repeat("a", 64), Citations: []AgentTaskKnowledgeCitationEvidence{{ID: "citation-a", Provider: "knowledge", KBID: "kb-a", DocumentID: "doc-a", ChunkID: "chunk-a", DocumentVersion: 4, ContentSHA256: strings.Repeat("b", 64), SourceRevision: "generation-2"}}}
}
func TestTaskKnowledgeEvidenceIsBoundedAndContainsReferencesOnly(t *testing.T) {
	receipt := knowledgeReceiptFixture()
	if err := receipt.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["citations"].([]any)[0].(map[string]any)["document_version"] != float64(4) {
		t.Fatal("version is not numeric", string(raw))
	}
	for _, key := range []string{"excerpt", "title", "url", "authority", "credential", "permissions", "data"} {
		if strings.Contains(string(raw), `"`+key+`"`) {
			t.Fatal("content or authority leaked", key)
		}
	}
	copy := receipt.Clone()
	copy.Citations[0].DocumentID = "other"
	if receipt.Citations[0].DocumentID != "doc-a" {
		t.Fatal("clone aliases immutable references")
	}
	receipt.Citations = nil
	if err = receipt.Validate(); err != nil {
		t.Fatal("real no-match rejected", err)
	}
	raw, _ = json.Marshal(receipt.Clone())
	if !strings.Contains(string(raw), `"citations":[]`) {
		t.Fatal("no-match projection is not explicit", string(raw))
	}
}
func TestTaskKnowledgeEvidenceRejectsMalformedAndOversizedReferences(t *testing.T) {
	for _, change := range []func(*AgentTaskKnowledgeEvidence){
		func(e *AgentTaskKnowledgeEvidence) { e.LibraryID = "" },
		func(e *AgentTaskKnowledgeEvidence) { e.ScopeSHA256 = "not-a-scope" },
		func(e *AgentTaskKnowledgeEvidence) { e.Operation = "write" },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations = append(e.Citations, e.Citations[0]) },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations[0].DocumentID = "" },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations[0].DocumentVersion = -1 },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations[0].ContentSHA256 = "malformed" },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations[0].ID = strings.Repeat("a", 256) },
		func(e *AgentTaskKnowledgeEvidence) { e.Citations = make([]AgentTaskKnowledgeCitationEvidence, 21) },
	} {
		e := knowledgeReceiptFixture()
		change(&e)
		if e.Validate() == nil {
			t.Fatal("invalid Knowledge receipt accepted", e)
		}
	}
}
