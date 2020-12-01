/* global describe, it, expect */
import { generateUpdates } from "./templates";
import { ResourceDefinition, resourcePath } from "../../util/resource";

const ResourceA: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/a.html`,
  type: "text",
  usesEffectiveDataAsFallbackValue: true,
  extensions: [],
};

const ResourceB: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/b.html`,
  type: "text",
  usesEffectiveDataAsFallbackValue: true,
  extensions: [],
};

describe("generateUpdates", () => {
  it("handles invalid addition", () => {
    const actual = generateUpdates(
      ["en"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
      ],
      ["en", "zh"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
      ]
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
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
      ],
      ["en", "zh"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
        {
          specifier: {
            def: ResourceA,
            locale: "zh",
          },
          path: "templates/zh/a.html",
          value: "zh a",
        },
      ]
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [
        {
          specifier: {
            def: ResourceA,
            locale: "zh",
          },
          path: "templates/zh/a.html",
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
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
      ],
      ["en"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "",
        },
      ]
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [],
      invalidEditionLocales: ["en"],
      deletions: [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: null,
        },
      ],
    });
  });

  it("does not generate unnecessary editions", () => {
    const actual = generateUpdates(
      ["en"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
        {
          specifier: {
            def: ResourceB,
            locale: "en",
          },
          path: "templates/en/b.html",
          value: "b",
        },
      ],
      ["en"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "edited a",
        },
        {
          specifier: {
            def: ResourceB,
            locale: "en",
          },
          path: "templates/en/b.html",
          value: "b",
        },
      ]
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
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
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
        {
          specifier: {
            def: ResourceA,
            locale: "zh",
          },
          path: "templates/zh/a.html",
          value: "a",
        },
      ],
      ["en"],
      [
        {
          specifier: {
            def: ResourceA,
            locale: "en",
          },
          path: "templates/en/a.html",
          value: "a",
        },
        {
          specifier: {
            def: ResourceA,
            locale: "zh",
          },
          path: "templates/zh/a.html",
          value: "a",
        },
      ]
    );

    expect(actual).toEqual({
      isModified: true,
      additions: [],
      invalidAdditionLocales: [],
      editions: [],
      invalidEditionLocales: [],
      deletions: [
        {
          specifier: {
            def: ResourceA,
            locale: "zh",
          },
          path: "templates/zh/a.html",
          value: null,
        },
      ],
    });
  });
});
