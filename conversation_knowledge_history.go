package agentsdk

import "encoding/json"

// HistoricalRemoteKnowledgeTools recognizes previously persisted remote-source
// receipts. These definitions are never advertised or executable. Local parsed
// sources and private attachment calls must be rejected before this comparison.
func HistoricalRemoteKnowledgeTools() []ConversationToolDefinition {
	extract := LegacyKnowledgeExtractionTool()
	extract.Version = "2"
	var schema map[string]any
	_ = json.Unmarshal(extract.InputSchema, &schema)
	schema["properties"].(map[string]any)["attachment_id"] = map[string]any{"type": "string", "pattern": "^att_[a-f0-9]{32}$"}
	schema["required"] = []string{}
	schema["oneOf"] = []any{map[string]any{"required": []string{"doc_id"}, "not": map[string]any{"required": []string{"attachment_id"}}}, map[string]any{"required": []string{"attachment_id"}, "not": map[string]any{"anyOf": []any{map[string]any{"required": []string{"doc_id"}}, map[string]any{"required": []string{"library_id"}}}}}}
	extract.InputSchema, _ = json.Marshal(schema)
	extract.Description += " For semantic text fields, constrain the capture to the requested value's shape (for example an email address must contain an address, not any word after the label). Labels, headings, instructions and notes saying a value is absent are not that value. Inspect returned values against the requested meaning; correct a pattern that captures such a note. Keep required fields in the plan even when absent so missing is returned. Type=text validates text, not its business meaning."
	extract.Description += " For a private file, first discover it with knowledge_attachments and supply attachment_id instead of doc_id/library_id. The server binds it to the current conversation; do not submit conversation or owner identifiers."
	compact := extract
	compact.Version = "3"
	compact.Description += " Local original extraction uses all available parsed blocks on the server, even when knowledge_read returned only one bounded page. Do not fetch every page just to run a known extraction rule. Source metadata is a compact reference; extracted values and their quotes carry the evidence. It is not a search index or a calculation engine."
	var read ConversationToolDefinition
	for _, d := range KnowledgeConversationTools() {
		if d.Key == "knowledge_read" {
			read = d
		}
	}
	read.Version = "3"
	read.Description = "Read an authorized document or a private file in the CURRENT conversation. Supply either doc_id (and library_id for library documents), OR attachment_id returned by knowledge_attachments. Never combine locators. The server supplies the conversation and owner. This grants no access to arbitrary paths or URLs. Inspect actual parsed text before knowledge_extract, preserve source positions and warnings, and do not infer full-file completeness from a successful read."
	read.InputSchema = json.RawMessage(`{"type":"object","properties":{"doc_id":{"type":"string","minLength":1,"maxLength":4096},"library_id":{"type":"string","minLength":1,"maxLength":96},"attachment_id":{"type":"string","pattern":"^att_[a-f0-9]{32}$"}},"oneOf":[{"required":["doc_id"],"not":{"required":["attachment_id"]}},{"required":["attachment_id"],"not":{"anyOf":[{"required":["doc_id"]},{"required":["library_id"]}]}}],"additionalProperties":false}`)
	page := read
	page.Version = "4"
	schema = nil
	_ = json.Unmarshal(page.InputSchema, &schema)
	schema["properties"].(map[string]any)["after"] = map[string]any{"type": "string", "maxLength": 512}
	schema["properties"].(map[string]any)["limit"] = map[string]any{"type": "integer", "minimum": 1, "maximum": 50}
	page.InputSchema, _ = json.Marshal(schema)
	page.Description += " Locally parsed originals return a bounded page of text parts, source metadata and next_after. after must come from that document's returned cursor; limit (1..50, default 20) bounds parts, not original file pages. A long row/cell may span parts: offset/end are UTF-8 byte positions within block_id. Complete means traversal of parsed blocks ended, not that every feature of the original file was decoded. Read only portions needed for the user's task; do not traverse a large file by default. knowledge_extract processes available parsed content on the server without requiring you to read every page. For questions across documents use knowledge_search; for totals use structured analysis/calculation. Remote sources may not support these paging arguments."
	return []ConversationToolDefinition{extract, compact, read, page}
}
