package agentsdk_test

import (
	"testing"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func TestConversationSourceVerificationRequestRequiresExactReaderAndRun(t *testing.T) {
	request := agentsdk.ConversationSourceVerificationRequest{
		Reader:     agentsdk.ConversationAuthority{Known: true, RuntimeID: "runtime", WorkspaceID: "workspace", UserID: "reader"},
		References: []agentsdk.ConversationRunReference{{ConversationID: "conversation", RunID: "run", BeforeStep: 2}},
		SourceIDs:  []string{"conversation://conversation/run/run"}, DecisionIDs: []string{"D-001"},
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	request.References[0].BeforeStep = -1
	if err := request.Validate(); err == nil {
		t.Fatal("negative BeforeStep was accepted")
	}
}
