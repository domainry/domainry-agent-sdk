package businessrpc

import (
	"context"
	"encoding/json"

	sdk "github.com/domainry/domainry-agent-sdk"
	model "github.com/domainry/domainry-report-sdk/model"
)

// AnalysisReader is an optional, independent data-owner port. It neither
// grants access to arbitrary SQL nor requires published report-query support.
// Servers resolve current identity and delegate the closed spec to Report.
type AnalysisReader interface {
	AnalysisCatalog(context.Context, model.AnalysisCatalogRequest, sdk.ConversationAuthority) (model.AnalysisCatalog, error)
	RunAnalysis(context.Context, model.AnalysisRequest, sdk.ConversationAuthority) (model.AnalysisResult, error)
	AuthorizeAnalysisResult(context.Context, model.AnalysisResultAuthorization, sdk.ConversationAuthority) error
}

func analysisOperation(operation string) bool {
	return operation == "analysis_catalog" || operation == "analysis_run" || operation == "analysis_authorize_result" || operation == "analysis_result_read" || operation == "analysis_catalog_read"
}
func dispatchAnalysis(ctx context.Context, backend Backend, in request) (any, error) {
	if in.Operation == "analysis_result_read" || in.Operation == "analysis_catalog_read" {
		return dispatchResultRead(ctx, backend, in)
	}
	reader, ok := backend.(AnalysisReader)
	if !ok {
		return nil, failure("unavailable")
	}
	switch in.Operation {
	case "analysis_catalog":
		var v model.AnalysisCatalogRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return reader.AnalysisCatalog(ctx, v, in.Authority)
	case "analysis_run":
		var v model.AnalysisRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		out, err := reader.RunAnalysis(ctx, v, in.Authority)
		if err != nil {
			return nil, err
		}
		// A successful read must fit both this response and the later saved-result
		// authorization request. Bound the exact transport envelopes, never trim
		// rows or weaken the shared transport limits.
		result, err := json.Marshal(out)
		if err != nil {
			return nil, failure("unavailable")
		}
		encoded, err := json.Marshal(response{Data: result})
		if err != nil {
			return nil, failure("unavailable")
		}
		payload, err := json.Marshal(model.AnalysisResultAuthorization{Request: v, Result: out})
		if err != nil {
			return nil, failure("unavailable")
		}
		authorization, err := json.Marshal(request{Descriptor: in.Descriptor, Operation: "analysis_authorize_result", Authority: in.Authority, Payload: payload})
		if err != nil {
			return nil, failure("unavailable")
		}
		if len(encoded) > MaxResponseBytes || len(authorization) > MaxRequestBytes {
			return nil, &sdk.Error{Class: "bad_request", Code: "backend.report.analysis.result_limit_exceeded"}
		}
		return out, nil
	case "analysis_authorize_result":
		var v model.AnalysisResultAuthorization
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := reader.AuthorizeAnalysisResult(ctx, v, in.Authority)
		return err == nil, err
	}
	return nil, failure("bad_request")
}
func (c *Client) AnalysisCatalog(ctx context.Context, v model.AnalysisCatalogRequest, a sdk.ConversationAuthority) (model.AnalysisCatalog, error) {
	var out model.AnalysisCatalog
	err := c.call(ctx, "analysis_catalog", a, v, &out)
	return out, err
}
func (c *Client) RunAnalysis(ctx context.Context, v model.AnalysisRequest, a sdk.ConversationAuthority) (model.AnalysisResult, error) {
	var out model.AnalysisResult
	err := c.call(ctx, "analysis_run", a, v, &out)
	return out, err
}
func (c *Client) AuthorizeAnalysisResult(ctx context.Context, v model.AnalysisResultAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "analysis_authorize_result", a, v, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

var _ AnalysisReader = (*Client)(nil)
