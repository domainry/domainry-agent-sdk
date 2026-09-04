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

Run `go test ./...` before publishing an immutable SDK version.
