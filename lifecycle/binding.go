// Package lifecycle defines the deployment-neutral Agent lifecycle extension.
// Hosts receive owner behavior, never Agent persistence repositories.
package lifecycle

import lifecyclecontract "github.com/domainry/domainry-lifecycle-sdk/contract"

// Binding is implemented by Agent bindings in both Module and SaaS modes.
// ArchiveWriter is Lifecycle-owned; the returned executor remains Agent-owned.
type Binding interface {
	LifecycleExecutor(lifecyclecontract.ArchiveWriter) lifecyclecontract.OwnerLifecycleExecutor
}

// SubjectBinding is optional because a remote Agent deployment may expose
// retention before it exposes data-subject execution. Embedded hosts use the
// returned owner handlers as opaque Lifecycle ports; they do not receive an
// Agent, Todo, or Knowledge repository.
type SubjectBinding interface {
	LifecycleSubjectHandlers() []lifecyclecontract.SubjectExecutionHandler
}
