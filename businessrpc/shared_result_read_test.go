package businessrpc

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/domainry/domainry-agent-sdk"
	model "github.com/domainry/domainry-report-sdk/model"
)

type sharedResultReadFixture struct {
	*backendFixture
	producer sdk.ConversationAuthority
	expected map[string]any
	denied   atomic.Bool
}

func (b *sharedResultReadFixture) read(operation string, input any, reader, producer sdk.ConversationAuthority) error {
	b.hit(operation)
	if b.denied.Load() || reader != testAuthority || producer != b.producer || !reflect.DeepEqual(input, b.expected[operation]) {
		return failure("forbidden")
	}
	return nil
}
func (b *sharedResultReadFixture) AuthorizeSharedReportResultRead(_ context.Context, input model.ReportQueryResultAuthorization, reader, producer sdk.ConversationAuthority) error {
	return b.read("report_shared_result_read", input, reader, producer)
}
func (b *sharedResultReadFixture) AuthorizeSharedReportCatalogRead(_ context.Context, input model.ReportCatalogReadAuthorization, reader, producer sdk.ConversationAuthority) error {
	return b.read("report_shared_catalog_read", input, reader, producer)
}
func (b *sharedResultReadFixture) AuthorizeSharedAnalysisResultRead(_ context.Context, input model.AnalysisResultAuthorization, reader, producer sdk.ConversationAuthority) error {
	return b.read("analysis_shared_result_read", input, reader, producer)
}
func (b *sharedResultReadFixture) AuthorizeSharedAnalysisCatalogRead(_ context.Context, input model.AnalysisCatalogReadAuthorization, reader, producer sdk.ConversationAuthority) error {
	return b.read("analysis_shared_catalog_read", input, reader, producer)
}

func TestSharedProfessionalReadRPCPreservesReaderAndOriginalProvenance(t *testing.T) {
	producer := testAuthority
	producer.UserID, producer.RoleKey = "professional-b", "professional-role"
	query := model.ReportQueryResultAuthorization{Query: model.ReportObjectSQLRequest{ReportKey: "sales", Parameters: map[string]any{"minimum": json.Number("9007199254740993")}}, Result: model.ReportQueryResult{Source: model.ReportQuerySource{ReadProof: "original-report-read-proof"}}}
	catalog := model.ReportCatalogReadAuthorization{Result: model.ReportCatalog{ReadProof: "original-catalog-proof"}}
	analysis := model.AnalysisResultAuthorization{Request: model.AnalysisRequest{DatasetKey: "sales"}, Result: model.AnalysisResult{Source: model.AnalysisSource{ReadProof: "original-analysis-proof"}}}
	analysisCatalog := model.AnalysisCatalogReadAuthorization{Request: model.AnalysisCatalogRequest{DatasetKey: "sales"}, Result: model.AnalysisCatalog{ReadProof: "original-analysis-catalog-proof"}}
	b := &sharedResultReadFixture{backendFixture: newBackend(), producer: producer, expected: map[string]any{
		"report_shared_result_read": query, "report_shared_catalog_read": catalog,
		"analysis_shared_result_read": analysis, "analysis_shared_catalog_read": analysisCatalog,
	}}
	b.allow.Store(false)
	handler, err := NewHandler(ServerOptions{Scope: testScope, Token: testToken, Backend: b})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	client, err := Open(t.Context(), ClientOptions{BaseURL: server.URL, Token: testToken, Scope: testScope, ExpectedSourceIdentity: testSource, ExpectedContractSHA256: ContractSHA256()})
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		operation string
		read      func(*Client, sdk.ConversationAuthority, sdk.ConversationAuthority) error
	}{
		{"report_shared_result_read", func(c *Client, a, p sdk.ConversationAuthority) error {
			return c.AuthorizeSharedReportResultRead(t.Context(), query, a, p)
		}},
		{"report_shared_catalog_read", func(c *Client, a, p sdk.ConversationAuthority) error {
			return c.AuthorizeSharedReportCatalogRead(t.Context(), catalog, a, p)
		}},
		{"analysis_shared_result_read", func(c *Client, a, p sdk.ConversationAuthority) error {
			return c.AuthorizeSharedAnalysisResultRead(t.Context(), analysis, a, p)
		}},
		{"analysis_shared_catalog_read", func(c *Client, a, p sdk.ConversationAuthority) error {
			return c.AuthorizeSharedAnalysisCatalogRead(t.Context(), analysisCatalog, a, p)
		}},
	}
	legacy, _ := fixture(t, newBackend(), nil)
	for _, check := range checks {
		t.Run(check.operation, func(t *testing.T) {
			if err := check.read(client, testAuthority, producer); err != nil || b.count(check.operation) != 1 {
				t.Fatal("shared read did not reach original owner", err)
			}
			wrong := producer
			wrong.UserID = "unrelated-user"
			if err := check.read(client, testAuthority, wrong); err == nil {
				t.Fatal("wrong proof producer accepted")
			}
			for _, mutation := range []func(*sdk.ConversationAuthority){
				func(p *sdk.ConversationAuthority) { p.WorkspaceID = "foreign" },
				func(p *sdk.ConversationAuthority) { p.RuntimeID = "foreign" },
				func(p *sdk.ConversationAuthority) { p.UserID = "" },
				func(p *sdk.ConversationAuthority) { p.Known = false },
			} {
				invalid := producer
				mutation(&invalid)
				before := b.count(check.operation)
				if err := check.read(client, testAuthority, invalid); err == nil || b.count(check.operation) != before {
					t.Fatal("invalid selector reached owner", err)
				}
			}
			b.denied.Store(true)
			if err := check.read(client, testAuthority, producer); err == nil {
				t.Fatal("revoked reader accepted")
			}
			b.denied.Store(false)
			b.active.Store(false)
			before := b.count(check.operation)
			if err := check.read(client, testAuthority, producer); err == nil || b.count(check.operation) != before {
				t.Fatal("inactive reader reached owner")
			}
			b.active.Store(true)
			if err := check.read(legacy, testAuthority, producer); err == nil {
				t.Fatal("missing shared owner treated as permission")
			}
		})
	}
	if b.count("tool") != 0 || b.count("query") != 0 || b.count("action_invoke") != 0 || b.count("workflow_start") != 0 {
		t.Fatal("read invoked an execution port")
	}
	if _, err := Open(t.Context(), ClientOptions{BaseURL: server.URL, Token: testToken, Scope: testScope, ExpectedSourceIdentity: testSource, ExpectedContractSHA256: strings.Repeat("0", 64)}); err == nil {
		t.Fatal("incompatible contract accepted")
	}
	request := sdk.ConversationToolRequest{Authority: testAuthority, ResultProducer: &producer}
	raw, err := json.Marshal(request)
	if err != nil || strings.Contains(string(raw), producer.UserID) || strings.Contains(string(raw), "ResultProducer") {
		t.Fatal("server provenance became public tool input", string(raw), err)
	}
}
