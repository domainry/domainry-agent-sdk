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
	required := map[string]bool{"task.start": false, "task.poll": false, "task.cancel": false, "interactive.run": false}
	for _, capability := range descriptor.Capabilities {
		if _, ok := required[capability]; ok {
			required[capability] = true
		}
	}
	for capability, present := range required {
		if !present {
			t.Fatalf("capability %q missing", capability)
		}
	}
	if binding.TaskRunner() == nil || binding.InteractiveRunner() == nil {
		t.Fatal("Agent Binding runners are incomplete")
	}
	for _, capability := range descriptor.Capabilities {
		switch capability {
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
