package agentsdk

import (
	"context"
	"time"
)

// Conversation is the durable chat capability, with an optional execution port. It is independent of
// InteractiveRunner (one-shot business routing) and TaskRunner (background work).
const CapabilityConversationV1 = "conversation.v1"
const CapabilityConversationStreamV1 = "conversation.stream.v1"

type ConversationAuthority struct {
	Known       bool   `json:"known"`
	RuntimeID   string `json:"runtime_id"`
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	RoleKey     string `json:"role_key,omitempty"`
}

type Conversation struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	RuntimeID     string    `json:"runtime_id"`
	WorkspaceID   string    `json:"workspace_id"`
	UserID        string    `json:"user_id"`
	Archived      bool      `json:"archived"`
	MemoryEnabled bool      `json:"memory_enabled"`
	LastSeq       int64     `json:"last_seq"`
	ActiveRunID   string    `json:"active_run_id,omitempty"`
	SummaryID     string    `json:"summary_id,omitempty"`
	Revision      int64     `json:"revision"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ConversationMessage struct {
	Citations      []ConversationCitation `json:"citations,omitempty"`    // current, authorized read projection
	AccessError    string                 `json:"access_error,omitempty"` // read projection; original content is retained internally
	InteractionID  string                 `json:"interaction_id,omitempty"`
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	RunID          string                 `json:"run_id"`
	Seq            int64                  `json:"seq"`
	Role           string                 `json:"role"`
	Content        string                 `json:"content"`
	CreatedAt      time.Time              `json:"created_at"`
}

type ConversationRun struct {
	AccessError        string `json:"access_error,omitempty"` // source data is withheld from this projection
	ID                 string `json:"id"`
	ConversationID     string `json:"conversation_id"`
	ClientMessageID    string `json:"client_message_id"`
	RequestHash        string `json:"-"`
	Status             string `json:"status"` // queued, running, completed, failed, cancelled
	UserSeq            int64  `json:"user_seq"`
	AssistantMessageID string `json:"assistant_message_id,omitempty"`
	Attempt            int    `json:"attempt"`
	// Draft belongs to Attempt and is never included in conversation history.
	DraftText    string                   `json:"draft_text,omitempty"`
	DraftBytes   int                      `json:"draft_bytes"`
	LastEventSeq int64                    `json:"last_event_seq"`
	Steps        []ConversationStepView   `json:"steps,omitempty"`
	Interaction  *ConversationInteraction `json:"interaction,omitempty"`
	LastInputSeq int64                    `json:"last_input_seq,omitempty"`
	WriteScope   *ConversationWriteScope  `json:"write_scope,omitempty"`
	Model        string                   `json:"model,omitempty"`
	Usage        map[string]any           `json:"usage,omitempty"`
	ErrorCode    string                   `json:"error_code,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

func (r ConversationRun) Terminal() bool {
	return r.Status == "completed" || r.Status == "failed" || r.Status == "cancelled"
}
func (r ConversationRun) Waiting() bool {
	return r.Status == "waiting_user" || r.Status == "waiting_confirmation" || r.Status == "needs_reconciliation"
}

