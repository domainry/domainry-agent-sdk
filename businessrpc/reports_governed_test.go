package businessrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	agent "github.com/domainry/domainry-agent-sdk"
	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

type governedReportFixture struct{ *reportBackendFixture }

func (b *governedReportFixture) ReportCatalog(_ context.Context, in reportmodel.ReportCatalogRequest, _ agent.ConversationAuthority) (reportmodel.ReportCatalog, error) {
	if b.denied.Load() {
		return reportmodel.ReportCatalog{}, failure("forbidden")
	}
	return reportmodel.ReportCatalog{Reports: []reportmodel.ReportCatalogEntry{{Key: "sales", DefinitionVersion: "definition-1", RowLimit: 100, Parameters: []reportmodel.ReportObjectSQLParameter{{Key: "large", Type: "integer", Default: json.Number("9007199254740993")}}}}, NextCursor: in.Page.Cursor, Truncated: true}, nil
}
func (b *governedReportFixture) QueryReport(ctx context.Context, in reportmodel.ReportObjectSQLRequest, a agent.ConversationAuthority) (reportmodel.ReportQueryResult, error) {
	out, err := b.ReportObjectSQL(ctx, in, a)
	return reportmodel.ReportQueryResult{Summary: out, Source: reportmodel.ReportQuerySource{ReportKey: "sales", Proof: "owner-test-proof", DefinitionVersion: "definition-1", DataVersion: "data-1", RowLimit: 100}}, err
}
func (b *governedReportFixture) AuthorizeReportResult(_ context.Context, in reportmodel.ReportQueryResultAuthorization, _ agent.ConversationAuthority) error {
	if b.denied.Load() || in.Query.Parameters["large"] != json.Number("9007199254740993") || in.Result.Source.Proof != "owner-test-proof" || in.Result.Summary.Rows[0].Measures["total"] != "9007199254740993.20" {
		return failure("forbidden")
	}
	return nil
}

func TestGovernedReportPortsPreserveDTOsAndRejectUntrustedPayloads(t *testing.T) {
	b := &governedReportFixture{reportBackendFixture: &reportBackendFixture{backendFixture: newBackend()}}
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
	catalog, err := c.ReportCatalog(t.Context(), reportmodel.ReportCatalogRequest{Page: reportmodel.ReportPageRequest{Cursor: "catalog-cursor"}}, testAuthority)
	if err != nil || catalog.NextCursor != "catalog-cursor" || catalog.Reports[0].Parameters[0].Default != json.Number("9007199254740993") {
		t.Fatal(catalog, err)
	}
	query := reportmodel.ReportObjectSQLRequest{ReportKey: "sales", Parameters: map[string]any{"large": json.Number("9007199254740993")}, Page: reportmodel.ReportPageRequest{Cursor: "owned-cursor"}}
	out, err := c.QueryReport(t.Context(), query, testAuthority)
	if err != nil || out.Source.Complete || out.Source.Proof != "owner-test-proof" {
		t.Fatal(out, err)
	}
	if err := c.AuthorizeReportResult(t.Context(), reportmodel.ReportQueryResultAuthorization{Query: query, Result: out}, testAuthority); err != nil {
		t.Fatal(err)
	}
	if b.reads.Load() != 1 || b.count("identity") != 3 {
		t.Fatal("source checks executed query or omitted current execution authorization")
	}
	for _, operation := range []string{"report_catalog", "report_query", "report_authorize_result"} {
		for _, extra := range []string{`"sql":"SELECT secret"`, `"subject":{"principal":{"known":true}}`, `"access_token":"secret"`} {
			raw, _ := json.Marshal(request{Descriptor: c.Descriptor(), Operation: operation, Authority: testAuthority, Payload: json.RawMessage("{" + extra + "}")})
			r, _ := http.NewRequest(http.MethodPost, s.URL+CallPath, strings.NewReader(string(raw)))
			r.Header.Set("Authorization", "Bearer "+testToken)
			r.Header.Set("Content-Type", "application/json")
			response, err := s.Client().Do(r)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != 400 {
				t.Fatalf("%s accepted %s: %d", operation, extra, response.StatusCode)
			}
		}
	}
	b.denied.Store(true)
	if _, err := c.ReportCatalog(t.Context(), reportmodel.ReportCatalogRequest{}, testAuthority); err == nil {
		t.Fatal("catalog ignored revocation")
	}
	if _, err := c.QueryReport(t.Context(), query, testAuthority); err == nil {
		t.Fatal("query ignored revocation")
	}
	if err := c.AuthorizeReportResult(t.Context(), reportmodel.ReportQueryResultAuthorization{Query: query, Result: out}, testAuthority); err == nil {
		t.Fatal("saved result ignored revocation")
	}
	legacy, _ := fixture(t, newBackend(), nil)
	if _, err := legacy.ReportCatalog(t.Context(), reportmodel.ReportCatalogRequest{}, testAuthority); err == nil {
		t.Fatal("missing catalog claimed success")
	}
	if _, err := legacy.QueryReport(t.Context(), query, testAuthority); err == nil {
		t.Fatal("missing query claimed success")
	}
	if err := legacy.AuthorizeReportResult(t.Context(), reportmodel.ReportQueryResultAuthorization{Query: query, Result: out}, testAuthority); err == nil {
		t.Fatal("missing evidence owner claimed success")
	}
}
