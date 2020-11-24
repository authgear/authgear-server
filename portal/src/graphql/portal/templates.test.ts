/* global describe, it, expect */
import { generateUpdates } from "./templates";
import { ResourceDefinition, resourcePath } from "../../util/resource";

const ResourceA: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/a.html`,
  type: "text",
};

const ResourceB: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/b.html`,
  type: "text",
};

describe("generateUpdates", () => {
  it("handles invalid addition", () => {
    const actual = generateUpdates(
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
      },
      ["en", "zh"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
      }
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: ["zh"],
      editions: [],
      invalidEditionLocales: [],
      deletions: [],
    });
  });

  it("handles valid addition", () => {
    const actual = generateUpdates(
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
      },
      ["en", "zh"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
        "templates/zh/a.html": {
          def: ResourceA,
          path: "templates/zh/a.html",
          locale: "zh",
          value: "zh a",
        },
      }
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [
        {
          def: ResourceA,
          path: "templates/zh/a.html",
          locale: "zh",
          value: "zh a",
        },
      ],
      invalidAdditionLocales: [],
      editions: [],
      invalidEditionLocales: [],
      deletions: [],
    });
  });

  it("does not allow implicit deletion of locale", () => {
    const actual = generateUpdates(
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
      },
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "",
        },
      }
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [
        {
          def: ResourceA,
          locale: "en",
          path: "templates/en/a.html",
          value: null,
        },
      ],
      invalidEditionLocales: ["en"],
      deletions: [],
    });
  });

  it("does not generate unnecessary editions", () => {
    const actual = generateUpdates(
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
        "templates/en/b.html": {
          def: ResourceB,
          path: "templates/en/b.html",
          locale: "en",
          value: "b",
        },
      },
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "edited a",
        },
        "templates/en/b.html": {
          def: ResourceB,
          path: "templates/en/b.html",
          locale: "en",
          value: "b",
        },
      }
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [
        {
          def: ResourceA,
          locale: "en",
          path: "templates/en/a.html",
          value: "edited a",
        },
      ],
      invalidEditionLocales: [],
      deletions: [],
    });
  });

  it("handles deletion", () => {
    const actual = generateUpdates(
      ["en", "zh"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
        "templates/zh/a.html": {
          def: ResourceA,
          path: "templates/zh/a.html",
          locale: "zh",
          value: "a",
        },
      },
      ["en"],
      {
        "templates/en/a.html": {
          def: ResourceA,
          path: "templates/en/a.html",
          locale: "en",
          value: "a",
        },
        "templates/zh/a.html": {
          def: ResourceA,
          path: "templates/zh/a.html",
          locale: "zh",
          value: "a",
        },
      }
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [],
      invalidEditionLocales: [],
      deletions: [
        {
          def: ResourceA,
          locale: "zh",
          path: "templates/zh/a.html",
          value: null,
        },
      ],
    });
  });
});
