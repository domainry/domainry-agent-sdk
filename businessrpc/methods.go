package businessrpc

import (
	"context"
	sdk "github.com/domainry/domainry-agent-sdk"
)

func (c *Client) BusinessCatalog(ctx context.Context, v sdk.ConversationBusinessCatalogQuery, a sdk.ConversationAuthority) (sdk.ConversationBusinessCatalogPage, error) {
	var result sdk.ConversationBusinessCatalogPage
	err := c.call(ctx, "catalog", a, v, &result)
	return result, err
}

func (c *Client) QueryBusinessRecords(ctx context.Context, v sdk.ConversationBusinessQuery, a sdk.ConversationAuthority) (sdk.ConversationBusinessRecordPage, error) {
	var result sdk.ConversationBusinessRecordPage
	err := c.call(ctx, "query", a, v, &result)
	return result, err
}

func (c *Client) GetBusinessRecord(ctx context.Context, v sdk.ConversationBusinessGet, a sdk.ConversationAuthority) (sdk.ConversationBusinessRecord, error) {
	var result sdk.ConversationBusinessRecord
	err := c.call(ctx, "get", a, v, &result)
	return result, err
}

func (c *Client) QueryRelatedBusinessRecords(ctx context.Context, v sdk.ConversationBusinessRelatedQuery, a sdk.ConversationAuthority) (sdk.ConversationBusinessRelatedPage, error) {
	var result sdk.ConversationBusinessRelatedPage
	err := c.call(ctx, "related", a, v, &result)
	return result, err
}

func (c *Client) RevalidateBusiness(ctx context.Context, v sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "revalidate", a, v, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) SealBusinessEvidence(ctx context.Context, v sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) (string, error) {
	var result string
	err := c.call(ctx, "seal", a, v, &result)
	return result, err
}

func (c *Client) AuthorizeBusinessAction(ctx context.Context, v sdk.ConversationBusinessAction, a sdk.ConversationAuthority) (sdk.ConversationToolAuthorization, error) {
	var result sdk.ConversationToolAuthorization
	err := c.call(ctx, "action_authorize", a, v, &result)
	return result, err
}

func (c *Client) InvokeBusinessAction(ctx context.Context, v sdk.ConversationBusinessActionRequest) (sdk.ConversationBusinessActionResult, error) {
	var result sdk.ConversationBusinessActionResult
	err := c.call(ctx, "action_invoke", v.Authority, v, &result)
	if err != nil {
		result.Status = "uncertain"
		result.ErrorCode = "external_result_unknown"
		return result, nil
	}
	return result, err
}

func (c *Client) ReconcileBusinessAction(ctx context.Context, v sdk.ConversationBusinessActionRequest) (sdk.ConversationBusinessActionResult, error) {
	var result sdk.ConversationBusinessActionResult
	err := c.call(ctx, "action_reconcile", v.Authority, v, &result)
	if err != nil {
		result.Status = "uncertain"
		result.ErrorCode = "external_result_unknown"
		return result, nil
	}
	return result, err
}

func (c *Client) RevalidateBusinessAction(ctx context.Context, v sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "action_revalidate", a, v, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) AuthorizeWorkflowStart(ctx context.Context, v sdk.ConversationWorkflowStart, a sdk.ConversationAuthority) (sdk.ConversationToolAuthorization, error) {
	var result sdk.ConversationToolAuthorization
	err := c.call(ctx, "workflow_authorize", a, v, &result)
	return result, err
}

func (c *Client) StartBusinessWorkflow(ctx context.Context, v sdk.ConversationWorkflowStartRequest) (sdk.ConversationWorkflowReceipt, error) {
	var result sdk.ConversationWorkflowReceipt
	err := c.call(ctx, "workflow_start", v.Authority, v, &result)
	if err != nil {
		result.Status = "uncertain"
		result.ErrorCode = "external_result_unknown"
		return result, nil
	}
	return result, err
}

func (c *Client) ReconcileBusinessWorkflow(ctx context.Context, v sdk.ConversationWorkflowStartRequest) (sdk.ConversationWorkflowReceipt, error) {
	var result sdk.ConversationWorkflowReceipt
	err := c.call(ctx, "workflow_reconcile", v.Authority, v, &result)
	if err != nil {
		result.Status = "uncertain"
		result.ErrorCode = "external_result_unknown"
		return result, nil
	}
	return result, err
}

func (c *Client) GetBusinessWorkflow(ctx context.Context, v sdk.ConversationWorkflowGet, a sdk.ConversationAuthority) (sdk.ConversationWorkflowState, error) {
	var result sdk.ConversationWorkflowState
	err := c.call(ctx, "workflow_get", a, v, &result)
	return result, err
}

func (c *Client) RevalidateBusinessWorkflow(ctx context.Context, v sdk.ConversationBusinessEvidence, a sdk.ConversationAuthority) error {
	var valid bool
	err := c.call(ctx, "workflow_revalidate", a, v, &valid)
	if err == nil && !valid {
		return failure("unavailable")
	}
	return err
}

func (c *Client) AuthorizeConversationTool(ctx context.Context, v sdk.ConversationToolRequest) (sdk.ConversationToolAuthorization, error) {
	var result sdk.ConversationToolAuthorization
	err := c.call(ctx, "tool_authorize", v.Authority, v, &result)
	return result, err
}
