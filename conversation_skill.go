package agentsdk

import (
	"context"
	"encoding/json"
	"time"
)

const AgentCapabilitySkills = "agent.skills"

type ConversationSkillResourceSummary struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MediaType   string `json:"media_type"`
}

type ConversationSkillSummary struct {
	Key           string                             `json:"key"`
	Version       string                             `json:"version"`
	Name          string                             `json:"name"`
	Description   string                             `json:"description"`
	AllowedTools  []string                           `json:"allowed_tools"`
	InputSchema   json.RawMessage                    `json:"input_schema,omitempty"`
	OutputSchema  json.RawMessage                    `json:"output_schema,omitempty"`
	Resources     []ConversationSkillResourceSummary `json:"resources,omitempty"`
	WorkflowSteps int                                `json:"workflow_steps"`
}

type ConversationSkillLoadRequest struct {
	Key         string `json:"key"`
	Version     string `json:"version"`
	ResourceKey string `json:"resource_key,omitempty"`
}

type ConversationSkillLoadResult struct {
	Summary      ConversationSkillSummary `json:"summary"`
	Instructions string                   `json:"instructions,omitempty"`
	Workflow     []SkillWorkflowStep      `json:"workflow,omitempty"`
	Resource     *SkillResource           `json:"resource,omitempty"`
}

type ConversationSkillVersion struct {
	Definition   SkillSchema `json:"definition"`
	Digest       string      `json:"digest"`
	State        string      `json:"state"` // draft, evaluated, published, retired
	Revision     int64       `json:"revision"`
	EvaluationID string      `json:"evaluation_id,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	PublishedAt  *time.Time  `json:"published_at,omitempty"`
}

type ConversationSkillPage struct {
	Items    []ConversationSkillSummary `json:"items"`
	Complete bool                       `json:"complete"`
}

func ConversationSkillSummaryFor(skill SkillSchema) ConversationSkillSummary {
	out := ConversationSkillSummary{
		Key: skill.Key, Version: skill.Version, Name: skill.Name, Description: skill.Description,
		AllowedTools: append([]string{}, skill.AllowedTools...),
		InputSchema:  append(json.RawMessage(nil), skill.InputSchema...), OutputSchema: append(json.RawMessage(nil), skill.OutputSchema...),
		WorkflowSteps: len(skill.Workflow),
	}
	for _, resource := range skill.Resources {
		out.Resources = append(out.Resources, ConversationSkillResourceSummary{Key: resource.Key, Name: resource.Name, Description: resource.Description, MediaType: resource.MediaType})
	}
	return out
}

// Feedback records an observed task outcome against the exact immutable Agent
// and Skill configuration used by that task. It does not edit configuration.
type ConversationCapabilityFeedbackCreate struct {
	ClientID        string   `json:"client_id"`
	TaskID          string   `json:"task_id"`
	ArtifactID      string   `json:"artifact_id,omitempty"`
	ArtifactVersion int64    `json:"artifact_version,omitempty"`
	Outcome         string   `json:"outcome"` // adopted, revised, failed
	Reason          string   `json:"reason"`
	Uncertainty     string   `json:"uncertainty,omitempty"`
	SkillKeys       []string `json:"skill_keys,omitempty"`
}

type ConversationCapabilityFeedback struct {
	ID                 string            `json:"id"`
	TaskID             string            `json:"task_id"`
	RunID              string            `json:"run_id"`
	ArtifactID         string            `json:"artifact_id,omitempty"`
	ArtifactVersion    int64             `json:"artifact_version,omitempty"`
	Outcome            string            `json:"outcome"`
	Reason             string            `json:"reason"`
	Uncertainty        string            `json:"uncertainty,omitempty"`
	AgentID            string            `json:"agent_id"`
	AgentRevision      int64             `json:"agent_revision"`
	AgentPromptVersion string            `json:"agent_prompt_version"`
	TargetSkillKeys    []string          `json:"target_skill_keys,omitempty"`
	SkillVersions      map[string]string `json:"skill_versions,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}

type ConversationImprovementCandidateCreate struct {
	ClientID    string          `json:"client_id"`
	Kind        string          `json:"kind"` // skill, agent_prompt, delegation_strategy
	TargetKey   string          `json:"target_key"`
	Version     string          `json:"version"`
	FeedbackIDs []string        `json:"feedback_ids"`
	Proposal    json.RawMessage `json:"proposal"`
	Reason      string          `json:"reason"`
}

type ConversationImprovementCandidate struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	TargetKey       string `json:"target_key"`
	Version         string `json:"version"`
	BaselineVersion string `json:"baseline_version,omitempty"`
	// BaselineProposal freezes the exact configuration compared by the candidate.
	// It lets publication archive the first built-in/deployed version so rollback
	// does not depend on that version having been a previous managed candidate.
	BaselineProposal json.RawMessage `json:"baseline_proposal,omitempty"`
	// BaselineSnapshot marks a server-created immutable version archive. It may
	// become active through rollback but is never created or edited by API input.
	BaselineSnapshot bool                               `json:"baseline_snapshot,omitempty"`
	FeedbackIDs      []string                           `json:"feedback_ids"`
	Proposal         json.RawMessage                    `json:"proposal"`
	Reason           string                             `json:"reason"`
	Status           string                             `json:"status"` // candidate, evaluated, published, retired, rejected, rolled_back
	Evaluation       *ConversationImprovementEvaluation `json:"evaluation,omitempty"`
	Revision         int64                              `json:"revision"`
	CreatedAt        time.Time                          `json:"created_at"`
	UpdatedAt        time.Time                          `json:"updated_at"`
}

