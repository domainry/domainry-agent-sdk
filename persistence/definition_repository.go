package persistence

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type DefinitionSnapshot struct {
	SchemaVersion string                                  `json:"schema_version"`
	SchemaHash    string                                  `json:"schema_hash"`
	SourceKind    string                                  `json:"source_kind"`
	SourceID      string                                  `json:"source_id"`
	Skills        []agentsdk.SkillSchema                  `json:"skills,omitempty"`
	Agents        []agentsdk.AgentSchema                  `json:"agents,omitempty"`
	Tasks         []agentsdk.AgentTaskDefinition          `json:"tasks,omitempty"`
	Entrypoints   []agentsdk.AgentEntrypointAssignment    `json:"entrypoints,omitempty"`
	Principals    []agentsdk.AgentServicePrincipalBinding `json:"service_principals,omitempty"`
}

type DefinitionRepository interface {
	SyncDefinitions(context.Context, DefinitionSnapshot) error
	DefinitionSnapshot(context.Context) (DefinitionSnapshot, error)
}
