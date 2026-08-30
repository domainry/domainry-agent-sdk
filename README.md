# domainry-agent-sdk

Deployment-neutral contract for Domainry Agent execution. Runtime owns task/workflow state, authorization, approvals, tool policy, leases and terminal commits; implementations own provider execution and normalized provider evidence.

The SDK contains no Runtime or implementation dependency. Project composition selects a Module or SaaS Factory explicitly.
