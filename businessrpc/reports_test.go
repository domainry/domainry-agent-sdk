package businessrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	agent "github.com/domainry/domainry-agent-sdk"
	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

type reportBackendFixture struct {
	*backendFixture
	reads  atomic.Int32
	denied atomic.Bool
}

func (b *reportBackendFixture) ReportSummary(_ context.Context, in reportmodel.ReportSummaryRequest, _ agent.ConversationAuthority) (reportmodel.ReportSummary, error) {
	b.reads.Add(1)
	if b.denied.Load() {
		return reportmodel.ReportSummary{}, failure("forbidden")
	}
	if in.ReportKey != "sales" || in.Parameters["large"] != json.Number("9007199254740993") || in.Page.Cursor != "owned-cursor" {
		return reportmodel.ReportSummary{}, failure("bad_request")
	}
	return reportmodel.ReportSummary{Key: in.ReportKey, Rows: []reportmodel.ReportResultRow{{Measures: map[string]string{"total": "9007199254740993.20"}}}, RowCount: 1, PageSize: 1, NextCursor: "next-cursor", TotalSemantics: "at_least", Truncated: true}, nil
}
func (b *reportBackendFixture) ReportObjectSQL(ctx context.Context, in reportmodel.ReportObjectSQLRequest, a agent.ConversationAuthority) (reportmodel.ReportSummary, error) {
	return b.ReportSummary(ctx, reportmodel.ReportSummaryRequest{ReportKey: in.ReportKey, Parameters: in.Parameters, Page: in.Page}, a)
}

func TestReportHostPortsPreserveSDKDTOsAndOwnerAuthorization(t *testing.T) {
	b := &reportBackendFixture{backendFixture: newBackend()}
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
	params := map[string]any{"large": json.Number("9007199254740993")}
	page := reportmodel.ReportPageRequest{PageSize: 1, Cursor: "owned-cursor"}
	first, err := c.ReportSummary(t.Context(), reportmodel.ReportSummaryRequest{ReportKey: "sales", Parameters: params, Page: page}, testAuthority)
	if err != nil || !first.Truncated || first.TotalSemantics != "at_least" || first.NextCursor != "next-cursor" || first.Rows[0].Measures["total"] != "9007199254740993.20" {
		t.Fatal(first, err)
	}
	if _, err := c.ReportObjectSQL(t.Context(), reportmodel.ReportObjectSQLRequest{ReportKey: "sales", Parameters: params, Page: page}, testAuthority); err != nil {
		t.Fatal(err)
	}
	if b.reads.Load() != 2 || b.count("identity") != 2 || b.count("tool") != 0 {
		t.Fatal("Report must use current execution and owner policy, not an invented tool grant")
	}
	for _, payload := range []string{`{"report_key":"sales","sql":"SELECT private"}`, `{"report_key":"sales","subject":{"principal":{"known":true}}}`, `{"report_key":"sales","access_token":"browser-secret"}`} {
		raw, _ := json.Marshal(request{Descriptor: c.Descriptor(), Operation: "report_summary", Authority: testAuthority, Payload: json.RawMessage(payload)})
		r, _ := http.NewRequest("POST", s.URL+CallPath, strings.NewReader(string(raw)))
		r.Header.Set("Authorization", "Bearer "+testToken)
		r.Header.Set("Content-Type", "application/json")
		response, err := s.Client().Do(r)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 400 {
			t.Fatalf("untrusted Report input accepted: %d", response.StatusCode)
		}
	}
	if b.reads.Load() != 2 {
		t.Fatal("untrusted payload reached Report")
	}
	b.denied.Store(true)
	if _, err := c.ReportSummary(t.Context(), reportmodel.ReportSummaryRequest{ReportKey: "sales", Parameters: params, Page: page}, testAuthority); err == nil {
		t.Fatal("Report revocation ignored")
	}
	plain, _ := fixture(t, newBackend(), nil)
	if _, err := plain.ReportSummary(t.Context(), reportmodel.ReportSummaryRequest{ReportKey: "sales"}, testAuthority); err == nil {
		t.Fatal("absent Report owner claimed available")
	}
}
