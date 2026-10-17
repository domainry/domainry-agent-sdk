import type { AgentRun, AgentRunEvent, AgentRunPlan, AgentRunPlanStep, AgentRunStep, AgentRunToolCall, AgentRunToolProgress } from "./types.js"

function cloneCall(call: AgentRunToolCall): AgentRunToolCall {
  return {
    ...call,
    progress: call.progress?.map((item) => ({ ...item })),
    subcalls: call.subcalls?.map(cloneCall),
  }
}

export function flattenAgentRunCalls(run: AgentRun): AgentRunToolCall[] {
  const calls: AgentRunToolCall[] = []
  const append = (items: AgentRunToolCall[]) => {
    for (const call of items) {
      calls.push(call)
      append(call.subcalls || [])
    }
  }
  for (const step of run.steps || []) append(step.calls || [])
  return calls
}

function projectPlan(run: AgentRun, events: AgentRunEvent[]): AgentRunPlan | undefined {
  let plan = run.plan ? { ...run.plan, steps: run.plan.steps.map((step) => ({ ...step })) } : undefined
  for (const event of events) {
    const data = event.data || {}
    if (event.type === "run.plan.created" && Array.isArray(data.stages)) {
      // A durable snapshot is authoritative. Replaying an older creation event
      // after reconnect must never reset completed progress to planned.
      if (plan) continue
      const stages = data.stages as Array<Record<string, unknown>>
      plan = {
        key: typeof data.key === "string" ? data.key : "run-plan",
        label: typeof data.label === "string" ? data.label : "Execution plan",
        created_at: typeof data.updated_at === "string" ? data.updated_at : event.created_at,
        updated_at: typeof data.updated_at === "string" ? data.updated_at : event.created_at,
        steps: stages.map((stage, index): AgentRunPlanStep => ({
          key: typeof stage.key === "string" ? stage.key : `step-${index + 1}`,
          label: typeof stage.label === "string" ? stage.label : `Step ${index + 1}`,
          detail: typeof stage.detail === "string" ? stage.detail : "",
          optional: stage.optional === true,
          call_id: typeof stage.call_id === "string" ? stage.call_id : undefined,
          index: index + 1,
          total: stages.length,
          status: "planned",
          updated_at: typeof data.updated_at === "string" ? data.updated_at : event.created_at,
        })),
      }
      continue
    }
    if (event.type !== "run.progress" || !plan) continue
    const key = typeof data.progress_key === "string" ? data.progress_key : ""
    const status = typeof data.status === "string" ? data.status : ""
    const step = plan.steps.find((item) => item.key === key)
    if (!step || !status) continue
    const currentRank = runPlanStatusRank(step.status)
    const nextRank = runPlanStatusRank(status)
    if (nextRank < currentRank || currentRank === 2 && status !== step.status) continue
    step.status = status
    if (typeof data.detail === "string" && data.detail) step.detail = data.detail
    if (!step.call_id && typeof data.call_id === "string" && data.call_id) step.call_id = data.call_id
    if (typeof data.updated_at === "string") step.updated_at = data.updated_at
    plan.updated_at = step.updated_at
  }
  return plan
}

function runPlanStatusRank(status: string) {
  if (["completed", "failed", "skipped"].includes(status)) return 2
  if (["started", "waiting_user", "waiting_confirmation"].includes(status)) return 1
  return 0
}

function statusRank(status: string) {
  if (["completed", "failed", "uncertain", "needs_reconciliation", "waiting_confirmation", "waiting_user"].includes(status)) return 2
  if (status === "running") return 1
  return 0
}

function progressStatusRank(status: string) {
  if (["completed", "failed", "skipped"].includes(status)) return 2
  if (status === "started") return 1
  return 0
}

function mergeProgress(current: AgentRunToolProgress | undefined, next: AgentRunToolProgress) {
  if (current && progressStatusRank(current.status) === 2 && next.status !== current.status) return current
  if (current && progressStatusRank(next.status) < progressStatusRank(current.status)) return current
  return { ...current, ...next }
}

function mergeToolProgress(call: AgentRunToolCall, event: AgentRunEvent) {
  const data = event.data || {}
  if (event.type === "tool.plan.created" && Array.isArray(data.stages)) {
    const stages = data.stages as Array<Record<string, unknown>>
    const current = new Map((call.progress || []).map((item) => [item.key, item]))
    const declared = stages.flatMap((stage, index): AgentRunToolProgress[] => {
      const key = typeof stage.key === "string" ? stage.key : ""
      const label = typeof stage.label === "string" ? stage.label : ""
      if (!key || !label) return []
      const next: AgentRunToolProgress = {
        key,
        label,
        detail: typeof stage.detail === "string" ? stage.detail : undefined,
        ...(stage.optional === true ? { optional: true } : {}),
        index: index + 1,
        total: stages.length,
        status: "planned",
        updated_at: typeof data.updated_at === "string" ? data.updated_at : event.created_at,
      }
      return [mergeProgress(current.get(key), next)]
    })
    const declaredKeys = new Set(declared.map((item) => item.key))
    call.progress = [...declared, ...(call.progress || []).filter((item) => !declaredKeys.has(item.key))]
    return
  }
  if (event.type !== "tool.progress") return
  const key = typeof data.progress_key === "string" ? data.progress_key : ""
  const status = typeof data.status === "string" ? data.status : ""
  const progress = call.progress?.find((item) => item.key === key)
  if (!progress || !status) return
  const next = mergeProgress(progress, {
    ...progress,
    status,
    detail: typeof data.detail === "string" ? data.detail : progress.detail,
    updated_at: typeof data.updated_at === "string" ? data.updated_at : progress.updated_at,
  })
  Object.assign(progress, next)
}

