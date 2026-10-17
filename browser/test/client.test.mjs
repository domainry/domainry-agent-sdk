import assert from "node:assert/strict"
import test from "node:test"

import { createAgentClient } from "../dist/index.js"

test("client carries conversation management filters, mutations, bindings and message references", async () => {
  const requests = []
  const client = createAgentClient({
    async request(path, init) {
      requests.push({ path, init })
      return { items: [] }
    },
    async binary() { return new Blob() },
    async stream() {},
  })

  await client.listConversations({ search: "renewal", includeArchived: true, beforeId: "cursor-1", limit: 25, binding: { kind: "product.record", key: "account:one" } })
  await client.createConversation({ clientId: "create-1", title: "Account", memoryEnabled: true, binding: { kind: "product.record", key: "account:one" } })
  await client.updateConversation("conversation-1", { expectedRevision: 3, archived: true })
  await client.deleteConversation("conversation-1", 4)
  await client.sendMessage("conversation-1", "message-1", "Review it", undefined, [{ kind: "product.record", payload: { record_id: "one" } }])

  assert.match(requests[0].path, /search=renewal/)
  assert.match(requests[0].path, /include_archived=true/)
  assert.match(requests[0].path, /before_id=cursor-1/)
  assert.match(requests[0].path, /limit=25/)
  assert.match(requests[0].path, /binding_kind=product.record/)
  assert.match(requests[0].path, /binding_key=account%3Aone/)
  assert.deepEqual(JSON.parse(requests[1].init.body), { client_id: "create-1", title: "Account", memory_enabled: true, binding: { kind: "product.record", key: "account:one" } })
  assert.deepEqual(JSON.parse(requests[2].init.body), { expected_revision: 3, archived: true })
  assert.equal(requests[3].init.method, "DELETE")
  assert.match(requests[3].path, /expected_revision=4/)
  assert.deepEqual(JSON.parse(requests[4].init.body).references, [{ kind: "product.record", payload: { record_id: "one" } }])
})
