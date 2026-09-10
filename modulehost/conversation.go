package modulehost

import agentsdk "github.com/domainry/domainry-agent-sdk"

// DeferredConversationHost is an optional persistence-host capability. When
// true, OpenModule must not start conversation workers or publish conversation
// services/adapters until BindApplicationHost supplies the live application
// ports. Hosts must finish binding before serving HTTP or accessing Conversations.
type DeferredConversationHost interface {
	DeferConversationHostBinding() bool
}

// ConversationApplicationHost supplies current authorization and optional
// business reads after the host's application services have been assembled.
// It is separate from TaskHost: ongoing conversations have no business task,
// process or entrypoint identity. A nil business source exposes no business tools.
// Hosts may also implement agentsdk.ConversationToolAvailability to supply live
// connection/tool switches without changing this required application contract.
type ConversationApplicationHost interface {
	ConversationAuthorizer() agentsdk.ConversationToolAuthorizer
	ConversationBusinessSource() agentsdk.ConversationBusinessSource
}
