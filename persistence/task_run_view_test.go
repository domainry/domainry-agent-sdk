package persistence

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
	agentmodel "github.com/domainry/domainry-agent-sdk/state"
)

func TestProjectAgentTaskRunRedactsInternalState(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	run := agentmodel.AgentTaskRun{
		ID: "run", WorkspaceID: "workspace", TaskKey: "review", TaskVersion: "1", Status: agentmodel.AgentTaskRunWaitingApproval,
		Input: map[string]any{"secret": "hidden"}, Output: map[string]any{"score": 90}, RawEvidenceRef: "raw-secret",
		Identity: agentsdk.ExecutionIdentity{Execution: agentsdk.PrincipalReference{UserID: "user", RoleKey: "reviewer"}},
		Approval: &agentmodel.AgentTaskApproval{ProposalID: "proposal", Execution: map[string]any{"action_receipt_id": "receipt"}},
		Evidence: agentmodel.AgentTaskExecutionEvidence{AuditRefs: []string{"audit"}}, CreatedAt: now, UpdatedAt: now,
	}
	raw, err := json.Marshal(ProjectAgentTaskRun(run))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, expected := range []string{"proposal", "receipt", "audit", `"score":90`} {
		if !strings.Contains(text, expected) {
			t.Errorf("projection missing %s: %s", expected, text)
		}
	}
	for _, forbidden := range []string{"hidden", "raw-secret", "fencing_token", "tool_invocations"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("projection leaked %s: %s", forbidden, text)
		}
	}
}

func TestProjectAgentTaskRunClonesMutableSlicesAndMaps(t *testing.T) {
	run := agentmodel.AgentTaskRun{Output: map[string]any{"value": "before"}, Evidence: agentmodel.AgentTaskExecutionEvidence{AuditRefs: []string{"audit-before"}}}
	view := ProjectAgentTaskRun(run)
	run.Output["value"] = "after"
	run.Evidence.AuditRefs[0] = "audit-after"
	if view.Output["value"] != "before" || view.AuditRefs[0] != "audit-before" {
		t.Fatalf("view aliases source state: %#v", view)
	}
}