function projectCalls(run: AgentRun, events: AgentRunEvent[]) {
  const source = { ...run, steps: run.steps?.map((step) => ({ ...step, calls: step.calls.map(cloneCall) })) }
  const calls = new Map(flattenAgentRunCalls(source).map((call) => [call.id, call]))
  const merge = (id: string, name: string, status: string, patch: Partial<AgentRunToolCall> = {}) => {
    if (!id) return
    const current = calls.get(id)
    if (!current) {
      calls.set(id, { id, name: name || "unknown_tool", status, ...patch })
      return
    }
    Object.assign(current, patch)
    if (name) current.name = name
    if (statusRank(status) >= statusRank(current.status)) current.status = status
  }
  for (const event of events) {
    const data = event.data || {}
    if (event.type === "step.completed" && Array.isArray(data.calls)) {
      for (const raw of data.calls as Array<Record<string, unknown>>) {
        merge(typeof raw.id === "string" ? raw.id : "", typeof raw.name === "string" ? raw.name : "", "queued", {
          arguments: typeof raw.arguments === "string" ? raw.arguments : undefined,
        })
      }
    }
    const callID = typeof data.call_id === "string" ? data.call_id : ""
    const tool = typeof data.tool === "string" ? data.tool : ""
    const patch: Partial<AgentRunToolCall> = {}
    if (typeof data.arguments === "string") patch.arguments = data.arguments
    if (typeof data.result_preview === "string") patch.result_preview = data.result_preview
    if (typeof data.result_truncated === "boolean") patch.result_truncated = data.result_truncated
    if (typeof data.error_code === "string") patch.error_code = data.error_code
    if (event.type === "tool.plan.created") merge(callID, tool, "queued")
    if (event.type === "tool.queued") merge(callID, tool, "queued", patch)
    if (event.type === "tool.started") merge(callID, tool, typeof data.status === "string" ? data.status : "running", patch)
    if (event.type === "tool.completed") merge(callID, tool, typeof data.status === "string" ? data.status : "completed", patch)
    if (event.type === "tool.uncertain") merge(callID, tool, "uncertain", patch)
    if (event.type === "tool.waiting") merge(callID, tool, typeof data.status === "string" ? data.status : "waiting_confirmation", patch)
    const call = calls.get(callID)
    if (call) mergeToolProgress(call, event)
  }
  return calls
}

function flattenCalls(calls: AgentRunToolCall[]) {
  const values: AgentRunToolCall[] = []
  const append = (call: AgentRunToolCall) => {
    values.push(call)
    for (const child of call.subcalls || []) append(child)
  }
  for (const call of calls) append(call)
  return values
}

function projectSteps(run: AgentRun, events: AgentRunEvent[], calls: Map<string, AgentRunToolCall>): AgentRunStep[] {
  const steps = new Map<number, AgentRunStep>()
  for (const step of run.steps || []) {
    steps.set(step.number, { ...step, calls: step.calls.map((call) => calls.get(call.id) || cloneCall(call)) })
  }
  const ensureStep = (number: number) => {
    let step = steps.get(number)
    if (!step) {
      step = { number, status: "generating", calls: [] }
      steps.set(number, step)
    }
    return step
  }
  for (const event of events) {
    const data = event.data || {}
    const number = typeof data.step === "number" ? data.step : -1
    if (number < 0) continue
    if (event.type === "step.started" || event.type === "model.attempt.started") {
      ensureStep(number).status = "generating"
      continue
    }
    if (event.type === "step.completed") {
      const step = ensureStep(number)
      step.status = data.finish_reason === "tool_calls" ? "tools" : "completed"
      if (typeof data.text === "string") step.text = data.text
      if (Array.isArray(data.calls)) {
        step.calls = (data.calls as Array<Record<string, unknown>>).flatMap((raw): AgentRunToolCall[] => {
          const id = typeof raw.id === "string" ? raw.id : ""
          if (!id) return []
          return [calls.get(id) || { id, name: typeof raw.name === "string" ? raw.name : "unknown_tool", arguments: typeof raw.arguments === "string" ? raw.arguments : undefined, status: "queued" }]
        })
      }
      continue
    }
    const callID = typeof data.call_id === "string" ? data.call_id : ""
    if (!callID || !event.type.startsWith("tool.")) continue
    const step = ensureStep(number)
    const call = calls.get(callID)
    if (call && !step.calls.some((item) => item.id === callID)) step.calls.push(call)
  }
  const ordered = [...steps.values()].sort((left, right) => left.number - right.number).map((step) => ({
    ...step,
    calls: step.calls.map((call) => calls.get(call.id) || call),
  }))
  const assigned = new Set(ordered.flatMap((step) => flattenCalls(step.calls).map((call) => call.id)))
  const unassigned = [...calls.values()].filter((call) => !assigned.has(call.id))
  if (ordered.length > 0 && unassigned.length > 0) ordered[ordered.length - 1].calls.push(...unassigned)
  return ordered
}

// projectAgentRun folds a bounded event window over a durable snapshot. It is
// presentation-agnostic: products decide labels and layout from the projection.
export function projectAgentRun(run: AgentRun, events: AgentRunEvent[]): AgentRun {
  const calls = projectCalls(run, events)
  return { ...run, plan: projectPlan(run, events), steps: projectSteps(run, events, calls) }
}
