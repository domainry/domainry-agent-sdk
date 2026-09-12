package agentsdk

import actioncontract "github.com/domainry/domainry-foundation/action"

// These permissions cover explicit file-management APIs, not model tools.
func ConversationAttachmentPermission(operation string) *actioncontract.PermissionDefinition {
	label := map[string]string{"attachments_upload": "Upload private conversation attachments", "attachments_index": "Index private conversation attachments", "attachments_list": "List private conversation attachments", "attachments_get": "Read private attachment metadata", "attachments_download": "Download private conversation attachments", "attachments_delete": "Delete private conversation attachments"}[operation]
	if operation == "attachments_check_index" {
		label = "Check existing private attachment index work"
	}
	if label == "" {
		return nil
	}
	return &actioncontract.PermissionDefinition{Key: ConversationActionPrefix + operation, Owner: AgentAuthorizationOwner, ResourceKey: AgentCapabilityConversation, OperationKey: operation, Label: label, Category: "Conversation attachments", LifecycleStatus: actioncontract.LifecycleActive}
}
