package agentsdk

import "regexp"

// A reference identifies an explicitly shared, immutable Knowledge library file.
// It carries no content or access grant. Private conversation attachments must
// first be explicitly copied through Knowledge's existing import workflow.
type ConversationDocumentReference struct {
	LibraryID  string `json:"library_id"`
	DocumentID string `json:"document_id"`
	Revision   int64  `json:"revision"`
	SHA256     string `json:"sha256"`
}

var sharedLibraryID = regexp.MustCompile(`^lib_[a-f0-9]{32}$`)
var sharedDocumentID = regexp.MustCompile(`^kdoc_[a-f0-9]{32}$`)
var sharedDocumentHash = regexp.MustCompile(`^[a-f0-9]{64}$`)

// ValidConversationDocumentReferences checks shape only; the Knowledge owner
// must separately authorize current access and verify every exact version.
func ValidConversationDocumentReferences(refs []ConversationDocumentReference) bool {
	if len(refs) > 4 {
		return false
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		key := ref.LibraryID + "/" + ref.DocumentID
		if !sharedLibraryID.MatchString(ref.LibraryID) || !sharedDocumentID.MatchString(ref.DocumentID) || ref.Revision < 1 || !sharedDocumentHash.MatchString(ref.SHA256) || seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}
