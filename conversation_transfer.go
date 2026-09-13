package agentsdk

import "time"

type ConversationDelegationTransfer struct {
	AgentID       string `json:"agent_id"`
	RemainingWork string `json:"remaining_work"`
}

// An assignment is an immutable record of who took responsibility, separate
// from the stable delegation and from each execution run.
type ConversationDelegationAssignment struct {
	Number            int64                           `json:"number"`
	AgentID           string                          `json:"agent_id"`
	AgentRevision     int64                           `json:"agent_revision"`
	ConversationID    string                          `json:"conversation_id"`
	TaskID            string                          `json:"task_id"`
	AgreementRevision int64                           `json:"agreement_revision"`
	Reason            string                          `json:"reason"`
	ActorID           string                          `json:"actor_id"`
	Source            *ConversationRunReference       `json:"source,omitempty"`
	PreviousDelivery  *ConversationDelegationDelivery `json:"previous_delivery,omitempty"`
	CreatedAt         time.Time                       `json:"created_at"`
}

// Handoff facts are assembled from stored execution, never accepted from a
// model or a client. References remain subject to current source access.
type ConversationDelegationHandoff struct {
	Source        *ConversationRunReference      `json:"source,omitempty"`
	RemainingWork string                         `json:"remaining_work"`
	Runs          []ConversationRunReference     `json:"runs"`
	Effects       []ConversationDelegationEffect `json:"effects"`
}
type ConversationDelegationEffect struct {
	Tool       string                      `json:"tool"`
	Arguments  string                      `json:"arguments"`
	Status     string                      `json:"status"`
	Completion string                      `json:"completion,omitempty"`
	ResourceID string                      `json:"resource_id,omitempty"`
	Reference  ConversationResultReference `json:"reference"`
}
