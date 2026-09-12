package businessrpc

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	sdk "github.com/domainry/domainry-agent-sdk"
)

type actionProbeBackend struct {
	*backendFixture
	seen chan sdk.ConversationBusinessAction
}

func (b *actionProbeBackend) AuthorizeBusinessAction(_ context.Context, action sdk.ConversationBusinessAction, _ sdk.ConversationAuthority) (sdk.ConversationToolAuthorization, error) {
	b.seen <- action
	return sdk.ConversationToolAuthorization{Granted: true}, nil
}

func TestActionCapabilityProbeSurvivesHTTPWithoutChangingConcreteInput(t *testing.T) {
	b := &actionProbeBackend{backendFixture: newBackend(), seen: make(chan sdk.ConversationBusinessAction, 1)}
	h, err := NewHandler(ServerOptions{Scope: testScope, Token: testToken, Backend: b})
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewServer(h)
	defer s.Close()
	c, err := Open(t.Context(), ClientOptions{BaseURL: s.URL, Token: testToken, Scope: testScope, ExpectedSourceIdentity: testSource, ExpectedContractSHA256: ContractSHA256()})
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []sdk.ConversationBusinessAction{
		{},
		{ObjectKey: "customer", ActionKey: "customer.create", Version: "1", Data: json.RawMessage("null")},
		{ObjectKey: "customer", ActionKey: "customer.create", Version: "1", Data: json.RawMessage(`{"count":9007199254740993}`)},
	} {
		if _, err := c.AuthorizeBusinessAction(t.Context(), input, testAuthority); err != nil {
			t.Fatal(err)
		}
		got := <-b.seen
		if got.ObjectKey != input.ObjectKey || got.ActionKey != input.ActionKey || got.Version != input.Version || string(got.Data) != string(input.Data) {
			t.Fatalf("HTTP changed action authorization input: got=%+v want=%+v", got, input)
		}
	}
	if b.count("action_invoke") != 0 {
		t.Fatal("capability probe executed a write")
	}
}
