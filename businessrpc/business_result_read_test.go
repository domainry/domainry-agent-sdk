package businessrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	sdk "github.com/domainry/domainry-agent-sdk"
)

type businessResultReadFixture struct {
	*backendFixture
	e      sdk.ConversationBusinessEvidence
	denied atomic.Bool
}

func (b *businessResultReadFixture) AuthorizeBusinessResultRead(_ context.Context, e sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) error {
	b.hit("result_read")
	if b.denied.Load() || a != testAuthority || !reflect.DeepEqual(e, b.e) {
		return failure("forbidden")
	}
	return nil
}

func TestBusinessResultReadRPCPreservesEvidenceWithoutToolGrant(t *testing.T) {
	b := &businessResultReadFixture{backendFixture: newBackend(), e: sdk.ConversationBusinessEvidence{Version: 1, Operation: "get_record", Source: testSource, ScopeSHA256: "owner-scope", Input: json.RawMessage(`{"object_key":"customer","record_id":"record-a"}`), Data: json.RawMessage(`{"amount":9007199254740993}`), HostProof: "source-proof"}}
	b.allow.Store(false)
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
	if err := c.AuthorizeBusinessResultRead(t.Context(), b.e, testAuthority); err != nil || b.count("result_read") != 1 || b.count("tool") != 0 || b.count("identity") != 1 {
		t.Fatal("independent read transport", err)
	}
	if err := c.RevalidateBusiness(t.Context(), b.e, testAuthority); err == nil || b.count("revalidate") != 0 {
		t.Fatal("ordinary revalidation escaped tool rights", err)
	}
	changed := b.e
	changed.Data = json.RawMessage(`{"amount":9007199254740992}`)
	if err := c.AuthorizeBusinessResultRead(t.Context(), changed, testAuthority); err == nil {
		t.Fatal("modified value accepted")
	}
	b.denied.Store(true)
	if err := c.AuthorizeBusinessResultRead(t.Context(), b.e, testAuthority); err == nil {
		t.Fatal("revoked owner access accepted")
	}
	b.denied.Store(false)
	before := b.count("result_read")
	foreign := testAuthority
	foreign.WorkspaceID = "other"
	if err := c.AuthorizeBusinessResultRead(t.Context(), b.e, foreign); err == nil || b.count("result_read") != before {
		t.Fatal("foreign workspace reached owner")
	}
	b.active.Store(false)
	if err := c.AuthorizeBusinessResultRead(t.Context(), b.e, testAuthority); err == nil || b.count("result_read") != before {
		t.Fatal("inactive principal reached owner")
	}
	legacy, _ := fixture(t, newBackend(), nil)
	var coded *sdk.Error
	if err := legacy.AuthorizeBusinessResultRead(t.Context(), b.e, testAuthority); !errors.As(err, &coded) || coded.Class != "unavailable" || coded.Code != sdk.BusinessResultReadUnsupportedCode {
		t.Fatal("missing optional source lost exact unsupported code", err)
	}
	if b.count("action_invoke") != 0 || b.count("action_reconcile") != 0 || b.count("workflow_start") != 0 || b.count("workflow_reconcile") != 0 {
		t.Fatal("reading re-executed a write")
	}
}
