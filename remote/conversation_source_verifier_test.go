package remote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

func TestConversationSourceVerifierUsesBoundedAuthenticatedOwnerRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/agent/conversation-sources/verify" || request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer service-token" {
			t.Fatalf("unexpected owner request: %s %s authorization=%q", request.Method, request.URL.Path, request.Header.Get("Authorization"))
		}
		var input agentsdk.ConversationSourceVerificationRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(writer).Encode(agentsdk.ConversationSourceVerificationReceipt{
			WorkspaceID: input.Reader.WorkspaceID, References: input.References, SourceIDs: input.SourceIDs,
			DecisionIDs: input.DecisionIDs, VerifiedAt: time.Now().UTC(),
		})
	}))
	defer server.Close()
	verifier, err := NewConversationSourceVerifier(ConversationSourceVerifierConfig{Endpoint: server.URL, AccessToken: "service-token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	request := agentsdk.ConversationSourceVerificationRequest{
		References: []agentsdk.ConversationRunReference{{ConversationID: "conversation-a", RunID: "run-a"}}, SourceIDs: []string{"source-a"},
		Reader: agentsdk.ConversationAuthority{Known: true, RuntimeID: "delivery", WorkspaceID: "workspace-a", UserID: "user-a"},
	}
	receipt, err := verifier.VerifyConversationSources(t.Context(), request)
	if err != nil || receipt.WorkspaceID != "workspace-a" {
		t.Fatalf("receipt=%#v err=%v", receipt, err)
	}
}
