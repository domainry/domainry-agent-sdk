// Package agentsdk defines the deployment-neutral Agent execution protocol.
// Runtime retains workflow, authorization, approval and task-run ownership.
package agentsdk

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Error struct {
	Class, Code, Message string
	Retryable            bool
	Cause                error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return strings.TrimSpace(e.Code)
}
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

const ProtocolVersionV1 = "domainry-agent-protocol-v1"

type DeploymentMode string

const (
	DeploymentModeModule DeploymentMode = "module"
	DeploymentModeSaaS   DeploymentMode = "saas"
)

type ApplicationRef struct {
	RuntimeID string `json:"runtime_id"`
}

func (r ApplicationRef) Validate() error {
	if strings.TrimSpace(r.RuntimeID) == "" {
		return fmt.Errorf("Agent runtime identity is required")
	}
	return nil
}

type Descriptor struct {
	ProtocolVersion string         `json:"protocol_version"`
	Mode            DeploymentMode `json:"mode"`
	Audience        string         `json:"audience,omitempty"`
	Capabilities    []string       `json:"capabilities"`
}

func (d Descriptor) Validate() error {
	if d.ProtocolVersion != ProtocolVersionV1 {
		return fmt.Errorf("unsupported Agent protocol %q", d.ProtocolVersion)
	}
	if d.Mode != DeploymentModeModule && d.Mode != DeploymentModeSaaS {
		return fmt.Errorf("invalid Agent deployment mode %q", d.Mode)
	}
	required := map[string]bool{"task.start": false, "task.poll": false, "task.cancel": false, "interactive.run": false}
	for _, capability := range d.Capabilities {
		if _, ok := required[strings.TrimSpace(capability)]; ok {
			required[strings.TrimSpace(capability)] = true
		}
	}
	for capability, present := range required {
		if !present {
			return fmt.Errorf("Agent capability %q is required", capability)
		}
	}
	return nil
}

type TaskDefinition struct {
	Key         string `json:"key"`
	Version     string `json:"version"`
	Instruction string `json:"instruction"`
}
type PrincipalReference struct {
	UserID                string `json:"user_id"`
	RoleKey               string `json:"role_key"`
	WorkspaceID           string `json:"workspace_id"`
	AuthorizationRevision string `json:"authorization_revision"`
}
type ExecutionIdentity struct {
	Mode                   string             `json:"mode"`
	Initiator              PrincipalReference `json:"initiator"`
	Execution              PrincipalReference `json:"execution"`
	ServicePrincipalKey    string             `json:"service_principal_key,omitempty"`
	ServiceRotationVersion int                `json:"service_rotation_version,omitempty"`
}

type ProviderRunStatus string

const (
	ProviderRunAccepted  ProviderRunStatus = "accepted"
	ProviderRunRunning   ProviderRunStatus = "running"
	ProviderRunCompleted ProviderRunStatus = "completed"
	ProviderRunFailed    ProviderRunStatus = "failed"
	ProviderRunCancelled ProviderRunStatus = "cancelled"
	ProviderRunUnknown   ProviderRunStatus = "unknown"
)

type TaskRequest struct {
	TaskRunID           string            `json:"task_run_id"`
	ProcessID           string            `json:"process_id,omitempty"`
	WorkspaceID         string            `json:"workspace_id"`
	Task                TaskDefinition    `json:"task"`
	Identity            ExecutionIdentity `json:"identity"`
	Input               map[string]any    `json:"input,omitempty"`
	ExecutionCredential string            `json:"execution_credential,omitempty"`
	CorrelationID       string            `json:"correlation_id,omitempty"`
	IdempotencyKey      string            `json:"idempotency_key"`
	Deadline            time.Time         `json:"deadline"`
}

type TaskResult struct {
	ExternalRunID string            `json:"external_run_id,omitempty"`
	Status        ProviderRunStatus `json:"status"`
	Outcome       string            `json:"outcome,omitempty"`
	Output        map[string]any    `json:"output,omitempty"`
	RawEvidence   []byte            `json:"raw_evidence,omitempty"`
	Model         string            `json:"model,omitempty"`
	Usage         map[string]any    `json:"usage,omitempty"`
	ErrorClass    string            `json:"error_class,omitempty"`
	ErrorCode     string            `json:"error_code,omitempty"`
	Retryable     bool              `json:"retryable,omitempty"`
}

type TaskRunner interface {
	Start(context.Context, TaskRequest) (TaskResult, error)
	Poll(context.Context, string, string) (TaskResult, error)
	Cancel(context.Context, string, string) (TaskResult, error)
}

type RouteCandidate struct {
	RouteType string `json:"route_type"`
	TargetKey string `json:"target_key"`
	Version   string `json:"version,omitempty"`
}
type GlobalContext struct {
	WorkspaceID string         `json:"workspace_id"`
	ActorID     string         `json:"actor_id"`
	Locale      string         `json:"locale,omitempty"`
	Timezone    string         `json:"timezone,omitempty"`
	Values      map[string]any `json:"values,omitempty"`
}
type InteractiveRequest struct {
	RunID               string           `json:"run_id"`
	SessionID           string           `json:"session_id"`
	Context             GlobalContext    `json:"context"`
	Message             string           `json:"message"`
	ExecutionCredential string           `json:"execution_credential,omitempty"`
	IdempotencyKey      string           `json:"idempotency_key"`
	Candidates          []RouteCandidate `json:"candidates,omitempty"`
	MaxSteps            int              `json:"max_steps,omitempty"`
	MaxToolCalls        int              `json:"max_tool_calls,omitempty"`
	Deadline            time.Time        `json:"deadline"`
}
type RouteResult struct {
	RouteType      string         `json:"route_type"`
	TargetKey      string         `json:"target_key"`
	TargetVersion  string         `json:"target_version,omitempty"`
	Input          map[string]any `json:"input,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}
type Handoff struct {
	TaskKey     string         `json:"task_key"`
	TaskVersion string         `json:"task_version,omitempty"`
	TaskRunID   string         `json:"task_run_id,omitempty"`
	Input       map[string]any `json:"input,omitempty"`
}
type InteractiveResult struct {
	RunID         string         `json:"run_id,omitempty"`
	ExternalRunID string         `json:"external_run_id,omitempty"`
	Status        string         `json:"status"`
	Message       string         `json:"message,omitempty"`
	Structured    map[string]any `json:"structured,omitempty"`
	Model         string         `json:"model,omitempty"`
	Usage         map[string]any `json:"usage,omitempty"`
	Route         *RouteResult   `json:"route,omitempty"`
	Handoff       *Handoff       `json:"handoff,omitempty"`
	EvidenceRefs  []string       `json:"evidence_refs,omitempty"`
}
type InteractiveRunner interface {
	Run(context.Context, InteractiveRequest) (InteractiveResult, error)
}

type Binding interface {
	Descriptor() Descriptor
	TaskRunner() TaskRunner
	InteractiveRunner() InteractiveRunner
	Close(context.Context) error
}
type Factory interface {
	Open(context.Context, ApplicationRef) (Binding, error)
}
