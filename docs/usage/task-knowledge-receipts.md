# Task Knowledge reference receipts

A successful trusted Task Knowledge host can return `modulehost.TaskToolResult.Knowledge` with an `AgentTaskKnowledgeEvidence`. The Agent tool ledger retains that reference-only receipt with the same fenced invocation, input/output hashes and terminal state. The field is independent of model-authored Task output.

The receipt contains a library, operation, scope hash and at most twenty normalized citation references. It includes document/page versions, record/chunk identifiers and content/source revision hashes when the Knowledge owner supplied them. It excludes excerpts, titles, URLs, raw provider payloads, credentials and permission bundles. An empty citation collection represents an authorized no-match, not an absent receipt.

`Validate` enforces bounded identifiers, supported operations, hashes, unique citations and nonnegative versions. `Clone` isolates the citation slice. Hosts and consumers must additionally bind the receipt to the actual tool arguments and execution status. A saved receipt never authorizes a later read; current Knowledge library/document and audience policy still apply.

Task and interactive public read models do not automatically expose the internal ledger or grant source access. The receipt is for trusted completion/evidence boundaries.
