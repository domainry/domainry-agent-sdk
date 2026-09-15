package agentsdk

// Tool requests carry references, never an execution user or role. The host
// assigns the publication idempotency key from the recorded tool invocation.
type ConversationDelegationExecutionPublish struct {
	ID          string                     `json:"id"`
	Publication ConversationExecutionShare `json:"publication"`
}

type ConversationDelegationExecutionRead struct {
	ID        string                   `json:"id"`
	Reference ConversationRunReference `json:"reference"`
}

type ConversationDelegationExecutionResultRead struct {
	ID   string                 `json:"id"`
	Read ConversationResultRead `json:"read"`
}

type ConversationDelegationExecutionIndex struct {
	DelegationID string                             `json:"delegation_id"`
	Publications []ConversationExecutionPublication `json:"publications"`
}

type ConversationDelegationExecutionView struct {
	DelegationID string          `json:"delegation_id"`
	Run          ConversationRun `json:"run"`
}

type ConversationDelegationExecutionResult struct {
	DelegationID string                  `json:"delegation_id"`
	Result       ConversationResultSlice `json:"result"`
}
