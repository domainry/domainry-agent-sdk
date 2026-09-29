package agentsdk

import (
	"encoding/json"
	"strings"
	"testing"

	actioncontract "github.com/domainry/domainry-foundation/action"
)

func TestConversationToolAuthorizationActionsCompileBoundProductTools(t *testing.T) {
	definitions := []ConversationToolDefinition{
		{
			Key: "crm_search_accounts", Version: "1", Description: "Search accounts.",
			InputSchema: json.RawMessage(`{"type":"object"}`), OutputSchema: json.RawMessage(`{"type":"object"}`),
			ActionKey: ConversationToolActionPrefix + "crm_search_accounts", Effect: "read", Idempotency: "natural", TimeoutMillis: 1000, MaxOutputBytes: 1024,
		},
	}
	actions, err := ConversationToolAuthorizationActions(definitions)
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("Actions=%d", len(actions))
	}
	action := actions[0]
	if action.Key != definitions[0].ActionKey || action.Owner != AgentAuthorizationOwner || action.SourceKind != "agent_tool" || action.OperationKey != definitions[0].Key || action.EffectClass != actioncontract.EffectRead {
		t.Fatalf("Action=%+v", action)
	}
	if action.Permission == nil || action.Permission.Key != action.Key || action.Permission.ResourceKey != "agent.conversation_tools" || action.Permission.OperationKey != definitions[0].Key {
		t.Fatalf("Permission=%+v", action.Permission)
	}
	if len(action.NonHTTP) != 1 || action.NonHTTP[0] != (actioncontract.NonHTTPBinding{Kind: "sdk", InvocationKey: action.Key}) {
		t.Fatalf("non-HTTP bindings=%+v", action.NonHTTP)
	}
}

func TestConversationToolAuthorizationActionsRejectAmbiguousOrInvalidDefinitions(t *testing.T) {
	valid := ConversationToolDefinition{Key: "crm_search_accounts", ActionKey: ConversationToolActionPrefix + "crm_search_accounts", Effect: "read", Idempotency: "natural"}
	tests := map[string][]ConversationToolDefinition{
		"empty key":           {{ActionKey: ConversationToolActionPrefix, Effect: "read", Idempotency: "natural"}},
		"wrong Action key":    {{Key: valid.Key, ActionKey: ConversationToolActionPrefix + "other", Effect: "read", Idempotency: "natural"}},
		"duplicate key":       {valid, valid},
		"invalid effect":      {{Key: valid.Key, ActionKey: valid.ActionKey, Effect: "delete", Idempotency: "natural"}},
		"invalid idempotency": {{Key: valid.Key, ActionKey: valid.ActionKey, Effect: "read", Idempotency: "maybe"}},
		"natural write":       {{Key: valid.Key, ActionKey: valid.ActionKey, Effect: "write", Idempotency: "natural"}},
	}
	for name, definitions := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ConversationToolAuthorizationActions(definitions); err == nil || strings.TrimSpace(err.Error()) == "" {
				t.Fatalf("invalid definitions accepted: %+v", definitions)
			}
		})
	}
}
