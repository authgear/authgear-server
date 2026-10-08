import React, { useCallback, useEffect, useState } from "react";
import { useStorybookState } from "storybook/manager-api";
import {
  deleteComment,
  fetchComments,
  patchCommentResolved,
  postComment,
  type Comment,
} from "./api";
import { AUTHOR_KEY } from "./constants";
import { refreshUnresolvedCounts } from "./refreshCounts";

const FIELD_STYLE: React.CSSProperties = {
  padding: "8px 10px",
  border: "1px solid #cbd5e1",
  borderRadius: 6,
  background: "#fff",
  color: "#000",
};

export function CommentsPanel() {
  const { storyId } = useStorybookState();
  const [comments, setComments] = useState<Comment[]>([]);
  const [author, setAuthor] = useState(
    () => (typeof localStorage !== "undefined" && localStorage.getItem(AUTHOR_KEY)) || ""
  );
  const [body, setBody] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [posting, setPosting] = useState(false);

  const load = useCallback(async (id: string) => {
    setLoading(true);
    setError(null);
    try {
      setComments(await fetchComments(id));
      void refreshUnresolvedCounts();
    } catch (err) {
      setComments([]);
      setError(
        err instanceof Error
          ? err.message
          : "Could not load comments. Deploy Storybook with the comments API to enable this panel."
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!storyId) return;
    void load(storyId);
  }, [storyId, load]);

  const onSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!storyId || !author.trim() || !body.trim()) return;
    setPosting(true);
    setError(null);
    try {
      localStorage.setItem(AUTHOR_KEY, author.trim());
      const created = await postComment({
        storyId,
        author: author.trim(),
        body: body.trim(),
      });
      setComments((current) => [...current, created]);
      setBody("");
      void refreshUnresolvedCounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not post comment.");
    } finally {
      setPosting(false);
    }
  };

  const onToggleResolved = async (comment: Comment, resolved: boolean) => {
    setError(null);
    try {
      const updated = await patchCommentResolved(comment.id, resolved);
      setComments((current) => current.map((item) => (item.id === updated.id ? updated : item)));
      void refreshUnresolvedCounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update comment.");
    }
  };

  const onDelete = async (comment: Comment) => {
    setError(null);
    try {
      await deleteComment(comment.id);
      setComments((current) => current.filter((item) => item.id !== comment.id));
      void refreshUnresolvedCounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete comment.");
    }
  };

  return (
    <div
      data-oursky-comments="true"
      style={{ padding: 16, fontSize: 13, lineHeight: 1.45, maxWidth: 640 }}
    >
      <p style={{ margin: "0 0 12px", color: "#64748b" }}>
        Comments are stored for the current story
        {storyId ? <code>{` ${storyId}`}</code> : null}. Each row has resolve and
        delete on the right.
      </p>
      {error ? <p style={{ margin: "0 0 12px", color: "#b91c1c" }}>{error}</p> : null}
      {loading ? <p style={{ margin: "0 0 12px" }}>Loading…</p> : null}
      <ul style={{ listStyle: "none", padding: 0, margin: "0 0 16px" }}>
        {comments.length === 0 && !loading && !error ? (
          <li style={{ color: "#64748b" }}>No comments yet.</li>
        ) : (
          comments.map((comment) => (
            <li
              key={comment.id}
              style={{
                display: "flex",
                gap: 12,
                alignItems: "flex-start",
                justifyContent: "space-between",
                borderBottom: "1px solid #e2e8f0",
                padding: "10px 0",
                opacity: comment.resolved ? 0.65 : 1,
              }}
            >
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{ fontWeight: 600 }}>{comment.author}</div>
                <div
                  style={{
                    whiteSpace: "pre-wrap",
                    textDecoration: comment.resolved ? "line-through" : "none",
                  }}
                >
                  {comment.body}
                </div>
                <div style={{ color: "#94a3b8", fontSize: 11, marginTop: 4 }}>
                  {new Date(comment.createdAt).toLocaleString()}
                </div>
              </div>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  flexShrink: 0,
                  paddingTop: 2,
                }}
              >
                <label
                  title="Resolved"
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 4,
                    fontSize: 12,
                    whiteSpace: "nowrap",
                    cursor: "pointer",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={Boolean(comment.resolved)}
                    onChange={(e) => void onToggleResolved(comment, e.target.checked)}
                    style={{
                      width: 14,
                      height: 14,
                      accentColor: "#0d9488",
                      background: "#fff",
                      backgroundColor: "#fff",
                      colorScheme: "light",
                    }}
                  />
                  <svg
                    width={14}
                    height={14}
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke={comment.resolved ? "#0d9488" : "#94a3b8"}
                    strokeWidth={2}
                    aria-hidden
                  >
                    <path d="M20 6L9 17l-5-5" />
                  </svg>
                </label>
                <button
                  type="button"
                  title="Delete"
                  aria-label="Delete comment"
                  onClick={() => void onDelete(comment)}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    width: 28,
                    height: 28,
                    padding: 0,
                    border: "1px solid #cbd5e1",
                    borderRadius: 4,
                    background: "#fff",
                    color: "#b91c1c",
                    cursor: "pointer",
                  }}
                >
                  <svg
                    width={14}
                    height={14}
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2}
                    aria-hidden
                  >
                    <polyline points="3 6 5 6 21 6" />
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
                    <path d="M8 6V4h8v2" />
                  </svg>
                </button>
              </div>
            </li>
          ))
        )}
      </ul>
      <form onSubmit={(e) => void onSubmit(e)} style={{ display: "grid", gap: 8 }}>
        <input
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Your name"
          maxLength={80}
          required
          style={FIELD_STYLE}
        />
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Leave a comment on this story"
          maxLength={4000}
          required
          rows={4}
          style={{ ...FIELD_STYLE, resize: "vertical" }}
        />
        <button
          type="submit"
          disabled={posting || !storyId}
          style={{
            justifySelf: "start",
            padding: "8px 14px",
            border: 0,
            borderRadius: 6,
            background: "#0d9488",
            color: "white",
            fontWeight: 600,
            cursor: posting ? "wait" : "pointer",
          }}
        >
          {posting ? "Posting…" : "Post comment"}
        </button>
      </form>
    </div>
  );
}
