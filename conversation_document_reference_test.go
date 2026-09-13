package agentsdk

import (
	"strings"
	"testing"
)

func TestSharedDocumentReferencesRejectPrivateAndAmbiguousSources(t *testing.T) {
	ref := ConversationDocumentReference{LibraryID: "lib_" + strings.Repeat("a", 32), DocumentID: "kdoc_" + strings.Repeat("b", 32), Revision: 2, SHA256: strings.Repeat("c", 64)}
	if !ValidConversationDocumentReferences(nil) || !ValidConversationDocumentReferences([]ConversationDocumentReference{ref}) {
		t.Fatal("valid reference rejected")
	}
	for _, refs := range [][]ConversationDocumentReference{
		{ref, ref},
		{{LibraryID: ref.LibraryID, DocumentID: "att_" + strings.Repeat("b", 32), Revision: 2, SHA256: ref.SHA256}},
		{{LibraryID: ref.LibraryID, DocumentID: ref.DocumentID, Revision: 0, SHA256: ref.SHA256}},
		{{LibraryID: ref.LibraryID, DocumentID: ref.DocumentID, Revision: 2, SHA256: "opaque"}},
		{{LibraryID: "other-owner/path", DocumentID: ref.DocumentID, Revision: 2, SHA256: ref.SHA256}},
		make([]ConversationDocumentReference, 5),
	} {
		if ValidConversationDocumentReferences(refs) {
			t.Fatalf("ambiguous source accepted: %+v", refs)
		}
	}
}
