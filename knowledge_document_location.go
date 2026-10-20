package agentsdk

// Locations and text spans supplied by the knowledge provider, not decoded by Agent.
type DocumentLocation struct {
	Page      int    `json:"page,omitempty"`
	Slide     int    `json:"slide,omitempty"`
	Section   string `json:"section,omitempty"`
	ChunkID   string `json:"chunk_id,omitempty"`
	Sheet     string `json:"sheet,omitempty"`
	Row       int    `json:"row,omitempty"`
	RowEnd    int    `json:"row_end,omitempty"`
	Cell      string `json:"cell,omitempty"`
	Line      int    `json:"line,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`
	Paragraph int    `json:"paragraph,omitempty"`
	JSONPath  string `json:"json_path,omitempty"`
	HTMLID    string `json:"html_id,omitempty"`
	StartMS   int64  `json:"start_ms,omitempty"`
	EndMS     int64  `json:"end_ms,omitempty"`
	MediaKind string `json:"media_kind,omitempty"`
}

// A cell range maps the exact text in its block to an actual file cell. A
// formula's cached value is file data, not a freshly evaluated result.
type KnowledgePassageCell struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Raw     string `json:"raw"`
	Formula string `json:"formula,omitempty"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}