type ConversationImprovementCandidatePage struct {
	Items    []ConversationImprovementCandidate `json:"items"`
	Complete bool                               `json:"complete"`
}

type ConversationImprovementEvaluationWrite struct {
	ClientID           string   `json:"client_id"`
	ExpectedRevision   int64    `json:"expected_revision"`
	SuiteVersion       string   `json:"suite_version"`
	ScenarioIDs        []string `json:"scenario_ids"`
	BaselineCompleted  int      `json:"baseline_completed"`
	CandidateCompleted int      `json:"candidate_completed"`
	BaselineOmissions  int      `json:"baseline_omissions"`
	CandidateOmissions int      `json:"candidate_omissions"`
	Regressions        []string `json:"regressions,omitempty"`
	Passed             bool     `json:"passed"`
	Notes              string   `json:"notes,omitempty"`
}

type ConversationImprovementEvaluation struct {
	ID                 string    `json:"id"`
	SuiteVersion       string    `json:"suite_version"`
	ScenarioIDs        []string  `json:"scenario_ids"`
	BaselineCompleted  int       `json:"baseline_completed"`
	CandidateCompleted int       `json:"candidate_completed"`
	BaselineOmissions  int       `json:"baseline_omissions"`
	CandidateOmissions int       `json:"candidate_omissions"`
	Regressions        []string  `json:"regressions,omitempty"`
	Passed             bool      `json:"passed"`
	Notes              string    `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type ConversationImprovementPublish struct {
	ClientID         string `json:"client_id"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type ConversationImprovementRollback struct {
	ClientID         string `json:"client_id"`
	ExpectedRevision int64  `json:"expected_revision"`
	TargetVersion    string `json:"target_version"`
	Reason           string `json:"reason"`
}

type ConversationCapabilityConfiguration struct {
	Kind        string          `json:"kind"`
	TargetKey   string          `json:"target_key"`
	Version     string          `json:"version"`
	CandidateID string          `json:"candidate_id"`
	Revision    int64           `json:"revision"`
	Payload     json.RawMessage `json:"payload"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ConversationSkillService interface {
	ConversationSkills(context.Context, ConversationAuthority) (ConversationSkillPage, error)
	ConversationSkill(context.Context, string, string, ConversationAuthority) (ConversationSkillVersion, error)
	ConversationSkillResource(context.Context, string, string, string, ConversationAuthority) (SkillResource, error)
	CreateConversationCapabilityFeedback(context.Context, ConversationCapabilityFeedbackCreate, ConversationAuthority) (ConversationCapabilityFeedback, error)
	ConversationImprovementCandidates(context.Context, ConversationAuthority) (ConversationImprovementCandidatePage, error)
	CreateConversationImprovementCandidate(context.Context, ConversationImprovementCandidateCreate, ConversationAuthority) (ConversationImprovementCandidate, error)
	EvaluateConversationImprovementCandidate(context.Context, string, ConversationImprovementEvaluationWrite, ConversationAuthority) (ConversationImprovementCandidate, error)
	PublishConversationImprovementCandidate(context.Context, string, ConversationImprovementPublish, ConversationAuthority) (ConversationImprovementCandidate, error)
	RollbackConversationImprovement(context.Context, string, ConversationImprovementRollback, ConversationAuthority) (ConversationImprovementCandidate, error)
}

func ConversationSkillLoadTool() ConversationToolDefinition {
	return ConversationToolDefinition{
		Key: "skill_load", Version: "1", ActionKey: ConversationToolActionPrefix + "skill_load", Effect: "read", Idempotency: "natural",
		Description:   "Load the exact instructions and reusable workflow for a Skill listed in the current Agent prompt. Supply the listed key and version. Omit resource_key for instructions and workflow; provide one listed resource_key to load that resource. A Skill never grants tools: use only tools already present in this run.",
		InputSchema:   json.RawMessage(`{"type":"object","properties":{"key":{"type":"string","minLength":1,"maxLength":96,"pattern":"^[A-Za-z0-9_.:-]+$"},"version":{"type":"string","minLength":1,"maxLength":128},"resource_key":{"type":"string","maxLength":96,"pattern":"^[A-Za-z0-9_.:-]*$"}},"required":["key","version"],"additionalProperties":false}`),
		OutputSchema:  json.RawMessage(`{"type":"object","properties":{"summary":{"type":"object"},"instructions":{"type":"string"},"workflow":{"type":"array","items":{"type":"object"}},"resource":{"type":"object"}},"required":["summary"],"additionalProperties":false}`),
		TimeoutMillis: 10000, MaxOutputBytes: 131072,
	}
}
