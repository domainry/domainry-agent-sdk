// Package agentsdk defines the deployment-neutral Agent execution protocol.
// Agent owns definitions, conversations, proposals and execution state;
// Runtime retains workflow state plus host authorization and business effects.
package agentsdk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/domainry/domainry-foundation/modulecapability"
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

func (e *Error) ErrorCode() string {
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.Code)
}

func (*Error) ErrorParams() map[string]string { return nil }

const ProtocolVersionV1 = "domainry-agent-protocol-v1"

const (
	CapabilityTaskStart        = "task.start"
	CapabilityTaskPoll         = "task.poll"
	CapabilityTaskCancel       = "task.cancel"
	CapabilityInteractiveRun   = "interactive.run"
	CapabilityLifecycleExecute = "lifecycle.execute"
)

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
	required := map[string]bool{
		CapabilityTaskStart:      false,
		CapabilityTaskPoll:       false,
		CapabilityTaskCancel:     false,
		CapabilityInteractiveRun: false,
	}
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

func (d Descriptor) HasCapability(capability string) bool {
	capability = strings.TrimSpace(capability)
	for _, candidate := range d.Capabilities {
		if strings.TrimSpace(candidate) == capability {
			return true
		}
	}
	return false
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
	NodeInstanceID      string            `json:"node_instance_id,omitempty"`
	WorkspaceID         string            `json:"workspace_id"`
	Task                TaskDefinition    `json:"task"`
	Identity            ExecutionIdentity `json:"identity"`
	Input               map[string]any    `json:"input,omitempty"`
	AllowedObjects      []string          `json:"allowed_objects,omitempty"`
	AllowedActions      []string          `json:"allowed_actions,omitempty"`
	AllowedOutcomes     []string          `json:"allowed_outcomes,omitempty"`
	AllowedTools        []string          `json:"allowed_tools,omitempty"`
	ExecutionCredential string            `json:"execution_credential,omitempty"`
	CorrelationID       string            `json:"correlation_id,omitempty"`
	IdempotencyKey      string            `json:"idempotency_key"`
	MaxAttempts         int               `json:"max_attempts,omitempty"`
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

const AgentRuntimeServiceAudience = "domainry-runtime"

type authorizedServiceActionContextKey struct{}

type authorizedServiceAction struct {
	key      string
	audience string
}

// WithAuthorizedServiceAction carries one host-approved service Action into an
// Agent application call. Only trusted host or credential-verification assembly
// should create this evidence; it never carries a bundle of sibling grants.
func WithAuthorizedServiceAction(ctx context.Context, actionKey, audience string) context.Context {
	return context.WithValue(ctx, authorizedServiceActionContextKey{}, authorizedServiceAction{
		key: strings.TrimSpace(actionKey), audience: strings.TrimSpace(audience),
	})
}

// HasAuthorizedServiceAction accepts only the current exact Action and service
// audience. An authorization for Start cannot authorize Poll or Cancel.
func HasAuthorizedServiceAction(ctx context.Context, actionKey, audience string) bool {
	if ctx == nil {
		return false
	}
	evidence, ok := ctx.Value(authorizedServiceActionContextKey{}).(authorizedServiceAction)
	return ok && evidence.key != "" && evidence.key == strings.TrimSpace(actionKey) && evidence.audience == strings.TrimSpace(audience)
}

type RouteCandidate struct {
	RouteType string `json:"route_type"`
	TargetKey string `json:"target_key"`
	Version   string `json:"version,omitempty"`
}
type GlobalContext struct {
	ContractVersion     string             `json:"contract_version"`
	ContextRevision     string             `json:"context_revision"`
	EntrypointKey       string             `json:"entrypoint_key"`
	AgentKey            string             `json:"agent_key"`
	RouteKey            string             `json:"route_key"`
	ObjectKey           string             `json:"object_key,omitempty"`
	RecordID            string             `json:"record_id,omitempty"`
	SelectedRecordIDs   []string           `json:"selected_record_ids,omitempty"`
	Locale              string             `json:"locale,omitempty"`
	Timezone            string             `json:"timezone,omitempty"`
	Principal           PrincipalReference `json:"principal"`
	AllowedTaskKeys     []string           `json:"allowed_task_keys"`
	AllowedWorkflowKeys []string           `json:"allowed_workflow_keys"`
	AvailableOperations []string           `json:"available_operation_ids"`
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
	modulecapability.Binding
	Descriptor() Descriptor
	TaskRunner() TaskRunner
	InteractiveRunner() InteractiveRunner
	Close(context.Context) error
}
type Factory interface {
	Open(context.Context, ApplicationRef) (Binding, error)
}
