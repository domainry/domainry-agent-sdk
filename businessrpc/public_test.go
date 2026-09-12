package businessrpc_test

import (
	sdk "github.com/domainry/domainry-agent-sdk"
	"github.com/domainry/domainry-agent-sdk/businessrpc"
)

// External-package compilation verifies consumers need no implementation or
// internal adapter import to bind the public client to existing Agent ports.
var (
	_ sdk.ConversationBusinessSource         = (*businessrpc.Client)(nil)
	_ sdk.ConversationBusinessRelationSource = (*businessrpc.Client)(nil)
	_ sdk.ConversationBusinessActionSource   = (*businessrpc.Client)(nil)
	_ sdk.ConversationBusinessWorkflowSource = (*businessrpc.Client)(nil)
	_ sdk.ConversationBusinessEvidenceSealer = (*businessrpc.Client)(nil)
	_ sdk.ConversationToolAuthorizer         = (*businessrpc.Client)(nil)
	_ businessrpc.ReportReader               = (*businessrpc.Client)(nil)
	_ businessrpc.AnalysisReader             = (*businessrpc.Client)(nil)
)
