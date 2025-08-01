import { describe, it, expect } from "@jest/globals";
import { getPaginationRenderData, encodeOffsetToCursor } from "./pagination";

describe("getPaginationRenderData", () => {
  it("no total count; 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 0,
        pageSize: 5,
      })
    ).toEqual({
      currentOffset: 0,
      offsets: [0],
      firstPageButtonEnabled: false,
      prevPageButtonEnabled: false,
      nextPageButtonEnabled: false,
      lastPageButtonEnabled: false,
    });
  });

  it("no total count; non 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 5,
        pageSize: 5,
      })
    ).toEqual({
      currentOffset: 5,
      offsets: [5],
      firstPageButtonEnabled: false,
      prevPageButtonEnabled: false,
      nextPageButtonEnabled: false,
      lastPageButtonEnabled: false,
    });
  });

  it("has divisible total count; 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 0,
        pageSize: 5,
        totalCount: 20,
      })
    ).toEqual({
      currentOffset: 0,
      offsets: [0, 5, 10, 15],
      firstPageButtonEnabled: false,
      prevPageButtonEnabled: false,
      nextPageButtonEnabled: true,
      lastPageButtonEnabled: true,
      maxOffset: 15,
    });
  });

  it("has divisible total count; non 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 5,
        pageSize: 5,
        totalCount: 20,
      })
    ).toEqual({
      currentOffset: 5,
      offsets: [0, 5, 10, 15],
      firstPageButtonEnabled: true,
      prevPageButtonEnabled: true,
      nextPageButtonEnabled: true,
      lastPageButtonEnabled: true,
      maxOffset: 15,
    });
  });

  it("has divisible total count; last page", () => {
    expect(
      getPaginationRenderData({
        offset: 15,
        pageSize: 5,
        totalCount: 20,
      })
    ).toEqual({
      currentOffset: 15,
      offsets: [0, 5, 10, 15],
      firstPageButtonEnabled: true,
      prevPageButtonEnabled: true,
      nextPageButtonEnabled: false,
      lastPageButtonEnabled: false,
      maxOffset: 15,
    });
  });

  it("has indivisible total count; 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 0,
        pageSize: 5,
        totalCount: 21,
      })
    ).toEqual({
      currentOffset: 0,
      offsets: [0, 5, 10, 15],
      firstPageButtonEnabled: false,
      prevPageButtonEnabled: false,
      nextPageButtonEnabled: true,
      lastPageButtonEnabled: true,
      maxOffset: 20,
    });
  });

  it("has indivisible total count; non 1st page", () => {
    expect(
      getPaginationRenderData({
        offset: 5,
        pageSize: 5,
        totalCount: 21,
      })
    ).toEqual({
      currentOffset: 5,
      offsets: [0, 5, 10, 15, 20],
      firstPageButtonEnabled: true,
      prevPageButtonEnabled: true,
      nextPageButtonEnabled: true,
      lastPageButtonEnabled: true,
      maxOffset: 20,
    });
  });

  it("has indivisible total count; last page", () => {
    expect(
      getPaginationRenderData({
        offset: 20,
        pageSize: 5,
        totalCount: 21,
      })
    ).toEqual({
      currentOffset: 20,
      offsets: [5, 10, 15, 20],
      firstPageButtonEnabled: true,
      prevPageButtonEnabled: true,
      nextPageButtonEnabled: false,
      lastPageButtonEnabled: false,
      maxOffset: 20,
    });
  });
});

describe("encodeOffsetToCursor", () => {
  it("encode with URL encoding without padding", () => {
    expect(encodeOffsetToCursor(1)).toEqual("b2Zmc2V0OjA");
  });

  it("undefined if offset is 0", () => {
    expect(encodeOffsetToCursor(0)).toEqual(undefined);
  });
});
