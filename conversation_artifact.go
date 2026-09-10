package agentsdk

import (
	"context"
	"time"
)

// Artifact revisions are immutable. An ID identifies a document; Version
// identifies the exact content reviewed, edited or exported by a user.
type ConversationArtifact struct {
	ID                   string    `json:"id"`
	Version              int64     `json:"version"`
	Title                string    `json:"title"`
	Kind                 string    `json:"kind"` // markdown, table, chart
	SHA256               string    `json:"sha256"`
	Bytes                int       `json:"bytes"`
	SourceConversationID string    `json:"source_conversation_id,omitempty"`
	SourceRunID          string    `json:"source_run_id,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Cells remain strings so decimal amounts and large integers are not rounded
// by a JSON or browser float conversion. An absent cell is represented by nil.
type ConversationArtifactColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"` // text, number, date
}
type ConversationArtifactTable struct {
	Columns []ConversationArtifactColumn `json:"columns"`
	Rows    [][]*string                  `json:"rows"`
}
type ConversationArtifactChart struct {
	Type     string   `json:"type"` // bar, line
	XColumn  string   `json:"x_column"`
	YColumns []string `json:"y_columns"`
}
type ConversationArtifactContent struct {
	Kind     string                     `json:"kind"`
	Markdown string                     `json:"markdown,omitempty"`
	Table    *ConversationArtifactTable `json:"table,omitempty"`
	Chart    *ConversationArtifactChart `json:"chart,omitempty"`
}
type ConversationArtifactVersion struct {
	Artifact ConversationArtifact        `json:"artifact"`
	Content  ConversationArtifactContent `json:"content"`
}
type ConversationArtifactCreate struct {
	ClientID string                      `json:"client_id"`
	Title    string                      `json:"title"`
	Content  ConversationArtifactContent `json:"content"`
	// Supply both source IDs or neither; provenance is resolved by the server.
	SourceConversationID string `json:"source_conversation_id,omitempty"`
	SourceRunID          string `json:"source_run_id,omitempty"`
}
type ConversationArtifactTextEdit struct {
	Find    string `json:"find"`
	Replace string `json:"replace"`
}
type ConversationArtifactCellEdit struct {
	Row    int     `json:"row"` // zero-based original row index
	Column string  `json:"column"`
	Value  *string `json:"value"`
}
type ConversationArtifactPatch struct {
	Title   *string                        `json:"title,omitempty"`
	Content *ConversationArtifactContent   `json:"content,omitempty"`
	Text    []ConversationArtifactTextEdit `json:"text,omitempty"`
	Cells   []ConversationArtifactCellEdit `json:"cells,omitempty"`
}
type ConversationArtifactEdit struct {
	ClientID        string                    `json:"client_id"`
	ExpectedVersion int64                     `json:"expected_version"`
	Patch           ConversationArtifactPatch `json:"patch"`
}
type ConversationArtifactQuery struct {
	Query                string `json:"query,omitempty"`
	SourceConversationID string `json:"source_conversation_id,omitempty"`
	Cursor               string `json:"cursor,omitempty"`
	Limit                int    `json:"limit,omitempty"`
}
type ConversationArtifactPage struct {
	Items      []ConversationArtifact `json:"items"`
	NextCursor string                 `json:"next_cursor,omitempty"`
	Complete   bool                   `json:"complete"`
	Omitted    bool                   `json:"omitted,omitempty"`
}
type ConversationArtifactVersions struct {
	Items      []ConversationArtifact `json:"items"`
	NextBefore int64                  `json:"next_before,omitempty"`
	Complete   bool                   `json:"complete"`
	Omitted    bool                   `json:"omitted,omitempty"`
}
type ConversationArtifactExportRequest struct {
	ClientID string `json:"client_id"`
	Version  int64  `json:"version"`
	Format   string `json:"format"` // markdown or csv
}
type ConversationArtifactExport struct {
	ID               string     `json:"id"`
	ArtifactID       string     `json:"artifact_id"`
	Version          int64      `json:"version"`
	Format           string     `json:"format"`
	Filename         string     `json:"filename"`
	ContentType      string     `json:"content_type"`
	SHA256           string     `json:"sha256"`
	Bytes            int        `json:"bytes"`
	FormulaGuarded   bool       `json:"formula_guarded,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	Downloads        int64      `json:"downloads"`
	LastDownloadedAt *time.Time `json:"last_downloaded_at,omitempty"`
}
type ConversationArtifactDownload struct {
	Export ConversationArtifactExport `json:"export"`
	Data   []byte                     `json:"data"`
}

type ConversationArtifactRead struct {
	ID       string   `json:"id"`
	Version  int64    `json:"version,omitempty"`
	Offset   int      `json:"offset,omitempty"`
	MaxBytes int      `json:"max_bytes,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Columns  []string `json:"columns,omitempty"`
}

// A tool page is explicitly partial and is not an editable full-content DTO.
type ConversationArtifactReadResult struct {
	Artifact    ConversationArtifact       `json:"artifact"`
	Markdown    string                     `json:"markdown,omitempty"`
	Table       *ConversationArtifactTable `json:"table,omitempty"`
	Chart       *ConversationArtifactChart `json:"chart,omitempty"`
	Offset      int                        `json:"offset"`
	NextOffset  int                        `json:"next_offset"`
	Complete    bool                       `json:"complete"`
	Projected   bool                       `json:"projected,omitempty"`
	RowTooLarge bool                       `json:"row_too_large,omitempty"`
}

type ConversationArtifactService interface {
	Artifacts(context.Context, ConversationArtifactQuery, ConversationAuthority) (ConversationArtifactPage, error)
	Artifact(context.Context, string, int64, ConversationAuthority) (ConversationArtifactVersion, error)
	ArtifactVersions(context.Context, string, int64, int, ConversationAuthority) (ConversationArtifactVersions, error)
	CreateArtifact(context.Context, ConversationArtifactCreate, ConversationAuthority) (ConversationArtifactVersion, error)
	EditArtifact(context.Context, string, ConversationArtifactEdit, ConversationAuthority) (ConversationArtifactVersion, error)
	ExportArtifact(context.Context, string, ConversationArtifactExportRequest, ConversationAuthority) (ConversationArtifactExport, error)
	DownloadArtifact(context.Context, string, ConversationAuthority) (ConversationArtifactDownload, error)
}

// Large bodies may be stored by the host outside the metadata database. The
// host must provide immutable, owner-scoped objects; references are opaque and
// never accepted from model or browser inputs. Content hash is verified again
// when reading. Orphan object retention belongs to the host's storage policy.
// Put must return the same reference for the same owner and content hash,
// including retries after restart, so mutation receipts remain replayable.
type ConversationArtifactStorage interface {
	PutArtifactContent(context.Context, string, []byte, ConversationAuthority) (string, error)
	ReadArtifactContent(context.Context, string, ConversationAuthority) ([]byte, error)
}
