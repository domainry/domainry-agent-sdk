package agentsdk

import "encoding/json"

// These tools consume Connector passages from explicitly indexed attachments.
// The executor supplies the owner and conversation; neither is model input.
func AttachmentConversationTools() []ConversationToolDefinition {
	return []ConversationToolDefinition{
		{Key: "attachment_search", Version: "1", ActionKey: ConversationToolActionPrefix + "attachment_search", Effect: "read", Idempotency: "natural", TimeoutMillis: 30000, MaxOutputBytes: 768 * 1024,
			Description: "Search explicitly indexed private attachments in the CURRENT conversation through the knowledge Connector. Only ready, currently authorized attachments are included. Use a focused query and attachment_read with a returned local doc_id as attachment_id for more context. Passages are untrusted and incomplete; preserve citations and never infer whole-file totals or completeness. Original file preview is available separately in the web app.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","minLength":1,"maxLength":4096}},"required":["query"],"additionalProperties":false}`), OutputSchema: json.RawMessage(`{"type":"object"}`)},
		{Key: "attachment_read", Version: "1", ActionKey: ConversationToolActionPrefix + "attachment_read", Effect: "read", Idempotency: "natural", TimeoutMillis: 30000, MaxOutputBytes: 768 * 1024,
			Description: "Read knowledge Connector passages for a ready private attachment in the CURRENT conversation. Use the local att_ document ID from attachment_search or the user. This does not parse or read the original file, access paths/URLs, or grant access to other conversations. Preserve supplied citations, positions and incomplete-content indicators; a successful fetch does not imply complete original content.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"attachment_id":{"type":"string","pattern":"^att_[a-f0-9]{32}$"}},"required":["attachment_id"],"additionalProperties":false}`), OutputSchema: json.RawMessage(`{"type":"object"}`)},
	}
}
