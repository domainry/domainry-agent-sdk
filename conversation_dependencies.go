package agentsdk

import (
	"encoding/json"
	"time"
)

// Dependencies link work, never Agent identities. Empty Fields means all task
// agreement fields. Version selects the exact upstream agreement being used.
type ConversationDependencyInput struct {
	AgreementRevision int64    `json:"agreement_revision,omitempty"`
	DelegationID      string   `json:"delegation_id"`
	BriefVersion      int64    `json:"brief_version"`
	Fields            []string `json:"fields,omitempty"`
}

type ConversationTaskDependency struct {
	InputSource *ConversationRunReference `json:"input_source,omitempty"`
	ConversationDependencyInput
	Digest string                    `json:"digest"`
	Values json.RawMessage           `json:"values"`
	Source *ConversationRunReference `json:"source,omitempty"`
}

type ConversationRequirementChange struct {
	SourceAgreementRevision int64     `json:"source_agreement_revision"`
	ID                      string    `json:"id"`
	SourceDelegationID      string    `json:"source_delegation_id"`
	SourceBriefVersion      int64     `json:"source_brief_version"`
	ViaDelegationID         string    `json:"via_delegation_id,omitempty"`
	ChangedFields           []string  `json:"changed_fields"`
	CreatedAt               time.Time `json:"created_at"`
}

// Each agreement revision is immutable. Execution adoption is recorded when
// the recipient's first frozen model step commits, separately from resume.
type ConversationAgreementRevision struct {
	StructuredInput *ConversationStructuredInput `json:"structured_input,omitempty"`
	InputSource     *ConversationRunReference    `json:"input_source,omitempty"`
	Revision        int64                        `json:"revision"`
	Brief           ConversationTaskBrief        `json:"brief"`
	Dependencies    []ConversationTaskDependency `json:"dependencies"`
	ChangedFields   []string                     `json:"changed_fields"`
	Reason          string                       `json:"reason"`
	FromUserID      string                       `json:"from_user_id,omitempty"`
	FromAgentID     string                       `json:"from_agent_id,omitempty"`
	Source          *ConversationRunReference    `json:"source,omitempty"`
	// ChangeSource traces the decision and reason; Source remains the brief's source.
	ChangeSource *ConversationRunReference `json:"change_source,omitempty"`
	CreatedAt    time.Time                 `json:"created_at"`
}

type ConversationAgreementHistory struct {
	Items      []ConversationAgreementRevision `json:"items"`
	NextBefore int64                           `json:"next_before,omitempty"`
	Complete   bool                            `json:"complete"`
}

type ConversationDependencyState struct {
	DelegationID             string `json:"delegation_id"`
	CurrentBriefVersion      int64  `json:"current_brief_version"`
	CurrentAgreementRevision int64  `json:"current_agreement_revision"`
	State                    string `json:"state"` // current, changed, needs_review
}
