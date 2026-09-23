package persistence

import (
	"context"
	"encoding/json"

	agentsdk "github.com/domainry/domainry-agent-sdk"
)

// This record stays within the trusted host boundary. Source policy and blob
// references are never taken from an artifact_create/edit tool argument.
type ConversationArtifactRecord struct {
	Artifact agentsdk.ConversationArtifact `json:"artifact"`
	Body     json.RawMessage               `json:"body,omitempty"`
	BodyRef  string                        `json:"body_ref,omitempty"`
	Sources  *agentsdk.ConversationSources `json:"sources"`
}

// ArtifactWrite is a fully validated replacement revision assembled by the
// application. An empty ID creates version 1; updates require ExpectedVersion.
type ConversationArtifactWrite struct {
	// Optional digest of the original logical command, computed by the trusted
	// application. It binds retries to the patch, not merely its resulting body.
	RequestSHA256   string                     `json:"request_sha256,omitempty"`
	ClientID        string                     `json:"client_id"`
	ExpectedVersion int64                      `json:"expected_version"`
	Record          ConversationArtifactRecord `json:"record"`
}
type ConversationArtifactExportWrite struct {
	RequestSHA256 string                              `json:"request_sha256,omitempty"`
	TTLSeconds    int64                               `json:"ttl_seconds"`
	ClientID      string                              `json:"client_id"`
	Export        agentsdk.ConversationArtifactExport `json:"export"`
	// Content is the exact immutable export prepared by the trusted application.
	// It is written to deployment-owned blob storage and never persisted in SQL
	// or echoed into an idempotency receipt.
	Content []byte `json:"-"`
}

type ConversationArtifactRepository interface {
	Artifacts(context.Context, agentsdk.ConversationArtifactQuery, agentsdk.ConversationAuthority) (agentsdk.ConversationArtifactPage, error)
	ArtifactRecord(context.Context, string, int64, agentsdk.ConversationAuthority) (ConversationArtifactRecord, error)
	ArtifactVersions(context.Context, string, int64, int, agentsdk.ConversationAuthority) (agentsdk.ConversationArtifactVersions, error)
	SaveArtifact(context.Context, ConversationArtifactWrite, agentsdk.ConversationAuthority) (ConversationArtifactRecord, error)
	SaveArtifactExport(context.Context, ConversationArtifactExportWrite, agentsdk.ConversationAuthority) (agentsdk.ConversationArtifactExport, error)
	ArtifactExport(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationArtifactExport, error)
	ArtifactExportContent(context.Context, string, agentsdk.ConversationAuthority) ([]byte, error)
	RecordArtifactDownload(context.Context, string, agentsdk.ConversationAuthority) (agentsdk.ConversationArtifactExport, error)
}

// A trusted host prepares content after validating the frozen call and its
// source permissions. Persistence rechecks the ledger/lease/confirmation and
// commits the effect and tool result in one transaction. Models never supply
// this envelope, source records, request digests or storage references.
type ConversationArtifactToolMutation struct {
	Write  *ConversationArtifactWrite       `json:"write,omitempty"`
	Export *ConversationArtifactExportWrite `json:"export,omitempty"`
}
type ConversationArtifactMutationRepository interface {
	ApplyArtifactTool(context.Context, agentsdk.ConversationToolRequest, ConversationArtifactToolMutation) (agentsdk.ConversationToolResult, error)
}
