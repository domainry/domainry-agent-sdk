package businessrpc

import (
	"context"

	sdk "github.com/domainry/domainry-agent-sdk"
	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

// ReportReader is an optional host port over the Report SDK's existing query
// DTOs. No definitions repository, SQL text, snapshot mutation, export command,
// access token or caller-supplied ReportSubject crosses this boundary. The host
// resolves the current Identity and delegates authorization and execution to
// the actual Report owner. Conversation tools are a separate consumer layer.
type ReportReader interface {
	ReportSummary(context.Context, reportmodel.ReportSummaryRequest, sdk.ConversationAuthority) (reportmodel.ReportSummary, error)
	ReportObjectSQL(context.Context, reportmodel.ReportObjectSQLRequest, sdk.ConversationAuthority) (reportmodel.ReportSummary, error)
}

// GovernedReportReader adds Report-owned discovery and persisted-result checks.
// It remains optional so hosts without this Report SDK extension fail closed.
type GovernedReportReader interface {
	ReportReader
	ReportCatalog(context.Context, reportmodel.ReportCatalogRequest, sdk.ConversationAuthority) (reportmodel.ReportCatalog, error)
	QueryReport(context.Context, reportmodel.ReportObjectSQLRequest, sdk.ConversationAuthority) (reportmodel.ReportQueryResult, error)
	AuthorizeReportResult(context.Context, reportmodel.ReportQueryResultAuthorization, sdk.ConversationAuthority) error
}

func reportOperation(operation string) bool {
	return operation == "report_summary" || operation == "report_object_sql" || operation == "report_catalog" || operation == "report_query" || operation == "report_authorize_result"
}

func dispatchReport(ctx context.Context, backend Backend, in request) (any, error) {
	if in.Operation == "report_catalog" || in.Operation == "report_query" || in.Operation == "report_authorize_result" {
		return dispatchGovernedReport(ctx, backend, in)
	}
	reader, ok := backend.(ReportReader)
	if !ok {
		return nil, failure("unavailable")
	}
	switch in.Operation {
	case "report_summary":
		var v reportmodel.ReportSummaryRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return reader.ReportSummary(ctx, v, in.Authority)
	case "report_object_sql":
		var v reportmodel.ReportObjectSQLRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return reader.ReportObjectSQL(ctx, v, in.Authority)
	}
	return nil, failure("bad_request")
}

func dispatchGovernedReport(ctx context.Context, backend Backend, in request) (any, error) {
	reader, ok := backend.(GovernedReportReader)
	if !ok {
		return nil, failure("unavailable")
	}
	switch in.Operation {
	case "report_catalog":
		var v reportmodel.ReportCatalogRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return reader.ReportCatalog(ctx, v, in.Authority)
	case "report_query":
		var v reportmodel.ReportObjectSQLRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return reader.QueryReport(ctx, v, in.Authority)
	case "report_authorize_result":
		var v reportmodel.ReportQueryResultAuthorization
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := reader.AuthorizeReportResult(ctx, v, in.Authority)
		return err == nil, err
	}
	return nil, failure("bad_request")
}

func (c *Client) ReportSummary(ctx context.Context, v reportmodel.ReportSummaryRequest, a sdk.ConversationAuthority) (reportmodel.ReportSummary, error) {
	var result reportmodel.ReportSummary
	err := c.call(ctx, "report_summary", a, v, &result)
	return result, err
}

func (c *Client) ReportObjectSQL(ctx context.Context, v reportmodel.ReportObjectSQLRequest, a sdk.ConversationAuthority) (reportmodel.ReportSummary, error) {
	var result reportmodel.ReportSummary
	err := c.call(ctx, "report_object_sql", a, v, &result)
	return result, err
}

var _ ReportReader = (*Client)(nil)

func (c *Client) ReportCatalog(ctx context.Context, v reportmodel.ReportCatalogRequest, a sdk.ConversationAuthority) (reportmodel.ReportCatalog, error) {
	var result reportmodel.ReportCatalog
	err := c.call(ctx, "report_catalog", a, v, &result)
	return result, err
}

func (c *Client) QueryReport(ctx context.Context, v reportmodel.ReportObjectSQLRequest, a sdk.ConversationAuthority) (reportmodel.ReportQueryResult, error) {
	var result reportmodel.ReportQueryResult
	err := c.call(ctx, "report_query", a, v, &result)
	return result, err
}

func (c *Client) AuthorizeReportResult(ctx context.Context, v reportmodel.ReportQueryResultAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "report_authorize_result", a, v, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

var _ GovernedReportReader = (*Client)(nil)
