import { addons } from "storybook/manager-api";

export type Comment = {
  id: string;
  storyId: string;
  author: string;
  body: string;
  createdAt: string;
  resolved?: boolean;
};

export function commentsApiPath(): string {
  const cfg = addons.getConfig() as { commentsApiPath?: string };
  const path = cfg.commentsApiPath?.trim() || "/api";
  return path.replace(/\/$/, "");
}

async function readError(res: Response): Promise<string> {
  const err = (await res.json().catch(() => ({}))) as { error?: string };
  return err.error || `HTTP ${res.status}`;
}

export async function fetchComments(storyId: string): Promise<Comment[]> {
  const res = await fetch(
    `${commentsApiPath()}/comments?storyId=${encodeURIComponent(storyId)}`
  );
  if (!res.ok) throw new Error(await readError(res));
  const data = (await res.json()) as { comments?: Comment[] };
  return (data.comments ?? []).map((comment) => ({
    ...comment,
    resolved: Boolean(comment.resolved),
  }));
}

export async function postComment(input: {
  storyId: string;
  author: string;
  body: string;
}): Promise<Comment> {
  const res = await fetch(`${commentsApiPath()}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!res.ok) throw new Error(await readError(res));
  const data = (await res.json()) as { comment: Comment };
  return data.comment;
}

export async function patchCommentResolved(
  id: string,
  resolved: boolean
): Promise<Comment> {
  const res = await fetch(
    `${commentsApiPath()}/comments/${encodeURIComponent(id)}/resolve`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ resolved }),
    }
  );
  if (!res.ok) throw new Error(await readError(res));
  const data = (await res.json()) as { comment: Comment };
  return { ...data.comment, resolved: Boolean(data.comment.resolved) };
}

export async function deleteComment(id: string): Promise<void> {
  const res = await fetch(
    `${commentsApiPath()}/comments/${encodeURIComponent(id)}/delete`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    }
  );
  if (!res.ok) throw new Error(await readError(res));
}

export async function fetchUnresolvedCounts(): Promise<Record<string, number>> {
  const res = await fetch(`${commentsApiPath()}/comments/counts`);
  if (!res.ok) throw new Error(await readError(res));
  const data = (await res.json()) as { counts?: Record<string, number> };
  return data.counts ?? {};
}
