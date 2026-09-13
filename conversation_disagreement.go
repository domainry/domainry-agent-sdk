package agentsdk

import "time"

// Claims are observations, not instructions or proof of another Agent's opinion.
// Their actual submitter and immutable execution sources are assigned by the server.
type ConversationDisagreementClaimInput struct {
	Conclusion    string                        `json:"conclusion"`
	DataScope     string                        `json:"data_scope"`
	Period        string                        `json:"period"`
	SourceVersion string                        `json:"source_version"`
	Calculation   string                        `json:"calculation"`
	Receipts      []ConversationResultReference `json:"receipts,omitempty"`
}
type ConversationDisagreementActor struct {
	UserID  string                    `json:"user_id"`
	AgentID string                    `json:"agent_id,omitempty"`
	Source  *ConversationRunReference `json:"source,omitempty"`
}
type ConversationDisagreementClaim struct {
	ID string `json:"id"`
	ConversationDisagreementClaimInput
	Actor     ConversationDisagreementActor `json:"actor"`
	CreatedAt time.Time                     `json:"created_at"`
}
type ConversationEvidenceComparison struct {
	DataScope     string `json:"data_scope"`
	Period        string `json:"period"`
	SourceVersion string `json:"source_version"`
	Calculation   string `json:"calculation"`
}
type ConversationDisagreementDecision struct {
	Outcome           string                         `json:"outcome"` // inspect, revise, ask_user, adopt
	AdoptClaimID      string                         `json:"adopt_claim_id,omitempty"`
	Basis             string                         `json:"basis"`
	Comparison        ConversationEvidenceComparison `json:"comparison"`
	OwnerAgentID      string                         `json:"owner_agent_id,omitempty"`
	NextAction        string                         `json:"next_action,omitempty"`
	BriefVersion      int64                          `json:"brief_version"`
	AgreementRevision int64                          `json:"agreement_revision"`
	DeliveryDigest    string                         `json:"delivery_digest"`
}
type ConversationDisagreementChange struct {
	Operation        string                               `json:"operation"` // raise, add_claim, decide
	ID               string                               `json:"id,omitempty"`
	ExpectedRevision int64                                `json:"expected_revision,omitempty"`
	Title            string                               `json:"title,omitempty"`
	Condition        *int                                 `json:"condition,omitempty"`
	Claims           []ConversationDisagreementClaimInput `json:"claims,omitempty"`
	Decision         *ConversationDisagreementDecision    `json:"decision,omitempty"`
}
type ConversationDisagreementSummary struct {
	ID                string                     `json:"id"`
	Revision          int64                      `json:"revision"`
	Title             string                     `json:"title"`
	Condition         *int                       `json:"condition,omitempty"`
	Requirement       string                     `json:"requirement,omitempty"`
	Status            string                     `json:"status"`
	BriefVersion      int64                      `json:"brief_version"`
	AgreementRevision int64                      `json:"agreement_revision"`
	DeliveryDigest    string                     `json:"delivery_digest"`
	OwnerAgentID      string                     `json:"owner_agent_id,omitempty"`
	OwnerUserID       string                     `json:"owner_user_id,omitempty"`
	NextAction        string                     `json:"next_action,omitempty"`
	Sources           []ConversationRunReference `json:"sources,omitempty"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}
type ConversationDisagreement struct {
	DecisionActor *ConversationDisagreementActor `json:"decision_actor,omitempty"`
	ConversationDisagreementSummary
	Claims   []ConversationDisagreementClaim   `json:"claims"`
	Decision *ConversationDisagreementDecision `json:"decision,omitempty"`
	Actor    ConversationDisagreementActor     `json:"actor"`
	Event    string                            `json:"event"`
	Reason   string                            `json:"reason"`
}
type ConversationDisagreementHistory struct {
	Items      []ConversationDisagreement `json:"items"`
	Complete   bool                       `json:"complete"`
	NextBefore int64                      `json:"next_before,omitempty"`
}
type ConversationDisagreementRead struct {
	ID             string `json:"id"`
	DisagreementID string `json:"disagreement_id"`
	BeforeRevision int64  `json:"before_revision,omitempty"`
}
