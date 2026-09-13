package businessrpc

import (
	"context"
	sdk "github.com/domainry/domainry-agent-sdk"
	model "github.com/domainry/domainry-report-sdk/model"
)

type SharedReportResultReadSource interface {
	AuthorizeSharedReportResultRead(context.Context, model.ReportQueryResultAuthorization, sdk.ConversationAuthority, sdk.ConversationAuthority) error
	AuthorizeSharedReportCatalogRead(context.Context, model.ReportCatalogReadAuthorization, sdk.ConversationAuthority, sdk.ConversationAuthority) error
}
type SharedAnalysisResultReadSource interface {
	AuthorizeSharedAnalysisResultRead(context.Context, model.AnalysisResultAuthorization, sdk.ConversationAuthority, sdk.ConversationAuthority) error
	AuthorizeSharedAnalysisCatalogRead(context.Context, model.AnalysisCatalogReadAuthorization, sdk.ConversationAuthority, sdk.ConversationAuthority) error
}

// Producer is source proof provenance, never the actual request authority.
type sharedReadInput[T any] struct {
	Producer sdk.ConversationAuthority `json:"producer"`
	Input    T                         `json:"input"`
}

func sharedResultReadOperation(operation string) bool {
	switch operation {
	case "report_shared_result_read":
		return true
	case "report_shared_catalog_read":
		return true
	case "analysis_shared_result_read":
		return true
	case "analysis_shared_catalog_read":
		return true
	}
	return false
}

func validSharedProducer(producer, reader sdk.ConversationAuthority) bool {
	return validAuthority(producer, Scope{RuntimeID: reader.RuntimeID, WorkspaceID: reader.WorkspaceID})
}

func dispatchSharedResultRead(ctx context.Context, backend Backend, in request) (any, error) {
	switch in.Operation {
	case "report_shared_result_read":
		var input sharedReadInput[model.ReportQueryResultAuthorization]
		if decode(in.Payload, &input) != nil || !validSharedProducer(input.Producer, in.Authority) {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(SharedReportResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeSharedReportResultRead(ctx, input.Input, in.Authority, input.Producer)
		return err == nil, err
	case "report_shared_catalog_read":
		var input sharedReadInput[model.ReportCatalogReadAuthorization]
		if decode(in.Payload, &input) != nil || !validSharedProducer(input.Producer, in.Authority) {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(SharedReportResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeSharedReportCatalogRead(ctx, input.Input, in.Authority, input.Producer)
		return err == nil, err
	case "analysis_shared_result_read":
		var input sharedReadInput[model.AnalysisResultAuthorization]
		if decode(in.Payload, &input) != nil || !validSharedProducer(input.Producer, in.Authority) {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(SharedAnalysisResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeSharedAnalysisResultRead(ctx, input.Input, in.Authority, input.Producer)
		return err == nil, err
	case "analysis_shared_catalog_read":
		var input sharedReadInput[model.AnalysisCatalogReadAuthorization]
		if decode(in.Payload, &input) != nil || !validSharedProducer(input.Producer, in.Authority) {
			return nil, failure("bad_request")
		}
		owner, ok := backend.(SharedAnalysisResultReadSource)
		if !ok {
			return nil, failure("unavailable")
		}
		err := owner.AuthorizeSharedAnalysisCatalogRead(ctx, input.Input, in.Authority, input.Producer)
		return err == nil, err
	}
	return nil, failure("bad_request")
}

func (c *Client) AuthorizeSharedReportResultRead(ctx context.Context, in model.ReportQueryResultAuthorization, a, producer sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "report_shared_result_read", a, sharedReadInput[model.ReportQueryResultAuthorization]{Producer: producer, Input: in}, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) AuthorizeSharedReportCatalogRead(ctx context.Context, in model.ReportCatalogReadAuthorization, a, producer sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "report_shared_catalog_read", a, sharedReadInput[model.ReportCatalogReadAuthorization]{Producer: producer, Input: in}, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) AuthorizeSharedAnalysisResultRead(ctx context.Context, in model.AnalysisResultAuthorization, a, producer sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "analysis_shared_result_read", a, sharedReadInput[model.AnalysisResultAuthorization]{Producer: producer, Input: in}, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) AuthorizeSharedAnalysisCatalogRead(ctx context.Context, in model.AnalysisCatalogReadAuthorization, a, producer sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "analysis_shared_catalog_read", a, sharedReadInput[model.AnalysisCatalogReadAuthorization]{Producer: producer, Input: in}, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}
