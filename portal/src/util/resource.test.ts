/* global describe, it, expect */
import {
  diffResourceUpdates,
  ResourceDefinition,
  resourcePath,
  validateLocales,
} from "./resource";

describe("render and parse template string", () => {
  it("round trip", () => {
    const template = resourcePath`templates/${"locale"}/${"type"}/dummy.html`;
    const input = "templates/en/messages/dummy.html";
    const expectedResult = {
      locale: "en",
      type: "messages",
    };
    const parsed = template.parse(input);
    expect(parsed).toEqual(expectedResult);

    const rendered = template.render(parsed!);
    expect(rendered).toEqual(input);
  });
});

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
  optional: true,
};

describe("validate locales", () => {
  it("should pass for valid locales & resources", () => {
    const result = validateLocales(
      "en",
      ["en"],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
        {
          specifier: { def: ResourceB, locale: "en" },
          path: "templates/en/b.html",
          value: "resource B",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: null,
      invalidLocales: [],
    });
  });
  it("should fail for locales without resources", () => {
    const result = validateLocales(
      "en",
      ["en", "zh"],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
        {
          specifier: { def: ResourceB, locale: "en" },
          path: "templates/en/b.html",
          value: "resource B",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: "locale-without-resources",
      invalidLocales: ["zh"],
    });
  });
  it("should allow removing optional resource for default locale", () => {
    const result = validateLocales(
      "en",
      ["en"],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
        {
          specifier: { def: ResourceB, locale: "en" },
          path: "templates/en/b.html",
          value: "",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: null,
      invalidLocales: [],
    });
  });
  it("should allow removing resources for non-default locale", () => {
    const result = validateLocales(
      "en",
      ["en"],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
        {
          specifier: { def: ResourceA, locale: "zh" },
          path: "templates/en/a.html",
          value: "",
        },
        {
          specifier: { def: ResourceB, locale: "zh" },
          path: "templates/en/b.html",
          value: "",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: null,
      invalidLocales: [],
    });
  });
  it("should allow removing required resources for default built-in locale", () => {
    const result = validateLocales(
      "en",
      ["en"],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: null,
      invalidLocales: [],
    });
  });
  it("should forbid removing required resources for default non-built-in locale", () => {
    const result = validateLocales(
      "zh",
      ["zh"],
      [
        {
          specifier: { def: ResourceA, locale: "zh" },
          path: "templates/zh/a.html",
          value: "",
        },
      ]
    );
    expect(result).toEqual({
      invalidReason: "required-default-resource",
      invalidLocales: ["zh"],
    });
  });
});

describe("diff resources update", () => {
  it("should generate no changes", () => {
    const diff = diffResourceUpdates(
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: false,
      newResources: [],
      editedResources: [],
      deletedResources: [],
    });
  });
  it("should generate added resources", () => {
    const diff = diffResourceUpdates(
      [],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ],
      editedResources: [],
      deletedResources: [],
    });
  });
  it("should generate deleted resources", () => {
    const diff = diffResourceUpdates(
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ],
      []
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [],
      editedResources: [],
      deletedResources: [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: null,
        },
      ],
    });
  });
  it("should generate edited resources", () => {
    const diff = diffResourceUpdates(
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A!!",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [],
      editedResources: [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A!!",
        },
      ],
      deletedResources: [],
    });
  });
  it("should all resource updates", () => {
    const diff = diffResourceUpdates(
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A",
        },
        {
          specifier: { def: ResourceB, locale: "en" },
          path: "templates/en/b.html",
          value: "resource B",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A!!",
        },
        {
          specifier: { def: ResourceB, locale: "zh" },
          path: "templates/zh/b.html",
          value: "resource B",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [
        {
          specifier: { def: ResourceB, locale: "zh" },
          path: "templates/zh/b.html",
          value: "resource B",
        },
      ],
      editedResources: [
        {
          specifier: { def: ResourceA, locale: "en" },
          path: "templates/en/a.html",
          value: "resource A!!",
        },
      ],
      deletedResources: [
        {
          specifier: { def: ResourceB, locale: "en" },
          path: "templates/en/b.html",
          value: null,
        },
      ],
    });
  });
});
