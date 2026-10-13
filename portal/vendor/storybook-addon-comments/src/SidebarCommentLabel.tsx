import React, { useSyncExternalStore } from "react";
import type { API } from "storybook/manager-api";
import { getUnresolvedCounts, subscribeUnresolvedCounts } from "./countsStore";

type HashEntry = {
  id: string;
  name: string;
  type: string;
  children?: string[];
};

function countForEntry(
  item: HashEntry,
  api: API,
  counts: Record<string, number>,
  seen: Set<string>
): number {
  if (seen.has(item.id)) return 0;
  seen.add(item.id);
  if (item.type === "story" || item.type === "docs") {
    return counts[item.id] ?? 0;
  }
  const children = item.children ?? [];
  return children.reduce((sum, childId) => {
    const child = api.getData(childId) as HashEntry | undefined;
    if (!child) return sum;
    return sum + countForEntry(child, api, counts, seen);
  }, 0);
}

export function SidebarCommentLabel({
  item,
  api,
}: {
  item: HashEntry;
  api: API;
}) {
  const counts = useSyncExternalStore(
    subscribeUnresolvedCounts,
    getUnresolvedCounts,
    getUnresolvedCounts
  );
  const n = countForEntry(item, api, counts, new Set());
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 6,
        minWidth: 0,
      }}
    >
      <span style={{ overflow: "hidden", textOverflow: "ellipsis" }}>{item.name}</span>
      {n > 0 ? (
        <span
          title={`${n} unresolved comment${n === 1 ? "" : "s"}`}
          style={{
            flexShrink: 0,
            minWidth: 16,
            height: 16,
            padding: "0 5px",
            borderRadius: 8,
            background: "#0d9488",
            color: "#fff",
            fontSize: 10,
            fontWeight: 700,
            lineHeight: "16px",
            textAlign: "center",
          }}
        >
          {n > 99 ? "99+" : n}
        </span>
      ) : null}
    </span>
  );
}
