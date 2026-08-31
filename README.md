# domainry-agent-sdk

Deployment-neutral contract for Domainry Agent execution. Runtime owns task/workflow state, authorization, approvals, tool policy, leases and terminal commits; implementations own provider execution and normalized provider evidence.

The SDK contains no Runtime or implementation dependency. Project composition selects a Module or SaaS Factory explicitly.

## Package layout

- The root package is the stable `Factory`, `Binding`, runner, and schema entrypoint.
- `persistence` owns Agent persistence capabilities and transaction-aware mutation contracts.
- `modulehost` describes infrastructure borrowed by an embedded Agent module.
- `saashost` describes the SaaS composition boundary.
- `contracttest` contains deployment-parity tests; `state` contains shared Agent state values.

Concrete SQL stores remain in the Agent implementation repository; the SDK exposes only their deployment-neutral contracts through `persistence`.

Run `go test ./...` before publishing an immutable SDK version.
