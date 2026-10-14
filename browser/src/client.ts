import type {
  AgentAttachment,
  AgentContentBlock,
  AgentConversation,
  AgentConversationBinding,
  AgentDirectory,
  AgentInteractionResponse,
  AgentMessage,
  AgentReference,
  AgentRun,
  AgentRunEvent,
} from "./types.js"

export type AgentStreamFrame = { id: string; event: string; data: string }

export type AgentClientTransport = {
  request<T>(path: string, init?: RequestInit): Promise<T>
  binary(path: string): Promise<Blob>
  stream(path: string, options: {
    signal?: AbortSignal
    lastEventId?: string
    onOpen?: () => void
    onEvent: (event: AgentStreamFrame) => void | Promise<void>
  }): Promise<void>
}

export function createAgentClient(transport: AgentClientTransport) {
  return {
    directory: () => transport.request<AgentDirectory>("/agent/agents"),
    listConversations(beforeId?: string, binding?: AgentConversationBinding) {
      const query = new URLSearchParams({ limit: "50" })
      if (beforeId) query.set("before_id", beforeId)
      if (binding) {
        query.set("binding_kind", binding.kind)
        query.set("binding_key", binding.key)
      }
      return transport.request<{ items: AgentConversation[]; next_cursor?: string }>(`/agent/conversations?${query.toString()}`)
    },
    createConversation(clientId: string, title = "New conversation", binding?: AgentConversationBinding) {
      return transport.request<AgentConversation>("/agent/conversations", { method: "POST", body: JSON.stringify({ client_id: clientId, title, ...(binding ? { binding } : {}) }) })
    },
    getConversation: (conversationId: string) => transport.request<AgentConversation>(`/agent/conversations/${encodeURIComponent(conversationId)}`),
    updateConversationTitle(conversationId: string, title: string, revision: number) {
      return transport.request<AgentConversation>(`/agent/conversations/${encodeURIComponent(conversationId)}`, {
        method: "PATCH", body: JSON.stringify({ title, expected_revision: revision }),
      })
    },
    listMessages(conversationId: string, beforeSeq?: number) {
      const query = new URLSearchParams({ limit: "100" })
      if (beforeSeq) query.set("before_seq", String(beforeSeq))
      return transport.request<{ items: AgentMessage[]; next_before_seq?: number }>(`/agent/conversations/${encodeURIComponent(conversationId)}/messages?${query.toString()}`)
    },
    sendMessage(conversationId: string, clientMessageId: string, message: string, content?: AgentContentBlock[], references?: AgentReference[]) {
      return transport.request<AgentRun>(`/agent/conversations/${encodeURIComponent(conversationId)}/messages`, {
        method: "POST", body: JSON.stringify({ client_message_id: clientMessageId, message, ...(content?.length ? { content } : {}), ...(references?.length ? { references } : {}) }),
      })
    },
    listAttachments(conversationId: string, after?: string) {
      const query = new URLSearchParams({ limit: "50" })
      if (after) query.set("after", after)
      return transport.request<{ items: AgentAttachment[]; next_after?: string; complete: boolean }>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments?${query.toString()}`)
    },
    uploadAttachment(conversationId: string, clientId: string, file: File) {
      const query = new URLSearchParams({ client_id: clientId, filename: file.name })
      return transport.request<AgentAttachment>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments?${query.toString()}`, {
        method: "POST", headers: { "Content-Type": "application/octet-stream" }, body: file,
      })
    },
    getAttachment: (conversationId: string, attachmentId: string) => transport.request<AgentAttachment>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}`),
    downloadAttachment: (conversationId: string, attachmentId: string) => transport.binary(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}/content`),
    indexAttachment(conversationId: string, attachmentId: string, expectedRevision: number) {
      const query = new URLSearchParams({ expected_revision: String(expectedRevision) })
      return transport.request<AgentAttachment>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}/index?${query.toString()}`, { method: "POST" })
    },
    checkAttachmentIndex(conversationId: string, attachmentId: string, expectedRevision: number) {
      const query = new URLSearchParams({ expected_revision: String(expectedRevision) })
      return transport.request<AgentAttachment>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}/index/check?${query.toString()}`, { method: "POST" })
    },
    deleteAttachment(conversationId: string, attachmentId: string, expectedRevision: number) {
      const query = new URLSearchParams({ expected_revision: String(expectedRevision) })
      return transport.request<AgentAttachment>(`/agent/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}?${query.toString()}`, { method: "DELETE" })
    },
    getRun: (conversationId: string, runId: string) => transport.request<AgentRun>(`/agent/conversations/${encodeURIComponent(conversationId)}/runs/${encodeURIComponent(runId)}`),
    streamRun(conversationId: string, runId: string, afterSeq: number, onEvent: (event: AgentRunEvent) => void | Promise<void>, options: { signal?: AbortSignal; onOpen?: () => void } = {}) {
      return transport.stream(`/agent/conversations/${encodeURIComponent(conversationId)}/runs/${encodeURIComponent(runId)}/events/stream`, {
        signal: options.signal,
        lastEventId: afterSeq > 0 ? String(afterSeq) : undefined,
        onOpen: options.onOpen,
        onEvent: async (frame) => {
          if (frame.event === "stream.error") throw new Error("Agent event stream interrupted")
          const event = JSON.parse(frame.data) as AgentRunEvent
          if (!Number.isSafeInteger(event.seq) || event.seq < 1 || event.run_id !== runId) throw new Error("Invalid Agent event stream payload")
          await onEvent(event)
        },
      })
    },
    respond(conversationId: string, runId: string, response: AgentInteractionResponse) {
      return transport.request<AgentRun>(`/agent/conversations/${encodeURIComponent(conversationId)}/runs/${encodeURIComponent(runId)}/respond`, { method: "POST", body: JSON.stringify(response) })
    },
    cancel: (conversationId: string, runId: string) => transport.request<AgentRun>(`/agent/conversations/${encodeURIComponent(conversationId)}/runs/${encodeURIComponent(runId)}/cancel`, { method: "POST" }),
    resume: (conversationId: string, runId: string) => transport.request<AgentRun>(`/agent/conversations/${encodeURIComponent(conversationId)}/runs/${encodeURIComponent(runId)}/resume`, { method: "POST" }),
  }
}

export type AgentClient = ReturnType<typeof createAgentClient>
