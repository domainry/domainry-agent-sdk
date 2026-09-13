package businessrpc

import (
	"context"
	sdk "github.com/domainry/domainry-agent-sdk"
	reportmodel "github.com/domainry/domainry-report-sdk/model"
)

type ReportResultReadSource interface {
	AuthorizeReportResultRead(context.Context, reportmodel.ReportQueryResultAuthorization, sdk.ConversationAuthority) error
	AuthorizeReportCatalogRead(context.Context, reportmodel.ReportCatalogReadAuthorization, sdk.ConversationAuthority) error
}
type AnalysisResultReadSource interface {
	AuthorizeAnalysisResultRead(context.Context, reportmodel.AnalysisResultAuthorization, sdk.ConversationAuthority) error
	AuthorizeAnalysisCatalogRead(context.Context, reportmodel.AnalysisCatalogReadAuthorization, sdk.ConversationAuthority) error
}

func (c *Client) AuthorizeBusinessResultRead(ctx context.Context, in sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "business_result_read", a, in, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

var _ sdk.ConversationBusinessResultReadSource = (*Client)(nil)

func dispatchResultRead(ctx context.Context, backend Backend, in request) (any, error) {
	switch in.Operation {
	case "report_result_read":
		var input reportmodel.ReportQueryResultAuthorization
		if decode(in.Payload, &input) != nil {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(ReportResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeReportResultRead(ctx, input, in.Authority)
		return err == nil, err
	case "report_catalog_read":
		var input reportmodel.ReportCatalogReadAuthorization
		if decode(in.Payload, &input) != nil {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(ReportResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeReportCatalogRead(ctx, input, in.Authority)
		return err == nil, err
	case "analysis_result_read":
		var input reportmodel.AnalysisResultAuthorization
		if decode(in.Payload, &input) != nil {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(AnalysisResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeAnalysisResultRead(ctx, input, in.Authority)
		return err == nil, err
	case "analysis_catalog_read":
		var input reportmodel.AnalysisCatalogReadAuthorization
		if decode(in.Payload, &input) != nil {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(AnalysisResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeAnalysisCatalogRead(ctx, input, in.Authority)
		return err == nil, err
	}
	return nil, failure("bad_request")
}
func (c *Client) AuthorizeReportResultRead(ctx context.Context, in reportmodel.ReportQueryResultAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "report_result_read", a, in, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}
func (c *Client) AuthorizeReportCatalogRead(ctx context.Context, in reportmodel.ReportCatalogReadAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "report_catalog_read", a, in, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}
func (c *Client) AuthorizeAnalysisResultRead(ctx context.Context, in reportmodel.AnalysisResultAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "analysis_result_read", a, in, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}
func (c *Client) AuthorizeAnalysisCatalogRead(ctx context.Context, in reportmodel.AnalysisCatalogReadAuthorization, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "analysis_catalog_read", a, in, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}
