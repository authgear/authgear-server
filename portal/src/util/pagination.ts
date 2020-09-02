export interface GetPaginationRenderDataInput {
  offset: number;
  pageSize: number;
  totalCount?: number;
}

export interface PaginationRenderData {
  currentOffset: number;
  offsets: number[];
  firstPageButtonEnabled: boolean;
  prevPageButtonEnabled: boolean;
  nextPageButtonEnabled: boolean;
  lastPageButtonEnabled: boolean;
  maxOffset?: number;
}

// The number of pages to show in addition to the current page.
// For example, if the current page is 47, and this value is 3,
// 44 45 46 47 48 49 50 should be shown.
const ADDITIONAL_PAGE_TO_SHOW = 3;

// eslint-disable-next-line complexity
export function getPaginationRenderData(
  input: GetPaginationRenderDataInput
): PaginationRenderData {
  const { offset, pageSize, totalCount } = input;

  if (pageSize === 0) {
    throw new Error("pageSize cannot be zero");
  }

  if (offset < 0) {
    throw new Error("offset cannot be negative");
  }

  if (offset % pageSize !== 0) {
    throw new Error("offset must be multiple of pageSize");
  }

  if (totalCount == null) {
    const offsets = [];
    for (
      let i = offset - ADDITIONAL_PAGE_TO_SHOW * pageSize;
      i <= offset + pageSize;
      i += pageSize
    ) {
      if (i >= 0) {
        offsets.push(i);
      }
    }

    // Allow going to first page if it is not at the first page.
    const firstPageButtonEnabled = offset !== 0;
    // Disallow going to last page because we do not have information to determine the last page.
    const lastPageButtonEnabled = false;
    // Allow going to previous page if it is not at the first page;
    const prevPageButtonEnabled = offset !== 0;
    // Allow going to next page because we do not have information to conclude there is no next page.
    const nextPageButtonEnabled = true;

    return {
      currentOffset: offset,
      offsets,
      firstPageButtonEnabled,
      lastPageButtonEnabled,
      prevPageButtonEnabled,
      nextPageButtonEnabled,
    };
  }

  // totalCount is available.
  if (totalCount < 0) {
    throw new Error("totalCount cannot be negative");
  }

  const div = Math.floor(totalCount / pageSize);
  const mod = totalCount % pageSize;

  // Suppose totalCount is 10 and pageSize is 5, then maxOffset is 5, not 10.
  // Suppose totalCount is 11 and pageSize is 5, then maxOffset is 10.
  // Therefore, if pageSize can divide totalCount, maxOffset has to be 1 pageSize smaller.
  // Anyway, maxOffset must be non-negative.
  let maxOffset = div * pageSize;
  if (mod === 0) {
    maxOffset -= pageSize;
  }
  if (maxOffset < 0) {
    maxOffset = 0;
  }

  const offsets = [];
  for (
    let i = offset - ADDITIONAL_PAGE_TO_SHOW * pageSize;
    i <= offset + ADDITIONAL_PAGE_TO_SHOW * pageSize;
    i += pageSize
  ) {
    if (i < 0 || i >= totalCount) {
      continue;
    }

    offsets.push(i);
  }

  // Allow going to first page if it is not at the first page.
  const firstPageButtonEnabled = offset !== 0;

  // Allow going to last page if it is not at the last page.
  const lastPageButtonEnabled = offset < maxOffset;

  // Allow going to previous page if it is not at the first page;
  const prevPageButtonEnabled = offset !== 0;

  // Allow going to next page if it is not at the last page.
  const nextPageButtonEnabled = offset < maxOffset;

  return {
    currentOffset: offset,
    offsets,
    firstPageButtonEnabled,
    lastPageButtonEnabled,
    prevPageButtonEnabled,
    nextPageButtonEnabled,
    maxOffset,
  };
}

export function encodeOffsetToCursor(offset: number): string {
  return btoa("offset:" + String(offset))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
