package agentsdk

import "testing"

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
