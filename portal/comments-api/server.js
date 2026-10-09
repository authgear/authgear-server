const http = require("http");
const { randomUUID } = require("crypto");
const { Pool } = require("pg");

const PORT = Number(process.env.PORT || 8080);
const MAX_BODY_BYTES = 32 * 1024;
const MAX_STORY_ID = 500;
const MAX_AUTHOR = 80;
const MAX_BODY = 4000;

const CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS story_comments (
  id TEXT PRIMARY KEY,
  story_id TEXT NOT NULL,
  author TEXT NOT NULL,
  body TEXT NOT NULL,
  resolved BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`;

const CREATE_INDEX = `
CREATE INDEX IF NOT EXISTS story_comments_story_id_idx
ON story_comments (story_id)`;

function rewritePostgresUrl(url) {
  if (!url) return url;
  return url.replace(/([?&])sslmode=require\b/g, "$1sslmode=no-verify");
}

function mapComment(row) {
  const createdAt = row.created_at instanceof Date ? row.created_at : new Date(row.created_at);
  return {
    id: row.id,
    storyId: row.story_id,
    author: row.author,
    body: row.body,
    createdAt: createdAt.toISOString(),
    resolved: Boolean(row.resolved),
  };
}

function send(res, status, body) {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    let size = 0;
    req.on("data", (chunk) => {
      size += chunk.length;
      if (size > MAX_BODY_BYTES) {
        reject(Object.assign(new Error("body too large"), { status: 413 }));
        req.destroy();
        return;
      }
      chunks.push(chunk);
    });
    req.on("end", () => {
      if (size === 0) {
        resolve({});
        return;
      }
      try {
        const parsed = JSON.parse(Buffer.concat(chunks).toString("utf8"));
        if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
          reject(Object.assign(new Error("invalid json"), { status: 400 }));
          return;
        }
        resolve(parsed);
      } catch (err) {
        if (err.status) {
          reject(err);
          return;
        }
        reject(Object.assign(new Error("invalid json"), { status: 400 }));
      }
    });
    req.on("error", reject);
  });
}

function decodeId(segment) {
  try {
    return decodeURIComponent(segment);
  } catch (err) {
    if (err instanceof URIError) return null;
    throw err;
  }
}

function requireText(value, max, field) {
  if (typeof value !== "string") {
    return `${field} is required`;
  }
  const trimmed = value.trim();
  if (!trimmed) return `${field} is required`;
  if (trimmed.length > max) return `${field} is too long`;
  return null;
}

function createHandler({ query, state }) {
  return async function handle(req, res) {
    const url = new URL(req.url || "/", "http://localhost");
    const path = url.pathname;

    if ((path === "/" || path === "/healthz") && req.method === "GET") {
      send(res, 200, {
        ok: true,
        postgres: state.postgres,
        schemaReady: state.schemaReady,
      });
      return;
    }

    if (!state.schemaReady) {
      send(res, 503, { error: "schema not ready" });
      return;
    }

    try {
      if (path === "/comments" && req.method === "GET") {
        const storyId = url.searchParams.get("storyId");
        const storyError = requireText(storyId, MAX_STORY_ID, "storyId");
        if (storyError) {
          send(res, 400, { error: storyError });
          return;
        }
        const result = await query(
          `SELECT id, story_id, author, body, resolved, created_at
           FROM story_comments
           WHERE story_id = $1
           ORDER BY created_at ASC`,
          [storyId.trim()]
        );
        send(res, 200, { comments: result.rows.map(mapComment) });
        return;
      }

      if (path === "/comments/counts" && req.method === "GET") {
        const result = await query(
          `SELECT story_id, COUNT(*)::int AS n
           FROM story_comments
           WHERE resolved = FALSE
           GROUP BY story_id`
        );
        const counts = {};
        for (const row of result.rows) {
          counts[row.story_id] = row.n;
        }
        send(res, 200, { counts });
        return;
      }

      if (path === "/comments" && req.method === "POST") {
        const body = await readBody(req);
        const storyError = requireText(body.storyId, MAX_STORY_ID, "storyId");
        const authorError = requireText(body.author, MAX_AUTHOR, "author");
        const bodyError = requireText(body.body, MAX_BODY, "body");
        const error = storyError || authorError || bodyError;
        if (error) {
          send(res, 400, { error });
          return;
        }
        const id = randomUUID();
        const result = await query(
          `INSERT INTO story_comments (id, story_id, author, body)
           VALUES ($1, $2, $3, $4)
           RETURNING id, story_id, author, body, resolved, created_at`,
          [id, body.storyId.trim(), body.author.trim(), body.body.trim()]
        );
        send(res, 201, { comment: mapComment(result.rows[0]) });
        return;
      }

      const resolveMatch = path.match(/^\/comments\/([^/]+)\/resolve$/);
      if (resolveMatch && req.method === "POST") {
        const id = decodeId(resolveMatch[1]);
        if (!id) {
          send(res, 400, { error: "invalid id" });
          return;
        }
        const body = await readBody(req);
        if (typeof body.resolved !== "boolean") {
          send(res, 400, { error: "resolved is required" });
          return;
        }
        const result = await query(
          `UPDATE story_comments
           SET resolved = $2
           WHERE id = $1
           RETURNING id, story_id, author, body, resolved, created_at`,
          [id, body.resolved]
        );
        if (result.rows.length === 0) {
          send(res, 404, { error: "comment not found" });
          return;
        }
        send(res, 200, { comment: mapComment(result.rows[0]) });
        return;
      }

      const deleteMatch = path.match(/^\/comments\/([^/]+)\/delete$/);
      if (deleteMatch && req.method === "POST") {
        const id = decodeId(deleteMatch[1]);
        if (!id) {
          send(res, 400, { error: "invalid id" });
          return;
        }
        const result = await query(`DELETE FROM story_comments WHERE id = $1`, [id]);
        if (result.rowCount === 0) {
          send(res, 404, { error: "comment not found" });
          return;
        }
        send(res, 200, { ok: true });
        return;
      }

      send(res, 404, { error: "not found" });
    } catch (err) {
      if (err.status) {
        send(res, err.status, { error: err.message });
        return;
      }
      console.error(err);
      send(res, 500, { error: "database error" });
    }
  };
}

function watchPool(pool) {
  // An idle client error with no listener exits the process.
  pool.on("error", (err) => {
    console.error("idle postgres client error", err);
  });
}

function start() {
  const state = { postgres: false, schemaReady: false };
  let query = async () => {
    throw new Error("postgres not configured");
  };
  const server = http.createServer((req, res) => {
    void createHandler({ query: (text, params) => query(text, params), state })(req, res);
  });

  server.listen(PORT, "0.0.0.0", () => {
    console.log(`listening on 0.0.0.0:${PORT}`);
    const raw = process.env.SKYROCKET_POSTGRES_URL;
    if (!raw) {
      console.error("SKYROCKET_POSTGRES_URL not set");
      return;
    }
    const pool = new Pool({ connectionString: rewritePostgresUrl(raw) });
    watchPool(pool);
    query = (text, params) => pool.query(text, params);
    const migrate = async () => {
      try {
        await query(CREATE_TABLE);
        await query(CREATE_INDEX);
        state.postgres = true;
        state.schemaReady = true;
      } catch (err) {
        state.postgres = false;
        state.schemaReady = false;
        console.error("comments schema migration failed", err);
        setTimeout(() => {
          void migrate();
        }, 5000);
      }
    };
    void migrate();
  });
}

module.exports = {
  rewritePostgresUrl,
  mapComment,
  createHandler,
  watchPool,
  start,
};

if (require.main === module) {
  start();
}
