package agentsdk

// Locations and text spans supplied by the knowledge provider, not decoded by Agent.
type DocumentLocation struct {
	Sheet string `json:"sheet,omitempty"`
	Row   int    `json:"row,omitempty"`
	Cell  string `json:"cell,omitempty"`
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
