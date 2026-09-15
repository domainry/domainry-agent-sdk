package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// ConversationImprovementRepository owns user-scoped observations, candidate
// versions, evaluations and the active configuration pointer. It never applies
// a configuration or executes a historical business effect.
type ConversationImprovementRepository interface {
	PublishedConversationSkills(context.Context, agentsdk.ConversationAuthority) ([]agentsdk.ConversationSkillVersion, error)
	ConversationSkillVersion(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationSkillVersion, error)
	ConversationCapabilityConfiguration(context.Context, string, string, agentsdk.ConversationAuthority) (agentsdk.ConversationCapabilityConfiguration, bool, error)
	CreateConversationCapabilityFeedback(context.Context, string, agentsdk.ConversationCapabilityFeedback, agentsdk.ConversationCapabilityFeedbackCreate, agentsdk.ConversationAuthority) (agentsdk.ConversationCapabilityFeedback, error)
	ConversationCapabilityFeedbacks(context.Context, []string, agentsdk.ConversationAuthority) ([]agentsdk.ConversationCapabilityFeedback, error)
	CreateConversationImprovementCandidate(context.Context, string, agentsdk.ConversationImprovementCandidate, agentsdk.ConversationImprovementCandidateCreate, agentsdk.ConversationAuthority) (agentsdk.ConversationImprovementCandidate, error)
	ConversationImprovementCandidate(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationImprovementCandidate, error)
	ConversationImprovementCandidates(context.Context, agentsdk.ConversationAuthority) ([]agentsdk.ConversationImprovementCandidate, error)
	EvaluateConversationImprovementCandidate(context.Context, string, agentsdk.ConversationImprovementEvaluationWrite, agentsdk.ConversationAuthority) (agentsdk.ConversationImprovementCandidate, error)
	PublishConversationImprovementCandidate(context.Context, string, agentsdk.ConversationImprovementPublish, agentsdk.ConversationAuthority) (agentsdk.ConversationImprovementCandidate, error)
	RollbackConversationImprovement(context.Context, string, agentsdk.ConversationImprovementRollback, agentsdk.ConversationAuthority) (agentsdk.ConversationImprovementCandidate, error)
}
