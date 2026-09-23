package modulehost

import (
	"context"

	agentsdk "github.com/domainry/domainry-agent-sdk"
	sharedartifact "github.com/domainry/domainry-foundation/artifact"
	ormmigration "github.com/domainry/domainry-orm/migration"
	"github.com/domainry/domainry-orm/sqlhost"
)

// Host contains the narrow Runtime-owned facilities needed by an embedded
// Agent module. Agent owns its schema and repositories while borrowing the
// host pool, dialect and sole migration ledger.
type Host interface {
	RuntimeID() string
	Database() Database
	Dialect() Dialect
	Migrations() MigrationRegistrar
}

// ArtifactHost is the optional deployment-owned shared Artifact kernel used by
// conversation attachments and generated files. Module implementations must
// fail closed for those capabilities when the host does not provide all three
// ports; they must not recreate owner-private metadata tables.
type ArtifactHost interface {
	ArtifactStore() sharedartifact.ManagedStore
	ArtifactContentStore() sharedartifact.ContentStore
	ArtifactContentWriter() sharedartifact.ContentWriter
}

type Executor = sqlhost.Executor
type Queryer = sqlhost.Queryer
type Database = sqlhost.Database

// Dialect is the host-selected domainry-orm renderer. Agent repositories must
// not infer a dialect from driver names or construct a second connection.
type Dialect interface {
	Identifier(string) string
	Table(string) string
	Placeholder(int) string
}

type SchemaMigration = ormmigration.Migration
type SchemaBaseline = ormmigration.Baseline
type SchemaTable = ormmigration.Table
type SchemaColumn = ormmigration.Column
type SchemaIndex = ormmigration.Index

type MigrationRegistrar interface {
	Driver() string
	Schema() string
	ApplyOwnedMigrations(context.Context, string, []SchemaMigration) error
}

type Factory interface {
	OpenModule(context.Context, agentsdk.ApplicationRef, Host) (agentsdk.Binding, error)
}
