package agentmodel

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// AgentTaskKnowledgeEvidence is a receipt from a trusted Task host, not model
// output or an access grant. It deliberately excludes text, URLs, authority and
// provider configuration. Current Knowledge authorization is required to read
// any referenced source later.
type AgentTaskKnowledgeEvidence struct {
	LibraryID   string                               `json:"library_id"`
	Operation   string                               `json:"operation"`
	ScopeSHA256 string                               `json:"scope_sha256"`
	Citations   []AgentTaskKnowledgeCitationEvidence `json:"citations"`
}

type AgentTaskKnowledgeCitationEvidence struct {
	ID              string `json:"id"`
	Provider        string `json:"provider"`
	KBID            string `json:"kb_id,omitempty"`
	DocumentID      string `json:"document_id"`
	RecordID        string `json:"record_id,omitempty"`
	ChunkID         string `json:"chunk_id,omitempty"`
	DocumentVersion int64  `json:"document_version,omitempty"`
	PageID          string `json:"page_id,omitempty"`
	PageVersion     int64  `json:"page_version,omitempty"`
	ContentSHA256   string `json:"content_sha256,omitempty"`
	SourceRevision  string `json:"source_revision,omitempty"`
}

func (e AgentTaskKnowledgeEvidence) Validate() error {
	invalid := func() error { return fmt.Errorf("invalid Agent Task Knowledge receipt") }
	identifier := func(s string, required bool, max int) bool {
		return (!required || s != "") && len(s) <= max && strings.TrimSpace(s) == s && !strings.ContainsAny(s, "\x00\r\n")
	}
	sha := func(s string) bool {
		b, err := hex.DecodeString(s)
		return err == nil && len(b) == 32 && strings.ToLower(s) == s
	}
	if !identifier(e.LibraryID, true, 96) || (e.Operation != "search" && e.Operation != "fetch") || !sha(e.ScopeSHA256) || len(e.Citations) > 20 {
		return invalid()
	}
	seen := map[string]bool{}
	for _, c := range e.Citations {
		if !identifier(c.ID, true, 255) || seen[c.ID] || !identifier(c.Provider, true, 96) || !identifier(c.KBID, false, 255) || !identifier(c.DocumentID, true, 255) || !identifier(c.RecordID, false, 255) || !identifier(c.ChunkID, false, 255) || !identifier(c.PageID, false, 255) || !identifier(c.SourceRevision, false, 255) || c.DocumentVersion < 0 || c.PageVersion < 0 || c.ContentSHA256 != "" && !sha(c.ContentSHA256) {
			return invalid()
		}
		seen[c.ID] = true
	}
	return nil
}

func (e *AgentTaskKnowledgeEvidence) Clone() *AgentTaskKnowledgeEvidence {
	if e == nil {
		return nil
	}
	copy := *e
	copy.Citations = append([]AgentTaskKnowledgeCitationEvidence{}, e.Citations...)
	return &copy
}
