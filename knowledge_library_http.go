package agentsdk

import actioncontract "github.com/domainry/domainry-foundation/action"

func KnowledgeLibraryPermission(operation string) *actioncontract.PermissionDefinition {
	label := map[string]string{"libraries_sources": "List library knowledge sources", "libraries_bind_source": "Bind library knowledge source", "libraries_create": "Create knowledge libraries", "libraries_list": "List own knowledge libraries", "libraries_get": "Read knowledge library", "libraries_update": "Update knowledge library settings", "libraries_members": "List knowledge library members", "libraries_set_member": "Set knowledge library member", "libraries_remove_member": "Remove knowledge library member"}[operation]
	if label == "" {
		return nil
	}
	return &actioncontract.PermissionDefinition{Key: ConversationActionPrefix + operation, Owner: AgentAuthorizationOwner, ResourceKey: AgentCapabilityConversation, OperationKey: operation, Label: label, Category: "Knowledge libraries", LifecycleStatus: actioncontract.LifecycleActive}
}
