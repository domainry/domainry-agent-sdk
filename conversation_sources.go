package agentsdk

// Server-owned provenance. Models never author this structure. Runs names the
// owner-scoped source runs whose protected results contributed to this input or
// summary, independently of whether the model retained their identifiers.
type ConversationSources struct {
	Version int                        `json:"version"`
	Runs    []ConversationRunReference `json:"runs"`
	Omitted []ConversationRunReference `json:"omitted,omitempty"` // summary can be rebuilt if access is restored
}

type ConversationRunReference struct {
	ConversationID string `json:"conversation_id"`
	RunID          string `json:"run_id"`
	// Zero audits the complete run. Positive values use one-based step labels:
	// 1 includes only the original input; 2 also includes internal step 0, etc.
	// Server-authored artifact sources stop before their generating step.
	BeforeStep int `json:"before_step,omitempty"`
}
