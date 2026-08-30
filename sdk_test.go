package agentsdk

import "testing"

func TestDescriptorAndApplicationValidation(t *testing.T) {
	if err := (ApplicationRef{}).Validate(); err == nil {
		t.Fatal("empty application accepted")
	}
	for _, mode := range []DeploymentMode{DeploymentModeModule, DeploymentModeSaaS} {
		if err := (Descriptor{ProtocolVersion: ProtocolVersionV1, Mode: mode, Capabilities: []string{"task.start", "task.poll", "task.cancel", "interactive.run"}}).Validate(); err != nil {
			t.Fatal(err)
		}
	}
	if err := (Descriptor{ProtocolVersion: "old", Mode: DeploymentModeModule}).Validate(); err == nil {
		t.Fatal("old protocol accepted")
	}
}
