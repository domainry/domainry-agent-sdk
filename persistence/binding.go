package persistence

// Binding exposes Agent-owned repositories without exposing concrete stores
// or the selected database topology.
type Binding interface {
	AgentStateRepository() AgentStateRepository
	AgentTaskRunRepository() AgentTaskRunRepository
}

type DefinitionBinding interface {
	DefinitionRepository() DefinitionRepository
}
