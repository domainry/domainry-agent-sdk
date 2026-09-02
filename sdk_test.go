package agentsdk

import (
	"context"
	"testing"
)

func TestDescriptorAndApplicationValidation(t *testing.T) {
	if err := (ApplicationRef{}).Validate(); err == nil {
		t.Fatal("empty application accepted")
	}
	for _, mode := range []DeploymentMode{DeploymentModeModule, DeploymentModeSaaS} {
		descriptor := Descriptor{ProtocolVersion: ProtocolVersionV1, Mode: mode, Capabilities: []string{CapabilityTaskStart, CapabilityTaskPoll, CapabilityTaskCancel, CapabilityInteractiveRun, CapabilityLifecycleExecute}}
		if err := descriptor.Validate(); err != nil {
			t.Fatal(err)
		}
		if !descriptor.HasCapability(CapabilityLifecycleExecute) {
			t.Fatal("lifecycle capability was not disclosed")
		}
	}
	if err := (Descriptor{ProtocolVersion: "old", Mode: DeploymentModeModule}).Validate(); err == nil {
		t.Fatal("old protocol accepted")
	}
}

func TestServiceActionAuthorizationIsExactAndAudienceBound(t *testing.T) {
	ctx := WithAuthorizedServiceAction(t.Context(), ActionAgentTaskExecutionStart, AgentRuntimeServiceAudience)
	if !HasAuthorizedServiceAction(ctx, ActionAgentTaskExecutionStart, AgentRuntimeServiceAudience) {
		t.Fatal("exact service Action evidence was rejected")
	}
	if HasAuthorizedServiceAction(ctx, ActionAgentTaskExecutionPoll, AgentRuntimeServiceAudience) {
		t.Fatal("Start service Action authorized Poll")
	}
	if HasAuthorizedServiceAction(ctx, ActionAgentTaskExecutionStart, "other-runtime") {
		t.Fatal("service Action escaped its audience")
	}
	if HasAuthorizedServiceAction(context.Background(), ActionAgentTaskExecutionStart, AgentRuntimeServiceAudience) {
		t.Fatal("missing service Action evidence was accepted")
	}
}
