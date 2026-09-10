package contracttest

import (
	"testing"

	agentsdk "github.com/domainry/domainry-agent-sdk"
	agentpersistence "github.com/domainry/domainry-agent-sdk/persistence"
)

func VerifyBinding(t *testing.T, binding agentsdk.Binding, mode agentsdk.DeploymentMode) {
	t.Helper()
	if binding == nil {
		t.Fatal("Agent Binding is nil")
	}
	descriptor := binding.Descriptor()
	if err := descriptor.Validate(); err != nil {
		t.Fatal(err)
	}
	if descriptor.Mode != mode {
		t.Fatalf("mode=%q want=%q", descriptor.Mode, mode)
	}
	if descriptor.HasCapability(agentsdk.CapabilityTaskStart) && binding.TaskRunner() == nil {
		t.Fatal("advertised TaskRunner is unavailable")
	}
	if descriptor.HasCapability(agentsdk.CapabilityInteractiveRun) && binding.InteractiveRunner() == nil {
		t.Fatal("advertised InteractiveRunner is unavailable")
	}
	for _, capability := range descriptor.Capabilities {
		switch capability {
		case agentsdk.CapabilityConversationV1:
			conversation, ok := binding.(agentsdk.ConversationBinding)
			if !ok || conversation.Conversations() == nil {
				t.Fatal("advertised conversation service is unavailable")
			}
		case "dialog.state":
			stateBinding, ok := binding.(agentsdk.AgentDialogStateBinding)
			if !ok || stateBinding.DialogState() == nil {
				t.Fatal("Agent Binding advertises dialog.state without a dialog state service")
			}
		case "execution.state":
			stateBinding, ok := binding.(agentpersistence.ExecutionStateBinding)
			if !ok || stateBinding.AgentTaskState() == nil || stateBinding.AgentInteractiveState() == nil {
				t.Fatal("Agent Binding advertises execution.state without task and interactive state services")
			}
		}
	}
}
