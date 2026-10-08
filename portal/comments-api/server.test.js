const http = require("http");
const assert = require("node:assert/strict");
const test = require("node:test");
const { createHandler, mapComment, rewritePostgresUrl } = require("./server");

function listen(handler) {
  const server = http.createServer((req, res) => {
    void handler(req, res);
  });
  return new Promise((resolve) => {
    server.listen(0, "127.0.0.1", () => {
      resolve({ server, port: server.address().port });
    });
  });
}

function request(port, method, path, body) {
  return new Promise((resolve, reject) => {
    const req = http.request(
      { hostname: "127.0.0.1", port, path, method, headers: { "Content-Type": "application/json" } },
      (res) => {
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => {
          const text = Buffer.concat(chunks).toString("utf8");
          resolve({ status: res.statusCode, body: text ? JSON.parse(text) : null });
        });
      }
    );
    req.on("error", reject);
    if (body !== undefined) req.write(JSON.stringify(body));
    req.end();
  });
}

test("rewrites only the sslmode query parameter", () => {
  assert.equal(
    rewritePostgresUrl("postgres://u:p@db:5432/app?sslmode=require"),
    "postgres://u:p@db:5432/app?sslmode=no-verify"
  );
  assert.equal(
    rewritePostgresUrl("postgres://u:sslmode=require@db:5432/app"),
    "postgres://u:sslmode=require@db:5432/app"
  );
});

test("maps a database row to the panel comment shape", () => {
  const comment = mapComment({
    id: "1",
    story_id: "story",
    author: "Ada",
    body: "hello",
    resolved: false,
    created_at: new Date("2026-10-08T00:00:00.000Z"),
  });
  assert.deepEqual(comment, {
    id: "1",
    storyId: "story",
    author: "Ada",
    body: "hello",
    createdAt: "2026-10-08T00:00:00.000Z",
    resolved: false,
  });
});

test("health responds before the schema is ready", async () => {
  const handler = createHandler({
    query: async () => {
      throw new Error("should not query");
    },
    state: { postgres: false, schemaReady: false },
  });
  const { server, port } = await listen(handler);
  try {
    const res = await request(port, "GET", "/healthz");
    assert.equal(res.status, 200);
    assert.deepEqual(res.body, { ok: true, postgres: false, schemaReady: false });
    const blocked = await request(port, "GET", "/comments?storyId=a");
    assert.equal(blocked.status, 503);
    assert.deepEqual(blocked.body, { error: "schema not ready" });
  } finally {
    server.close();
  }
});

test("lists, creates, resolves, and deletes a comment", async () => {
  const rows = [];
  const query = async (text, params) => {
    if (text.includes("WHERE story_id = $1")) {
      return { rows: rows.filter((row) => row.story_id === params[0]) };
    }
    if (text.includes("GROUP BY story_id")) {
      const counts = new Map();
      for (const row of rows) {
        if (row.resolved) continue;
        counts.set(row.story_id, (counts.get(row.story_id) || 0) + 1);
      }
      return {
        rows: [...counts.entries()].map(([story_id, n]) => ({ story_id, n })),
      };
    }
    if (text.startsWith("INSERT")) {
      const row = {
        id: params[0],
        story_id: params[1],
        author: params[2],
        body: params[3],
        resolved: false,
        created_at: new Date("2026-10-08T00:00:00.000Z"),
      };
      rows.push(row);
      return { rows: [row] };
    }
    if (text.startsWith("UPDATE")) {
      const row = rows.find((item) => item.id === params[0]);
      if (!row) return { rows: [] };
      row.resolved = params[1];
      return { rows: [row] };
    }
    if (text.startsWith("DELETE")) {
      const index = rows.findIndex((item) => item.id === params[0]);
      if (index === -1) return { rowCount: 0 };
      rows.splice(index, 1);
      return { rowCount: 1 };
    }
    throw new Error(`unexpected query: ${text}`);
  };
  const handler = createHandler({
    query,
    state: { postgres: true, schemaReady: true },
  });
  const { server, port } = await listen(handler);
  try {
    const missing = await request(port, "GET", "/comments");
    assert.equal(missing.status, 400);

    const created = await request(port, "POST", "/comments", {
      storyId: "components-checkboxwithtooltip--default",
      author: " Ada ",
      body: " hello ",
    });
    assert.equal(created.status, 201);
    assert.equal(created.body.comment.author, "Ada");
    assert.equal(created.body.comment.body, "hello");
    assert.equal(created.body.comment.resolved, false);

    const listed = await request(
      port,
      "GET",
      "/comments?storyId=components-checkboxwithtooltip--default"
    );
    assert.equal(listed.status, 200);
    assert.equal(listed.body.comments.length, 1);

    const counts = await request(port, "GET", "/comments/counts");
    assert.deepEqual(counts.body.counts, {
      "components-checkboxwithtooltip--default": 1,
    });

    const resolved = await request(port, "POST", `/comments/${created.body.comment.id}/resolve`, {
      resolved: true,
    });
    assert.equal(resolved.status, 200);
    assert.equal(resolved.body.comment.resolved, true);

    const missingRow = await request(port, "POST", "/comments/missing/delete");
    assert.equal(missingRow.status, 404);
    assert.deepEqual(missingRow.body, { error: "comment not found" });

    const deleted = await request(port, "POST", `/comments/${created.body.comment.id}/delete`);
    assert.equal(deleted.status, 200);
    assert.deepEqual(deleted.body, { ok: true });

    const unknown = await request(port, "GET", "/nope");
    assert.equal(unknown.status, 404);
    assert.deepEqual(unknown.body, { error: "not found" });
  } finally {
    server.close();
  }
});
