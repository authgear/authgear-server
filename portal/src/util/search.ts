import { useCallback } from "react";

type SearchItem = Record<string, unknown>;

export function exactKeywordSearch<X extends SearchItem, K extends keyof X>(
  list: X[],
  keyList: X[K] extends string | undefined ? K[] : never,
  searchString: string
): X[] {
  const needle = searchString.toLowerCase();
  const matchedItems = [];
  for (const item of list) {
    for (const key of keyList) {
      const lowered = ((item[key] ?? "") as string).toLowerCase();
      if (lowered.includes(needle)) {
        matchedItems.push(item);
      }
    }
  }
  return matchedItems;
}

export function useExactKeywordSearch<X extends SearchItem, K extends keyof X>(
  list: X[],
  keyList: X[K] extends string | undefined ? K[] : never
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
