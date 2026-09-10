# domainry-agent-sdk

Deployment-neutral contract for Domainry Agent execution. Agent implementations own definitions, dialog/session behavior, proposals and provider-execution state, including task, interactive-run, claim/lease/fencing and tool-ledger state. Runtime owns workflow state, current business authorization, Runtime tool/effect policy and business records.

The SDK contains no Runtime or implementation dependency. Project composition selects a Module or SaaS Factory explicitly.

## Package layout

- The root package is the stable `Factory`, `Binding`, runner, schema, tool-catalog, and dialog-state entrypoint.
- `persistence` owns Agent execution-state application ports, persistence capabilities and the shared redacted task-run view. Agent's HTTP adapter and worker assembly consume `ExecutionStateBinding`; Runtime does not mutate Agent repositories or implement a second task/interactive state machine.
- `modulehost` describes infrastructure borrowed by an embedded Agent module and the narrow Runtime fact/effect ports used by Agent application services.
- `saashost` describes the SaaS composition boundary.
- `contracttest` contains deployment-parity tests; `state` contains shared Agent state values.

Concrete SQL stores remain in the Agent implementation repository; the SDK exposes only their deployment-neutral contracts through `persistence`.

`ConversationBinding` is an optional extension exposing `ConversationService` for
durable personal conversations. It is separate from `InteractiveRunner` and its
legacy session projection. `conversation.go` owns messages, run/event state,
explicit personal memory, summary and stateless model contracts;
`conversation_http.go` owns the `/agent/conversations` Action/OpenAPI manifest.
`persistence.ConversationRepository` owns atomic enqueue, lease/fencing, frozen
model inputs, draft deltas and summary persistence ports. Pure-text Conversation remains supported.
`ConversationStreamingModel` optionally streams text through a synchronous commit callback;
`conversation.stream.v1` advertises this capability. Drafts belong to one attempt and
are separate from complete messages. `ConversationStatusProvider` supplies local
readiness without invoking the model. Descriptors may advertise Conversation alone;
task capabilities still require the complete start/poll/cancel trio.

`ConversationAttachmentService` optionally exposes private conversation file
upload, list, detail, original-byte download and deletion. Hosts supply both
`ConversationAttachmentStorage` and a current `ConversationAttachmentAuthorizer`;
these file operations have separate Identity actions and are not model tools.
Uploads accept a client idempotency key, filename and up to 16 MiB of bytes;
authority and indexing permissions are never file input. Browser HTTP uses raw
binary, while trusted SaaS RPC uses bounded base64. Storage must permanently
fence late writes after deletion, including across restarts. Attachment metadata
and durable cleanup jobs use the optional persistence attachment port. A stored
file is not indexed or shared; personal libraries and shared-library membership
use the separate contract below.

`KnowledgeLibraryService` optionally manages a unique personal library and
shared libraries with reader/editor/manager memberships. Hosts inject a
`KnowledgeLibraryAuthorizer` for live Identity action checks and validation of
active workspace users. The repository independently enforces current library
membership and serializes settings/member mutations with the expected library
revision, preserving at least one manager. Libraries belong to the trusted
Runtime/Workspace; personal libraries never accept other members. These seven
management operations are HTTP/SDK capabilities, not model tools. Library
creation does not bind a remote source, upload documents, or change existing
knowledge-search permissions; document ownership and indexing use the separate
optional `KnowledgeDocumentService`.

`ConversationLibraryKnowledgeSource` optionally extends the revalidating
knowledge source with a readable-library catalog and scoped search/read.
`LibraryKnowledgeConversationTools` publishes `knowledge_libraries` plus v2
search/read schemas with optional `library_id`; `KnowledgeConversationTools`
keeps its legacy v1 schemas. Library IDs only select trusted host bindings:
implementations check current membership, archive state and Identity action
access before/after remote calls and whenever saved evidence is reused.
Evidence and citations preserve `library_id`; the same document ID may exist
in different libraries. Hosts must isolate each library's remote KB from other
libraries and the legacy source. `KnowledgeLibrary.knowledge_configured` reports
a binding's presence only, not remote availability or indexing completion.

`KnowledgeDocumentService` exposes library document upload/list/get/download/delete
via Module HTTP and the SaaS SDK, with separate `KnowledgeDocumentPermission`
actions and current library role checks. `KnowledgeDocumentStorage` binds immutable
originals to Runtime/Workspace/Library, independently of uploader membership.
`ManagedKnowledgeDocumentSource` adds explicit passage projection and a stable
physical source identity; implementations must exclude unmapped response fields.
Host activation is explicit, requires a dedicated source and persistent management
markers, and does not automatically convert conversation attachments or enable model writes.
`KnowledgeLibrary.documents_configured` reports whether the host has configured
document management for that library. It is separate from read-only knowledge
binding availability and does not grant Identity actions or prove index readiness.

`KnowledgeAttachmentImportService` optionally saves an owned conversation attachment
as an independent library document. Its HTTP/SaaS input carries only source IDs,
expected attachment revision and client receipt. The host checks the distinct import
action, target upload and source download permissions; the repository rechecks
source state inside target reservation/commit. Private provenance remains server-side.
A committed copy survives source deletion; subsequent access follows target-library
permissions. Retrying a completed receipt never reads the former private source.

`KnowledgeDocumentSource` is a separate trusted host port for pushing original
bytes, inspecting the actual upstream indexing state, and requesting deletion.
It is not a model tool or public file-management endpoint. Applications must
authorize recorded document/library ownership and persist lifecycle work;
acknowledgements do not prove indexing or deletion is complete. Missing status
is scoped to current upstream permissions and is not a global existence oracle.

