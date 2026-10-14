import assert from "node:assert/strict"
import test from "node:test"

import { flattenAgentRunCalls, projectAgentRun } from "../dist/index.js"

test("older progress events cannot roll a durable terminal snapshot backwards", () => {
  const run = {
    id: "run-1",
    status: "running",
    steps: [{
      number: 1,
      status: "tools",
      calls: [{
        id: "call-1",
        name: "search",
        status: "completed",
        progress: [{ key: "search", label: "Search", detail: "Search complete", index: 1, total: 1, status: "completed" }],
      }],
    }],
  }
  const events = [{
    run_id: "run-1",
    seq: 1,
    type: "tool.plan.created",
    created_at: "2026-01-01T00:00:00Z",
    data: { call_id: "call-1", tool: "search", stages: [{ key: "search", label: "Search", detail: "Searching" }] },
  }, {
    run_id: "run-1",
    seq: 2,
    type: "tool.progress",
    created_at: "2026-01-01T00:00:01Z",
    data: { call_id: "call-1", progress_key: "search", status: "started", detail: "Searching" },
  }]

  const [call] = flattenAgentRunCalls(projectAgentRun(run, events))

  assert.equal(call.status, "completed")
  assert.deepEqual(call.progress, [{ key: "search", label: "Search", detail: "Search complete", index: 1, total: 1, status: "completed" }])
})

test("events project plans, steps and calls without product-specific knowledge", () => {
  const run = { id: "run-2", status: "running", steps: [] }
  const events = [{
    run_id: "run-2",
    seq: 1,
    type: "step.completed",
    created_at: "2026-01-01T00:00:00Z",
    data: { step: 0, finish_reason: "tool_calls", calls: [{ id: "call-2", name: "custom.lookup", arguments: "{}" }] },
  }, {
    run_id: "run-2",
    seq: 2,
    type: "tool.completed",
    created_at: "2026-01-01T00:00:01Z",
    data: { step: 0, call_id: "call-2", tool: "custom.lookup", status: "completed", result_preview: "{}" },
  }]

  const projected = projectAgentRun(run, events)

  assert.equal(projected.steps[0].status, "tools")
  assert.equal(projected.steps[0].calls[0].name, "custom.lookup")
  assert.equal(projected.steps[0].calls[0].status, "completed")
})
