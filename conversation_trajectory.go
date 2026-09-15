package agentsdk

import (
	"context"
	"time"
)

const (
	ConversationTrajectoryVersion = 1

	ConversationReplayDisplay      = "display"
	ConversationReplayModelFixture = "model_fixture"
	ConversationReplayLiveRerun    = "live_rerun"
)

// ConversationForkOrigin is provenance only. It does not make the new
// conversation subordinate to the source and never grants source access.
type ConversationForkOrigin struct {
	ConversationID   string    `json:"conversation_id"`
	RunID            string    `json:"run_id"`
	BoundaryEventSeq int64     `json:"boundary_event_seq"`
	CreatedAt        time.Time `json:"created_at"`
}

type ConversationForkRequest struct {
	ClientID string `json:"client_id"`
	Title    string `json:"title,omitempty"`
	AgentID  string `json:"agent_id,omitempty"`
}

type ConversationTrajectoryMessage struct {
	Role          string                     `json:"role"`
	Content       string                     `json:"content"`
	ContentBlocks []ConversationContentBlock `json:"content_blocks,omitempty"`
	ToolCalls     []ConversationToolCall     `json:"tool_calls,omitempty"`
	ToolCallID    string                     `json:"tool_call_id,omitempty"`
	IsError       bool                       `json:"is_error,omitempty"`
}

type ConversationTrajectoryToolDefinition struct {
	Definition ConversationToolDefinition `json:"definition"`
	SHA256     string                     `json:"sha256"`
}

type ConversationTrajectoryRequest struct {
	Index           int                                    `json:"index"`
	Step            int                                    `json:"step"` // -1 is the initial reply request
	Purpose         string                                 `json:"purpose"`
	Model           ConversationModelIdentity              `json:"model"`
	ReasoningEffort string                                 `json:"reasoning_effort,omitempty"`
	Messages        []ConversationTrajectoryMessage        `json:"messages"`
	Tools           []ConversationTrajectoryToolDefinition `json:"tools,omitempty"`
	Context         *ConversationContextView               `json:"context,omitempty"`
	SHA256          string                                 `json:"sha256"`
}

type ConversationTrajectoryResponse struct {
	RequestIndex int                           `json:"request_index"`
	Step         int                           `json:"step"`
	FinishReason string                        `json:"finish_reason"`
	Model        string                        `json:"model,omitempty"`
	Message      ConversationTrajectoryMessage `json:"message"`
	Usage        map[string]any                `json:"usage,omitempty"`
	SHA256       string                        `json:"sha256"`
}

type ConversationTrajectoryTool struct {
	ParentCallID  string                               `json:"parent_call_id,omitempty"`
	DispatchIndex int                                  `json:"dispatch_index,omitempty"`
	Step          int                                  `json:"step"`
	Call          ConversationToolCall                 `json:"call"`
	Definition    ConversationTrajectoryToolDefinition `json:"definition"`
	State         string                               `json:"state"`
	Result        *ConversationToolResult              `json:"result,omitempty"`
	SHA256        string                               `json:"sha256"`
}

// ConversationTrajectory is an authorized, detached projection. Provider
// continuation state, credentials and authorization evidence are never part of
// this contract. Reading it does not call a model or a tool.
type ConversationTrajectory struct {
	Version          int                              `json:"version"`
	Mode             string                           `json:"mode"`
	Source           ConversationRunReference         `json:"source"`
	BoundaryEventSeq int64                            `json:"boundary_event_seq"`
	RunStatus        string                           `json:"run_status"`
	Requests         []ConversationTrajectoryRequest  `json:"requests"`
	Responses        []ConversationTrajectoryResponse `json:"responses"`
	Tools            []ConversationTrajectoryTool     `json:"tools"`
	SHA256           string                           `json:"sha256"`
	RecordedAt       time.Time                        `json:"recorded_at"`
}

type ConversationTrajectoryExport struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SHA256      string `json:"sha256"`
	Data        []byte `json:"data"`
}

type ConversationTrajectoryReplayRequest struct {
	Mode string                   `json:"mode"`
	Fork *ConversationForkRequest `json:"fork,omitempty"`
}

type ConversationTrajectoryReplay struct {
	Mode              string                           `json:"mode"`
	Source            ConversationRunReference         `json:"source"`
	TrajectorySHA256  string                           `json:"trajectory_sha256"`
	ConsumedRequests  int                              `json:"consumed_requests"`
	RecordedResponses []ConversationTrajectoryResponse `json:"recorded_responses,omitempty"`
	RecordedTools     []ConversationTrajectoryTool     `json:"recorded_tools,omitempty"`
	EffectsExecuted   bool                             `json:"effects_executed"`
	ReadyForInput     bool                             `json:"ready_for_input,omitempty"`
	Fork              *Conversation                    `json:"fork,omitempty"`
}

type ConversationTrajectoryCompareRequest struct {
	Other ConversationRunReference `json:"other"`
}

type ConversationTrajectoryDifference struct {
	Kind        string `json:"kind"`
	Index       int    `json:"index"`
	LeftSHA256  string `json:"left_sha256,omitempty"`
	RightSHA256 string `json:"right_sha256,omitempty"`
}

type ConversationTrajectoryComparison struct {
	Left        ConversationRunReference           `json:"left"`
	Right       ConversationRunReference           `json:"right"`
	LeftSHA256  string                             `json:"left_sha256"`
	RightSHA256 string                             `json:"right_sha256"`
	Equal       bool                               `json:"equal"`
	Differences []ConversationTrajectoryDifference `json:"differences"`
}

// ConversationTrajectoryService is optional so existing conversation bindings
// continue to implement only the base service.
type ConversationTrajectoryService interface {
	ForkConversation(context.Context, string, string, ConversationForkRequest, ConversationAuthority) (Conversation, error)
	ConversationTrajectory(context.Context, string, string, ConversationAuthority) (ConversationTrajectory, error)
	ExportConversationTrajectory(context.Context, string, string, ConversationAuthority) (ConversationTrajectoryExport, error)
	ReplayConversationTrajectory(context.Context, string, string, ConversationTrajectoryReplayRequest, ConversationAuthority) (ConversationTrajectoryReplay, error)
	CompareConversationTrajectories(context.Context, string, string, ConversationTrajectoryCompareRequest, ConversationAuthority) (ConversationTrajectoryComparison, error)
}
