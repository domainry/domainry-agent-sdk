package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// ConversationTaskMutationRepository atomically saves task_start with its
// frozen tool call receipt and completion event.
type ConversationTaskMutationRepository interface {
	ApplyConversationTaskTool(context.Context, agentsdk.ConversationToolRequest, agentsdk.ConversationTask) (agentsdk.ConversationToolResult, error)
}

// ConversationTaskWorkerRepository is an Agent-owned queue boundary. Launch
// creates a normal conversation run so the existing worker, lease, execution
// ledger and current authorization checks remain the only execution path.
type ConversationTaskWorkerRepository interface {
	LaunchConversationTask(context.Context, string) (ConversationTaskLaunch, bool, error)
	ConversationTask(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationTask, error)
}

// ConversationTaskReadRepository returns owner-scoped executor records. The
// application layer reauthorizes and projects runs, results and artifacts.
type ConversationTaskReadRepository interface {
	ConversationTask(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationTask, error)
	ConversationTasks(context.Context, agentsdk.ConversationTaskQuery, agentsdk.ConversationAuthority) (ConversationTaskRecordPage, error)
}

type ConversationTaskRecordPage struct {
	Items      []agentsdk.ConversationTask
	NextCursor string
	Complete   bool
}

type ConversationTaskLaunch struct {
	Task      agentsdk.ConversationTask
	Run       agentsdk.ConversationRun
	Authority agentsdk.ConversationAuthority
}
