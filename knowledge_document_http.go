package agentsdk

import actioncontract "github.com/domainry/domainry-foundation/action"

func KnowledgeDocumentPermission(operation string) *actioncontract.PermissionDefinition {
	label := map[string]string{"documents_transfer": "Copy or move library document", "documents_import_attachment": "Save attachment to library", "documents_upload": "Upload library document", "documents_list": "List library documents", "documents_get": "Read library document status", "documents_download": "Download library original", "documents_delete": "Delete library document"}[operation]
	if label == "" {
		return nil
	}
	return &actioncontract.PermissionDefinition{Key: ConversationActionPrefix + operation, Owner: AgentAuthorizationOwner, ResourceKey: AgentCapabilityConversation, OperationKey: operation, Label: label, Category: "Knowledge documents", LifecycleStatus: actioncontract.LifecycleActive}
}