type ConversationEvent struct {
	RunID     string         `json:"run_id"`
	Seq       int64          `json:"seq"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// The summary is a replaceable context projection. Original messages remain the
// source of truth, and through_seq is a complete-turn boundary.
type ConversationSummary struct {
	Sources        *ConversationSources       `json:"sources,omitempty"`
	Rebuild        bool                       `json:"rebuild,omitempty"` // trusted worker rebuilds from original messages after source access changes
	ID             string                     `json:"id"`
	ConversationID string                     `json:"conversation_id"`
	PreviousID     string                     `json:"previous_id,omitempty"`
	ThroughSeq     int64                      `json:"through_seq"`
	Content        ConversationSummaryContent `json:"content"`
	SourceHash     string                     `json:"source_hash"`
	Model          string                     `json:"model"`
	Version        int                        `json:"version"`
	CreatedAt      time.Time                  `json:"created_at"`
}

type ConversationSummaryContent struct {
	Goal        string   `json:"goal"`
	Constraints []string `json:"constraints"`
	Facts       []string `json:"facts"`
	Decisions   []string `json:"decisions"`
	OpenItems   []string `json:"open_items"`
}

// Memories are explicitly authored preferences, never automatically extracted
// by a tool-free conversation model.
type ConversationMemory struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Enabled   bool      `json:"enabled"`
	Revision  int64     `json:"revision"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ConversationCreate struct {
	ClientID      string `json:"client_id"`
	Title         string `json:"title,omitempty"`
	MemoryEnabled bool   `json:"memory_enabled"`
}
type ConversationUpdate struct {
	ExpectedRevision int64   `json:"expected_revision"`
	Title            *string `json:"title,omitempty"`
	Archived         *bool   `json:"archived,omitempty"`
	MemoryEnabled    *bool   `json:"memory_enabled,omitempty"`
}
type ConversationSend struct {
	ClientMessageID string                  `json:"client_message_id"`
	Message         string                  `json:"message"`
	WriteScope      *ConversationWriteScope `json:"write_scope,omitempty"`
}

// A scope is explicitly submitted by the authenticated user and frozen on one
// run. It never grants Identity permissions or access to another owner's data.
// PersonalMemory covers creating, updating, disabling and deleting personal
// memories for this request; it does not authorize unrelated business effects.
type ConversationWriteScope struct {
	PersonalMemory    bool `json:"personal_memory"`
	PersonalTodos     bool `json:"personal_todos,omitempty"`
	PersonalArtifacts bool `json:"personal_artifacts,omitempty"`
}

// Grants are resource-specific. A memory grant must never authorize a todo or
// an unrelated future write tool merely because it is registered as personal.
func (s *ConversationWriteScope) Allows(tool string) bool {
	if s == nil {
		return false
	}
	switch tool {
	case "memory_save", "memory_forget":
		return s.PersonalMemory
	case "todo_create", "todo_update", "todo_delete":
		return s.PersonalTodos
	case "artifact_create", "artifact_edit", "artifact_export":
		return s.PersonalArtifacts
	default:
		return false
	}
}

type ConversationQuery struct {
	Search          string `json:"search,omitempty"`
	IncludeArchived bool   `json:"include_archived,omitempty"`
	BeforeID        string `json:"before_id,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}
type ConversationMessageQuery struct {
	BeforeSeq int64 `json:"before_seq,omitempty"`
	AfterSeq  int64 `json:"after_seq,omitempty"`
	Limit     int   `json:"limit,omitempty"`
}
type ConversationPage struct {
	Items      []Conversation `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
}
type ConversationMessagePage struct {
	Items         []ConversationMessage `json:"items"`
	NextBeforeSeq int64                 `json:"next_before_seq,omitempty"`
	NextAfterSeq  int64                 `json:"next_after_seq,omitempty"`
}
type ConversationEventPage struct {
	Items    []ConversationEvent `json:"items"`
	NextSeq  int64               `json:"next_seq"`
	Terminal bool                `json:"terminal"`
}
type ConversationMemoryWrite struct {
	ID               string `json:"id,omitempty"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	Enabled          bool   `json:"enabled"`
	ExpectedRevision int64  `json:"expected_revision"`
}

type ConversationService interface {
	Create(context.Context, ConversationCreate, ConversationAuthority) (Conversation, error)
	List(context.Context, ConversationQuery, ConversationAuthority) (ConversationPage, error)
	Get(context.Context, string, ConversationAuthority) (Conversation, error)
	Update(context.Context, string, ConversationUpdate, ConversationAuthority) (Conversation, error)
	Delete(context.Context, string, int64, ConversationAuthority) error
	Send(context.Context, string, ConversationSend, ConversationAuthority) (ConversationRun, error)
	Messages(context.Context, string, ConversationMessageQuery, ConversationAuthority) (ConversationMessagePage, error)
	Run(context.Context, string, string, ConversationAuthority) (ConversationRun, error)
	Events(context.Context, string, string, int64, int, ConversationAuthority) (ConversationEventPage, error)
	Cancel(context.Context, string, string, ConversationAuthority) (ConversationRun, error)
	Resume(context.Context, string, string, ConversationAuthority) (ConversationRun, error)
	Memories(context.Context, ConversationAuthority) ([]ConversationMemory, error)
	WriteMemory(context.Context, ConversationMemoryWrite, ConversationAuthority) (ConversationMemory, error)
	DeleteMemory(context.Context, string, int64, ConversationAuthority) error
}

// Optional extension: existing Binding implementations need not implement it.
type ConversationBinding interface{ Conversations() ConversationService }

type ConversationModelMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ConversationModelRequest struct {
	Sources        *ConversationSources       `json:"sources,omitempty"` // server-only; providers receive Messages
	Messages       []ConversationModelMessage `json:"messages"`
	Purpose        string                     `json:"purpose"` // reply or summary; neither permits tools
	IdempotencyKey string                     `json:"idempotency_key"`
	MaxOutputBytes int                        `json:"max_output_bytes"`
}
type ConversationModelResult struct {
	Content string         `json:"content"`
	Model   string         `json:"model,omitempty"`
	Usage   map[string]any `json:"usage,omitempty"`
}
type ConversationModel interface {
	GenerateConversation(context.Context, ConversationModelRequest) (ConversationModelResult, error)
}

// ConversationFactory lets a host discover explicit conversation configuration
// before opening modules. Persistent conversations do not require a legacy
// Agent, Task, Skill or entrypoint definition in the business manifest.
type ConversationFactory interface {
	ConversationEnabled() bool
}

// Optional extension for reply generation. The callback is synchronous and
// sequential. Implementations must stop on cancellation or callback error and
// return the exact concatenation of accepted deltas only on normal completion.
// Summarization continues to use GenerateConversation.
type ConversationStreamingModel interface {
	StreamConversation(context.Context, ConversationModelRequest, func(string) error) (ConversationModelResult, error)
}

// Readiness does not call a paid model endpoint. It checks local configuration,
// persistence and worker lifecycle; generation failures remain run failures.
type ConversationStatusProvider interface {
	ConversationReady(context.Context) error
	ConversationStreaming() bool
}
