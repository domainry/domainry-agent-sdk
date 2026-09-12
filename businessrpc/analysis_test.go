package businessrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	agent "github.com/domainry/domainry-agent-sdk"
	model "github.com/domainry/domainry-report-sdk/model"
)

type analysisBackendFixture struct {
	*backendFixture
	denied  atomic.Bool
	padding int
}

func (b *analysisBackendFixture) AnalysisCatalog(_ context.Context, in model.AnalysisCatalogRequest, _ agent.ConversationAuthority) (model.AnalysisCatalog, error) {
	if b.denied.Load() {
		return model.AnalysisCatalog{}, failure("forbidden")
	}
	b.hit("analysis_catalog")
	return model.AnalysisCatalog{Datasets: []model.AnalysisDataset{{Key: "sale", Version: "v1", Columns: []model.AnalysisColumn{{Key: "amount", Type: "decimal", Unit: "CNY"}}}}, NextCursor: in.Page.Cursor}, nil
}
func (b *analysisBackendFixture) RunAnalysis(_ context.Context, in model.AnalysisRequest, _ agent.ConversationAuthority) (model.AnalysisResult, error) {
	if b.denied.Load() {
		return model.AnalysisResult{}, failure("forbidden")
	}
	if in.MaxRows == 1 {
		return model.AnalysisResult{}, &agent.Error{Class: "bad_request", Code: "backend.report.analysis.result_limit_exceeded"}
	}
	b.hit("analysis_run")
	value := "9007199254740993.20"
	out := model.AnalysisResult{Spec: in, Rows: []model.AnalysisRow{{Values: map[string]*string{"total": &value, "missing": nil}}}, Source: model.AnalysisSource{DatasetKey: in.DatasetKey, Complete: true, Proof: "owner-proof"}}
	if b.padding > 0 {
		padding := strings.Repeat("x", b.padding)
		out.Rows[0].Values["padding"] = &padding
	}
	return out, nil
}
func (b *analysisBackendFixture) AuthorizeAnalysisResult(_ context.Context, in model.AnalysisResultAuthorization, _ agent.ConversationAuthority) error {
	if b.denied.Load() || in.Result.Source.Proof != "owner-proof" || len(in.Request.Filters) != 1 || in.Request.Filters[0].Any[0].Values[0] != json.Number("9007199254740993") || *in.Result.Rows[0].Values["total"] != "9007199254740993.20" {
		return failure("forbidden")
	}
	b.hit("analysis_authorize_result")
	return nil
}

func TestAnalysisRPCPreservesRecursiveSpecNumbersAndCurrentOwnerChecks(t *testing.T) {
	b := &analysisBackendFixture{backendFixture: newBackend()}
	handler, err := NewHandler(ServerOptions{Scope: testScope, Token: testToken, Backend: b})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	c, err := Open(t.Context(), ClientOptions{BaseURL: server.URL, Token: testToken, Scope: testScope, ExpectedSourceIdentity: testSource, ExpectedContractSHA256: ContractSHA256()})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := c.AnalysisCatalog(t.Context(), model.AnalysisCatalogRequest{Page: model.ReportPageRequest{Cursor: "catalog-cursor"}}, testAuthority)
	if err != nil || catalog.NextCursor != "catalog-cursor" || len(catalog.Datasets) != 1 {
		t.Fatal(catalog, err)
	}
	r := model.AnalysisRequest{DatasetKey: "sale", Filters: []model.AnalysisFilter{{Any: []model.AnalysisFilter{{Field: "amount", Operator: "ge", Values: []any{json.Number("9007199254740993")}}}}}, Calculations: []model.AnalysisCalculation{{Key: "x", Expression: model.AnalysisExpression{Operator: "negate", Arguments: []model.AnalysisExpression{{Reference: "total"}}}}}}
	out, err := c.RunAnalysis(t.Context(), r, testAuthority)
	if err != nil || out.Spec.Filters[0].Any[0].Values[0] != json.Number("9007199254740993") || out.Rows[0].Values["missing"] != nil {
		t.Fatal(out, err)
	}
	check := model.AnalysisResultAuthorization{Request: r, Result: out}
	if err := c.AuthorizeAnalysisResult(t.Context(), check, testAuthority); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"analysis_catalog", "analysis_run", "analysis_authorize_result"} {
		if b.count(key) != 1 {
			t.Fatalf("%s count=%d", key, b.count(key))
		}
	}
	if b.count("identity") != 3 {
		t.Fatal("missing per-request execution authorization")
	}
	limited := r
	limited.MaxRows = 1
	_, err = c.RunAnalysis(t.Context(), limited, testAuthority)
	var tool *agent.Error
	if !errors.As(err, &tool) || tool.Code != "backend.report.analysis.result_limit_exceeded" {
		t.Fatal("limit reason lost across RPC", err)
	}
	for _, operation := range []string{"analysis_catalog", "analysis_run", "analysis_authorize_result"} {
		for _, payload := range []string{`{"sql":"SELECT secret"}`, `{"subject":{"principal":{"known":true}}}`, `{"access_token":"secret"}`} {
			raw, _ := json.Marshal(request{Descriptor: c.Descriptor(), Operation: operation, Authority: testAuthority, Payload: json.RawMessage(payload)})
			req, _ := http.NewRequest(http.MethodPost, server.URL+CallPath, strings.NewReader(string(raw)))
			req.Header.Set("Authorization", "Bearer "+testToken)
			req.Header.Set("Content-Type", "application/json")
			response, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != 400 {
				t.Fatalf("%s accepted %s: %d", operation, payload, response.StatusCode)
			}
		}
	}
	b.denied.Store(true)
	if _, err := c.AnalysisCatalog(t.Context(), model.AnalysisCatalogRequest{}, testAuthority); err == nil {
		t.Fatal("revoked catalog accepted")
	}
	if _, err := c.RunAnalysis(t.Context(), r, testAuthority); err == nil {
		t.Fatal("revoked analysis accepted")
	}
	if err := c.AuthorizeAnalysisResult(t.Context(), check, testAuthority); err == nil {
		t.Fatal("revoked history accepted")
	}
	legacy, _ := fixture(t, newBackend(), nil)
	if _, err := legacy.AnalysisCatalog(t.Context(), model.AnalysisCatalogRequest{}, testAuthority); err == nil {
		t.Fatal("absent owner claimed catalog")
	}
	if _, err := legacy.RunAnalysis(t.Context(), r, testAuthority); err == nil {
		t.Fatal("absent owner claimed execution")
	}
	if err := legacy.AuthorizeAnalysisResult(t.Context(), check, testAuthority); err == nil {
		t.Fatal("absent owner claimed proof")
	}
}

