import { useCallback } from "react";

export function exactKeywordSearch<X, K extends keyof X>(
  list: X[],
  keyList: X[K] extends string | undefined | null ? K[] : never,
  searchString: string
): X[] {
  const matchedSet = new Set();
  const needle = searchString.toLowerCase();
  const matchedItems = [];
  for (const item of list) {
    for (const key of keyList) {
      const value = item[key];
      if (typeof value === "string") {
        const lowered = value.toLowerCase();
        const isMatch = lowered.includes(needle);
        const matched = matchedSet.has(item);
        if (isMatch && !matched) {
          matchedSet.add(item);
          matchedItems.push(item);
        }
      }
    }
  }
  return matchedItems;
}

export function useExactKeywordSearch<X, K extends keyof X>(
  list: X[],
  keyList: X[K] extends string | undefined | null ? K[] : never
): {
  search: (searchString: string) => X[];
} {
  const search = useCallback(
    (searchString: string) => {
      if (searchString.trim() === "") {
        return list;
      }
      const matchedItems = exactKeywordSearch(list, keyList, searchString);
      return matchedItems;
    },
    [list, keyList]
  );

  return { search };
}
