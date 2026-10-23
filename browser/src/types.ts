export type AgentConversation = {
  id: string
  title: string
  title_origin: "unset" | "generated" | "user"
  title_status: "pending" | "ready" | "failed"
  archived: boolean
  memory_enabled: boolean
  revision: number
  active_run_id?: string
  binding?: AgentConversationBinding
  created_at: string
  updated_at: string
}

export type AgentConversationBinding = { kind: string; key: string }

export type AgentConversationListOptions = {
  beforeId?: string
  search?: string
  includeArchived?: boolean
  limit?: number
  binding?: AgentConversationBinding
}

export type AgentConversationCreate = {
  clientId: string
  title?: string
  agentId?: string
  memoryEnabled?: boolean
  binding?: AgentConversationBinding
}

export type AgentConversationUpdate = {
  expectedRevision: number
  title?: string
  archived?: boolean
  memoryEnabled?: boolean
}

export type AgentCitation = {
  id: string
  title?: string
  excerpt?: string
  provider?: string
  source?: string
  object_key?: string
  record_id?: string
}

export type AgentImageReference = {
  attachment_id: string
  conversation_id?: string
  filename?: string
  content_type?: string
  bytes?: number
  sha256?: string
  revision?: number
  detail?: "auto" | "low" | "high"
}

export type AgentContentBlock =
  | { type: "text"; text: string }
  | { type: "image"; image: AgentImageReference }

export type AgentReference = {
  kind: string
  payload: Record<string, unknown>
}

export type AgentMessage = {
  generation?: { step: number; run_attempt: number; model_attempt: number; status: "generating" | "completed" | "interrupted" }
  id: string
  run_id?: string
  seq: number
  role: "user" | "assistant" | string
  content: string
  content_blocks?: AgentContentBlock[]
  references?: AgentReference[]
  citations?: AgentCitation[]
  access_error?: string
  created_at: string
}

export type AgentAttachment = {
  id: string
  conversation_id: string
  filename: string
  content_type: string
  bytes: number
  sha256: string
  visibility: "conversation_private" | string
  state: "uploading" | "stored" | "indexing" | "ready" | "failed" | "needs_reconcile" | "deleting" | "deleted" | string
  index_status?: string
  revision: number
  error_code?: string
  indexing?: {
    last_check_revision?: number
    requested: boolean
    can_start: boolean
    can_check: boolean
    max_bytes?: number
    reason?: string
  }
  created_at: string
  updated_at: string
}

export type AgentRunPlanStep = {
  key: string
  label: string
  detail?: string
  optional?: boolean
  binding?: { kind: "model" | "tool" | "tool_progress"; tools?: string[]; progress_key?: string }
  call_id?: string
  index: number
  total: number
  status: "planned" | "started" | "waiting_user" | "waiting_confirmation" | "completed" | "skipped" | "failed" | string
  updated_at?: string
}

export type AgentRunPlan = {
  key: string
  label: string
  steps: AgentRunPlanStep[]
  created_at?: string
  updated_at?: string
}

export type AgentRunToolProgress = {
  key: string
  label: string
  detail?: string
  optional?: boolean
  index: number
  total: number
  status: "planned" | "started" | "completed" | "skipped" | "failed" | string
  updated_at?: string
}

export type AgentRunToolCall = {
  id: string
  name: string
  status: string
  effect?: "read" | "write" | string
  arguments?: string
  result_preview?: string
  result_truncated?: boolean
  error_code?: string
  progress?: AgentRunToolProgress[]
  subcalls?: AgentRunToolCall[]
}

export type AgentRunStep = {
  message_id?: string
  message_seq?: number
  attempt?: number
  model_attempts?: number
  number: number
  status: string
  calls: AgentRunToolCall[]
  text?: string
  started_at?: string
  completed_at?: string
  duration_ms?: number
  last_model_error?: string
}

export type AgentRunMetrics = {
  steps: number
  model_calls: number
  model_retries: number
  tool_calls: number
  tool_attempts: number
  parallel_tool_batches: number
  parallel_tool_calls: number
  peak_parallel_tools: number
  authorization_checks: number
  confirmation_decisions: number
  context_compactions: number
  compacted_results: number
  compacted_intervals: number
  peak_context_bytes: number
  context_limit_bytes: number
  cache_read_input_tokens: number
  cache_creation_input_tokens: number
}

export type AgentRunAuditEvent = {
  seq: number
  type: "run" | "context" | "step" | "model" | "tool" | "authorization" | "confirmation" | "interaction" | string
  status: string
  step: number
  attempt?: number
  model_attempt?: number
  retry_delay_ms?: number
  call_id?: string
  tool?: string
  action_key?: string
  authorization_revision?: string
  interaction_id?: string
  actor_id?: string
  error_code?: string
  duration_ms?: number
  occurred_at: string
}

export type AgentInteractionOperation = {
  call_id: string
  tool: string
  tool_version: string
  action_key: string
  arguments: string
}

export type AgentInteraction = AgentInteractionOperation & {
  id: string
  kind: "input" | "confirmation" | "reconciliation"
  status: string
  question: string
  choices?: string[]
  revision: number
  expires_at: string
  operations?: AgentInteractionOperation[]
}

export type AgentInteractionResponse = {
  interaction_id: string
  client_id: string
  expected_revision: number
  decision: "answer" | "approve" | "reject"
  answer?: string
  scope?: "listed_operations"
}

export type AgentRunModelAttempt = {
  step: number
  run_attempt: number
  number: number
  status: "started" | "retry_scheduled" | "failed" | "completed" | string
  error_code?: string
  retry_at?: string
  retry_delay_ms?: number
  started_at: string
  completed_at?: string
}

export type AgentRun = {
  id: string
  conversation_id?: string
  status: "queued" | "running" | "completed" | "failed" | "cancelled" | "waiting_user" | "waiting_confirmation" | "needs_reconciliation" | string
  error_code?: string
  access_error?: string
  interaction?: AgentInteraction
  model?: string
  metrics?: AgentRunMetrics
  audit?: AgentRunAuditEvent[]
  audit_complete?: boolean
  plan?: AgentRunPlan
  steps?: AgentRunStep[]
  started_at?: string
  completed_at?: string
  duration_ms?: number
  queue_duration_ms?: number
  attempt?: number
  draft_text?: string
  draft_bytes?: number
  draft_message_id?: string
  draft_message_seq?: number
  last_event_seq?: number
  model_attempts?: AgentRunModelAttempt[]
}

export type AgentRunEvent = {
  run_id: string
  seq: number
  type: string
  data?: Record<string, unknown>
  created_at: string
}

export type AgentDirectory = {
  items: Array<{ id: string; name: string; enabled: boolean }>
  models: string[]
}
