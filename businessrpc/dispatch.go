package businessrpc

import (
	"bytes"
	"context"
	sdk "github.com/domainry/domainry-agent-sdk"
)

func dispatch(ctx context.Context, b Backend, in request) (any, error) {
	if sharedResultReadOperation(in.Operation) {
		return dispatchSharedResultRead(ctx, b, in)
	}
	if analysisOperation(in.Operation) {
		return dispatchAnalysis(ctx, b, in)
	}
	if reportOperation(in.Operation) {
		return dispatchReport(ctx, b, in)
	}
	switch in.Operation {
	case "business_result_read":
		var v sdk.ConversationBusinessEvidence
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := sdk.AuthorizeBusinessResultRead(ctx, b, v, in.Authority)
		return err == nil, err
	case "catalog":
		var v sdk.ConversationBusinessCatalogQuery
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.BusinessCatalog(ctx, v, in.Authority)
	case "query":
		var v sdk.ConversationBusinessQuery
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.QueryBusinessRecords(ctx, v, in.Authority)
	case "get":
		var v sdk.ConversationBusinessGet
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.GetBusinessRecord(ctx, v, in.Authority)
	case "related":
		var v sdk.ConversationBusinessRelatedQuery
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.QueryRelatedBusinessRecords(ctx, v, in.Authority)
	case "revalidate":
		var v sdk.ConversationBusinessEvidence
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := b.RevalidateBusiness(ctx, v, in.Authority)
		return err == nil, err
	case "seal":
		var v sdk.ConversationBusinessEvidence
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.SealBusinessEvidence(ctx, v, in.Authority)
	case "action_authorize":
		var v sdk.ConversationBusinessAction
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		// The empty in-process action is a capability probe. RawMessage(nil)
		// encodes as null, which must not turn that probe into a concrete action
		// on the receiving side. Concrete actions retain their exact Data bytes.
		if v.ObjectKey == "" && v.ActionKey == "" && v.Version == "" && v.RecordID == "" && bytes.Equal(bytes.TrimSpace(v.Data), []byte("null")) {
			v.Data = nil
		}
		return b.AuthorizeBusinessAction(ctx, v, in.Authority)
	case "action_invoke":
		var v sdk.ConversationBusinessActionRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		if v.Authority != in.Authority {
			return nil, failure("bad_request")
		}
		return b.InvokeBusinessAction(ctx, v)
	case "action_reconcile":
		var v sdk.ConversationBusinessActionRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		if v.Authority != in.Authority {
			return nil, failure("bad_request")
		}
		return b.ReconcileBusinessAction(ctx, v)
	case "action_revalidate":
		var v sdk.ConversationBusinessEvidence
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := b.RevalidateBusinessAction(ctx, v, in.Authority)
		return err == nil, err
	case "workflow_authorize":
		var v sdk.ConversationWorkflowStart
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.AuthorizeWorkflowStart(ctx, v, in.Authority)
	case "workflow_start":
		var v sdk.ConversationWorkflowStartRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		if v.Authority != in.Authority {
			return nil, failure("bad_request")
		}
		return b.StartBusinessWorkflow(ctx, v)
	case "workflow_reconcile":
		var v sdk.ConversationWorkflowStartRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		if v.Authority != in.Authority {
			return nil, failure("bad_request")
		}
		return b.ReconcileBusinessWorkflow(ctx, v)
	case "workflow_get":
		var v sdk.ConversationWorkflowGet
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		return b.GetBusinessWorkflow(ctx, v, in.Authority)
	case "workflow_revalidate":
		var v sdk.ConversationBusinessEvidence
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		err := b.RevalidateBusinessWorkflow(ctx, v, in.Authority)
		return err == nil, err
	case "tool_authorize":
		var v sdk.ConversationToolRequest
		if decode(in.Payload, &v) != nil {
			return nil, failure("bad_request")
		}
		if v.Authority != in.Authority {
			return nil, failure("bad_request")
		}
		return b.AuthorizeConversationTool(ctx, v)
	}
	return nil, failure("bad_request")
}
