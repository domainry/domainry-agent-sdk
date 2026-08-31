package agentsdk

import (
	"sort"
	"strings"
)

const (
	AgentToolQueryRecords  = "query_records"
	AgentToolGetRecord     = "get_record"
	AgentToolInvokeAction  = "invoke_action"
	AgentToolReadRecord    = "readRecord"
	AgentToolCreateRecord  = "createRecord"
	AgentToolUpdateRecord  = "updateRecord"
	AgentToolDeleteRecord  = "deleteRecord"
	AgentToolCallConnector = "callConnector"
	AgentToolSendMessage   = "sendMessage"
	AgentToolSendEmail     = "sendEmail"
)

// AgentTool describes a provider-visible Agent tool and the static
// prerequisites that authoring and execution must enforce. Runtime-owned
// authorization still makes the final decision for every concrete call.
type AgentTool struct {
	Key                    string
	RequiresAllowedObjects bool
	RequiresAllowedActions bool
	Writes                 bool
}

var agentTools = map[string]AgentTool{
	AgentToolQueryRecords:  {Key: AgentToolQueryRecords, RequiresAllowedObjects: true},
	AgentToolGetRecord:     {Key: AgentToolGetRecord, RequiresAllowedObjects: true},
	AgentToolInvokeAction:  {Key: AgentToolInvokeAction, RequiresAllowedObjects: true, RequiresAllowedActions: true, Writes: true},
	AgentToolReadRecord:    {Key: AgentToolReadRecord},
	AgentToolCreateRecord:  {Key: AgentToolCreateRecord, Writes: true},
	AgentToolUpdateRecord:  {Key: AgentToolUpdateRecord, Writes: true},
	AgentToolDeleteRecord:  {Key: AgentToolDeleteRecord, Writes: true},
	AgentToolCallConnector: {Key: AgentToolCallConnector, Writes: true},
	AgentToolSendMessage:   {Key: AgentToolSendMessage, Writes: true},
	AgentToolSendEmail:     {Key: AgentToolSendEmail, Writes: true},
}

func LookupAgentTool(key string) (AgentTool, bool) {
	tool, exists := agentTools[strings.TrimSpace(key)]
	return tool, exists
}

func AgentToolKeys() []string {
	result := make([]string, 0, len(agentTools))
	for key := range agentTools {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