func TestRecursiveContractShapeIncludesNestedChangesAndKeepsAcyclicEncoding(t *testing.T) {
	type tree struct {
		Value    string `json:"value"`
		Children []tree `json:"children"`
	}
	type otherTree struct {
		Value    int         `json:"value"`
		Children []otherTree `json:"children"`
	}
	first, _ := json.Marshal(shape(reflect.TypeFor[tree]()))
	second, _ := json.Marshal(shape(reflect.TypeFor[otherTree]()))
	if !strings.Contains(string(first), "recursive-ref") || string(first) == string(second) {
		t.Fatal(string(first), string(second))
	}
	type leaf struct {
		Value string `json:"value"`
	}
	flat, _ := json.Marshal(shape(reflect.TypeFor[leaf]()))
	if string(flat) != `["object",[["value",false,"string"]]]` {
		t.Fatal("acyclic encoding changed", string(flat))
	}
	type siblings struct {
		A leaf `json:"a"`
		B leaf `json:"b"`
	}
	pair, _ := json.Marshal(shape(reflect.TypeFor[siblings]()))
	if strings.Count(string(pair), `"value"`) != 2 || strings.Contains(string(pair), "recursive-ref") {
		t.Fatal("sibling fields disappeared", string(pair))
	}
}

func TestAnalysisRPCRejectsResultsThatCannotBeReauthorizedOverHTTP(t *testing.T) {
	for _, padding := range []int{MaxResponseBytes - 4096, MaxResponseBytes - 768, MaxResponseBytes + 1} {
		t.Run(fmt.Sprint(padding), func(t *testing.T) {
			b := &analysisBackendFixture{backendFixture: newBackend(), padding: padding}
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
			r := model.AnalysisRequest{DatasetKey: "sale", Filters: []model.AnalysisFilter{{Any: []model.AnalysisFilter{{Field: "amount", Operator: "ge", Values: []any{json.Number("9007199254740993")}}}}}}
			// Inspect the unbounded owner fixture, so the middle case proves that
			// fitting a response alone is insufficient for persisted results.
			owner, _ := b.RunAnalysis(t.Context(), r, testAuthority)
			raw, _ := json.Marshal(owner)
			responseBytes, _ := json.Marshal(response{Data: raw})
			if padding == MaxResponseBytes-768 && len(responseBytes) > MaxResponseBytes {
				t.Fatal("fixture no longer fits response", len(responseBytes))
			}
			out, err := c.RunAnalysis(t.Context(), r, testAuthority)
			if padding == MaxResponseBytes-4096 {
				if err != nil {
					t.Fatal(err)
				}
				if err := c.AuthorizeAnalysisResult(t.Context(), model.AnalysisResultAuthorization{Request: r, Result: out}, testAuthority); err != nil {
					t.Fatal("accepted result cannot round trip", err)
				}
				return
			}
			var coded *agent.Error
			if !errors.As(err, &coded) || coded.Code != "backend.report.analysis.result_limit_exceeded" || len(out.Rows) != 0 {
				t.Fatal("oversize result lost precise failure or returned partial rows", err)
			}
		})
	}
}