`conversation.execution.v1` is an optional execution extension:
`ConversationAgentModel` streams one normally terminated model step;
`ConversationToolHost` provides a live catalog, authorization, invocation and
reconciliation. `persistence.ConversationExecutionRepository` freezes steps and
records logical tool calls under a lease/fence. Model-native continuation state
is server-only; `ConversationRun.Steps` contains bounded public previews.
`PersonalConversationTools` and `ConversationToolActions` are the source-owned
catalog and permission manifest for implemented read tools. Registering an
Action does not grant it to a user. These additions do not add mandatory methods
to existing `Binding`, `ConversationModel` or `ConversationRepository` ports.

`ConversationToolAvailability` optionally supplies live connection state and
tool switches. Agent filters the authorized host catalog through this policy
before freezing each model step, checks it again after concrete authorization,
and prevents disconnected business/knowledge sources from being called during
historical-result revalidation. Return true only for enabled tools with usable
connections in the supplied runtime/workspace/user scope, including explicit
true for local tools. Return false for disabled, disconnected, expired or
unknown state; errors fail closed with a sanitized public code. Checks have a
bounded context and must not refresh credentials or perform tool effects.
The policy only removes registered tools; it cannot grant permissions or change
their schemas, effects, timeouts, result limits or idempotency strategies.
Existing hosts without this optional port retain their host-owned catalog and
authorization behavior. Account management/refresh remains a host responsibility.

`ConversationBusinessSource` supplies scoped business discovery and reads,
including typed filters, field selection, page/cursor results and revalidation
of saved evidence. Cursors are read positions, not grants; hosts still resolve
current Identity and enforce record/field policy. A missing total is unknown.

An optional `ConversationBusinessEvidenceSealer` can attest the immutable read
result before Agent stores it. Its `host_proof` is an integrity proof, not a
credential or grant. The host must verify the complete evidence before sealing
and recheck current access on every reuse. A sealer failure prevents the result
from entering model context. Existing sources may omit this extension and keep
their existing revalidation semantics; frozen content must never be rewritten
to pretend a historical read is current.

Module hosts assembled in two stages can implement
`modulehost.DeferredConversationHost`. When enabled, the module opens storage
and definitions first; `Conversations()` returns nil until the startup-only
`BindApplicationHost` receives `modulehost.ConversationApplicationHost` with a
current authorizer and optional business source. Conversation workers and HTTP
adapters are created only then. Finish binding before serving HTTP. Ordinary
hosts without this optional marker keep their existing startup behavior.

Run `go test ./...` before publishing an immutable SDK version.

Business catalog selectors use `kind=objects` (default), `actions`, `workflows`, or `relations`.
Object details expose readable fields; `readable=false` is discovery-only.
For operations, `object_key` filters a bounded list, and `action_key` or
`workflow_key` expands the payload schema. A null schema is not a declared
empty payload. `ValidSelector` checks selector combinations; hosts must still
validate size limits and live permissions. Discovery is never execution approval.

`ConversationBusinessWorkflowSource` optionally adds `workflow_start` and
`workflow_get`. Expand a workflow's catalog entry for its `execution_version`
and input schema. Start receives only that version, the declared key and business
input from the model; Agent supplies trusted authority, durable confirmation and
stable invocation metadata separately. An `accepted` receipt is not completion:
retain its process ID and query the host's actual state and business outcome.
Terminal states can also indicate rejection, cancellation or failure. Raw workflow
variables and internal node outputs do not belong in the public state projection.

Workflow hosts own instance and participant authorization and must recheck it on
every query and historical reuse. Start reconciliation checks owned durable
receipts; it never reclaims an existing unresolved start. Only a known-absent
operation may enter an atomic start that refuses to reclaim existing work. A
signed historical progress snapshot may retain its original waiting state after
completion, provided its integrity and current access are independently verified.

`kind=relations` requires the source `object_key` and pages published forward /
reverse relationship keys. Object summaries may expose a bounded `relations`
list and `relations_next_cursor`. The optional `ConversationBusinessRelationSource`
adds `query_related_records`; unsupported hosts do not publish that tool.
Each call follows one relationship from one readable record, with at most 25
targets, typed filters, fields and a cursor. Hosts bind cursors to the source and
relationship, and reauthorize source rows, relationship fields and targets on
every page and during `RevalidateBusiness`. No recursive expansion or model-
selected target object is accepted; further hops consume ordinary run budgets.

`ConversationFactory.ConversationEnabled` allows hosts to start configured
persistent conversations even when their manifest has no legacy Agent/Task
entries. It is an optional startup declaration, not a change to those entrypoints.

`ConversationBusinessActionSource` optionally adds `invoke_action`. Executable
action details expose `execution_version`, target kind and the actual payload
schema, including optimistic-concurrency fields when required. The generic
catalog projection version is not an execution version. Model arguments contain
only the declared target, action version and business payload; trusted authority,
frozen confirmation and stable idempotency metadata arrive separately from Agent.

The host must enforce live business and record permissions, exact contracts and
its existing action rules. Agent currently requires confirmation of each concrete
business write. Personal memory, todo or artifact write scopes do not authorize
business mutations. A confirmed request cannot satisfy an additional host approval
requirement by itself.

Action results contain durable invocation and affected-record references, not raw
handler output. An uncertain result needs reconciliation against the owned host
receipt. Reconciliation must not reclaim or repeat an existing unresolved action;
a conclusively absent action may start only through an atomic create-or-replay
boundary that cannot reclaim an intervening execution. Historical acknowledgement
revalidation is read-only and checks the entire receipt and current access.

## Browser gateway

`browsergateway.NewHandler` composes the SDK Agent and Identity bindings into a same-origin browser boundary. The caller provides static files and may supply its existing HTTP router so conversation requests retain host admission, authorization and audit. It does not import the Agent implementation or open module databases.
