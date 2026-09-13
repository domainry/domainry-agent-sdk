package agentsdk

// A participant grant applies only to its recorded delegation. Identity's
// current role policy is still required; it never transfers tool permissions.
// The first supported scopes are view and communicate. Other scopes require
// their execution/receipt sharing contracts before they can be granted.
type ConversationDelegationParticipant struct {
	UserID     string   `json:"user_id"`
	Operations []string `json:"operations"`
	Revision   int64    `json:"revision,omitempty"`
}

// Clients choose recipients and scopes, not the server's grant revision.
type ConversationDelegationParticipantInput struct {
	UserID     string   `json:"user_id"`
	Operations []string `json:"operations"`
}

func ConversationParticipantOperations() []string { return []string{"view", "communicate"} }
