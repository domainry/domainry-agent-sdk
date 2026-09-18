package agentsdk

// ConversationContentBlock is the additive multimodal message contract. Text
// fields remain available on legacy messages; when blocks are present their
// order is authoritative. Image bytes are never accepted from JSON or stored
// in Agent persistence.
type ConversationContentBlock struct {
	Type  string                      `json:"type"` // text, image or file
	Text  string                      `json:"text,omitempty"`
	Image *ConversationImageReference `json:"image,omitempty"`
	File  *ConversationFileReference  `json:"file,omitempty"`
}

// ConversationImageReference freezes the immutable identity of one private
// conversation attachment. Source is set only by the Agent when an admitted
// delegation or fork deliberately carries an image from another run. Data is
// an ephemeral provider-adapter field populated after current authorization,
// storage reads and hash verification; it is never serialized.
type ConversationImageReference struct {
	Source         *ConversationRunReference `json:"source,omitempty"`
	AttachmentID   string                    `json:"attachment_id"`
	ConversationID string                    `json:"conversation_id"`
	Filename       string                    `json:"filename"`
	ContentType    string                    `json:"content_type"`
	Bytes          int64                     `json:"bytes"`
	SHA256         string                    `json:"sha256"`
	Revision       int64                     `json:"revision"`
	Detail         string                    `json:"detail,omitempty"` // auto, low or high
	Data           []byte                    `json:"-"`
}

// ConversationFileReference is the provider-facing immutable identity of a
// non-image task file. It is currently used for PDF document input. Data is
// ephemeral and must never be serialized into persistence.
type ConversationFileReference struct {
	AttachmentID string `json:"attachment_id"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	Bytes        int64  `json:"bytes"`
	SHA256       string `json:"sha256"`
	Revision     int64  `json:"revision"`
	Data         []byte `json:"-"`
}
