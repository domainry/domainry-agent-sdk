package persistence

import (
	"context"
	"time"

	agentstate "github.com/domainry/domainry-agent-sdk/state"
)

// LifecycleQuery carries host policy timing without exposing Agent storage.
// Eligibility and reference rules remain owned by the Agent module.
type LifecycleQuery struct {
	PolicyKey       string
	Now             time.Time
	Retention       time.Duration
	StatusRetention map[string]time.Duration
	Limit           int
}

type LifecycleCandidate struct {
	State      agentstate.AgentStateRecord
	ResourceID string
}

type AgentLifecycleRepository interface {
	ListLifecycleCandidates(context.Context, string, LifecycleQuery) ([]LifecycleCandidate, error)
	DeleteLifecycleCandidate(context.Context, string, LifecycleCandidate) (bool, error)
}
