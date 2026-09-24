package agentsdk_test

import (
	"strings"
	"testing"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func TestConversationProvenancePublicationRequiresBoundedContentAndClientIdentities(t *testing.T) {
	valid := agentsdk.ConversationProvenancePublication{
		ClientID: "deck-turn:1", ConversationClientID: "deck-thread:1", ConversationTitle: "Feature discovery",
		UserMessage: "Please add cancellation.", AssistantMessage: "The requirement is ready.",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []agentsdk.ConversationProvenancePublication{
		{},
		{ClientID: "bad/client", ConversationClientID: "thread", UserMessage: "user", AssistantMessage: "assistant"},
		{ClientID: "turn", ConversationClientID: "thread", UserMessage: "", AssistantMessage: "assistant"},
		{ClientID: "turn", ConversationClientID: "thread", UserMessage: "user", AssistantMessage: strings.Repeat("x", agentsdk.ConversationProvenanceContentMaxBytes+1)},
	} {
		if err := invalid.Validate(); err == nil {
			t.Fatalf("invalid provenance publication accepted: %#v", invalid)
		}
	}
}
