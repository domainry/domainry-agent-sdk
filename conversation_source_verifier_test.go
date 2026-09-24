package agentsdk_test

import (
	"testing"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func TestConversationSourceVerificationRequestRequiresExactReaderAndRun(t *testing.T) {
	request := agentsdk.ConversationSourceVerificationRequest{
		Reader:     agentsdk.ConversationAuthority{Known: true, RuntimeID: "runtime", WorkspaceID: "workspace", UserID: "reader"},
		References: []agentsdk.ConversationRunReference{{ConversationID: "conversation", RunID: "run", BeforeStep: 2}},
		SourceIDs:  []string{"conversation://conversation/turn/run"},
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	request.References[0].BeforeStep = -1
	if err := request.Validate(); err == nil {
		t.Fatal("negative BeforeStep was accepted")
	}
}

func TestConversationRunSourceIdentityIsCanonical(t *testing.T) {
	reference := agentsdk.ConversationRunReference{ConversationID: "conversation-1", RunID: "run_1"}
	identity, err := agentsdk.ConversationRunSourceID(reference)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := agentsdk.ParseConversationRunSourceID(identity)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ConversationID != reference.ConversationID || parsed.RunID != reference.RunID {
		t.Fatalf("parsed=%#v want=%#v", parsed, reference)
	}
	for _, invalid := range []string{
		"source-a",
		"conversation://conversation-1/run/run_1",
		"conversation://conversation-1/turn/run_1/extra",
		"conversation://conversation 1/turn/run_1",
	} {
		if _, err := agentsdk.ParseConversationRunSourceID(invalid); err == nil {
			t.Fatalf("accepted non-canonical source %q", invalid)
		}
	}
}
