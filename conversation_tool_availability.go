package agentsdk

import "context"

// ConversationToolAvailability is an optional, live host policy for connection
// state and tool switches. The service asks only about registered tools that
// passed the host's catalog authorization. Return true only when this tool is
// enabled and its required connections are currently usable for this authority;
// return false for disabled, disconnected, expired or unknown connections. Local
// tools without a connection still need an explicit true when a policy is bound.
//
// Implementations must honor ctx, must not refresh credentials or perform tool
// effects here, and must not retain user credentials in model-visible metadata.
// This check does not replace authorization of the concrete action/resources.
// Without this optional policy, the existing host owns availability through
// ConversationTools and AuthorizeConversationTool, preserving legacy hosts.
type ConversationToolAvailability interface {
	ConversationToolAvailable(context.Context, ConversationAuthority, string) (bool, error)
}
