import assert from "node:assert/strict"
import test from "node:test"

import { createAgentClient } from "../dist/index.js"

test("client carries open conversation bindings and message references", async () => {
  const requests = []
  const client = createAgentClient({
    async request(path, init) {
      requests.push({ path, init })
      return { items: [] }
    },
    async binary() { return new Blob() },
    async stream() {},
  })

  await client.listConversations(undefined, { kind: "product.record", key: "account:one" })
  await client.createConversation("create-1", "Account", { kind: "product.record", key: "account:one" })
  await client.sendMessage("conversation-1", "message-1", "Review it", undefined, [{ kind: "product.record", payload: { record_id: "one" } }])

  assert.match(requests[0].path, /binding_kind=product.record/)
  assert.match(requests[0].path, /binding_key=account%3Aone/)
  assert.deepEqual(JSON.parse(requests[1].init.body), { client_id: "create-1", title: "Account", binding: { kind: "product.record", key: "account:one" } })
  assert.deepEqual(JSON.parse(requests[2].init.body).references, [{ kind: "product.record", payload: { record_id: "one" } }])
})
