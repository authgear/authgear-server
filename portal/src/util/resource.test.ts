/**
 * We have to use node to run this specifically because jsdom does not have TextEncoder.
 * See https://github.com/jsdom/jsdom/issues/2524
 *
 * @jest-environment node
 */
import { describe, it, expect } from "@jest/globals";
import {
  diffResourceUpdates,
  ResourceDefinition,
  resourcePath,
  decodeForText,
  encodeForText,
  decodeForPrettifiedJSON,
  encodeForPrettifiedJSON,
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
  extensions: [],
};

const ResourceB: ResourceDefinition = {
  resourcePath: resourcePath`templates/${"locale"}/b.html`,
  type: "text",
  extensions: [],
  optional: true,
};

describe("diff resources update", () => {
  it("should generate no changes", () => {
    const diff = diffResourceUpdates(
      [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
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
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
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
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
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
          specifier: { def: ResourceA, locale: "en", extension: null },
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
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A!!",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [],
      editedResources: [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
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
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A",
        },
        {
          specifier: { def: ResourceB, locale: "en", extension: null },
          path: "templates/en/b.html",
          nullableValue: "resource B",
        },
      ],
      [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          nullableValue: "resource A!!",
        },
        {
          specifier: { def: ResourceB, locale: "zh", extension: null },
          path: "templates/zh/b.html",
          nullableValue: "resource B",
        },
      ]
    );
    expect(diff).toEqual({
      needUpdate: true,
      newResources: [
        {
          specifier: { def: ResourceB, locale: "zh", extension: null },
          path: "templates/zh/b.html",
          value: "resource B",
        },
      ],
      editedResources: [
        {
          specifier: { def: ResourceA, locale: "en", extension: null },
          path: "templates/en/a.html",
          value: "resource A!!",
        },
      ],
      deletedResources: [
        {
          specifier: { def: ResourceB, locale: "en", extension: null },
          path: "templates/en/b.html",
          value: null,
        },
      ],
    });
  });
});

describe("text encoding/decoding functions", () => {
  describe("decodeForText", () => {
    it("should decode base64 encoded text", () => {
      const encoded = "SGVsbG8sIFdvcmxkIQ=="; // "Hello, World!" in base64
      const result = decodeForText(encoded);
      expect(result).toBe("Hello, World!");
    });

    it("should decode UTF-8 text", () => {
      const encoded = "8J+YgA=="; // "😀" emoji in base64
      const result = decodeForText(encoded);
      expect(result).toBe("😀");
    });

    it("should handle empty string", () => {
      const encoded = ""; // empty string in base64
      const result = decodeForText(encoded);
      expect(result).toBe("");
    });
  });

  describe("encodeForText", () => {
    it("should encode text to base64", () => {
      const text = "Hello, World!";
      const result = encodeForText(text);
      expect(result).toBe("SGVsbG8sIFdvcmxkIQ==");
    });

    it("should encode UTF-8 text", () => {
      const text = "😀";
      const result = encodeForText(text);
      expect(result).toBe("8J+YgA==");
    });

    it("should handle empty string", () => {
      const text = "";
      const result = encodeForText(text);
      expect(result).toBe("");
    });

    it("should round trip with decodeForText", () => {
      const original = "Hello, 世界! 🌍";
      const encoded = encodeForText(original);
      const decoded = decodeForText(encoded);
      expect(decoded).toBe(original);
    });
  });
});

describe("prettified JSON encoding/decoding functions", () => {
  describe("decodeForPrettifiedJSON", () => {
    it("should decode and prettify compact JSON", () => {
      const compactJson = '{"name":"John","age":30,"city":"New York"}';
      const encoded = encodeForText(compactJson);
      const result = decodeForPrettifiedJSON(encoded);
      const expected = JSON.stringify(
        { name: "John", age: 30, city: "New York" },
        null,
        2
      );
      expect(result).toBe(expected);
    });

    it("should handle nested objects", () => {
      const compactJson =
        '{"user":{"name":"Alice","preferences":{"theme":"dark","lang":"en"}}}';
      const encoded = encodeForText(compactJson);
      const result = decodeForPrettifiedJSON(encoded);
      const expected = JSON.stringify(
        {
          user: {
            name: "Alice",
            preferences: {
              theme: "dark",
              lang: "en",
            },
          },
        },
        null,
        2
      );
      expect(result).toBe(expected);
    });

    it("should handle arrays", () => {
      const compactJson = '[{"id":1,"name":"Item 1"},{"id":2,"name":"Item 2"}]';
      const encoded = encodeForText(compactJson);
      const result = decodeForPrettifiedJSON(encoded);
      const expected = JSON.stringify(
        [
          { id: 1, name: "Item 1" },
          { id: 2, name: "Item 2" },
        ],
        null,
        2
      );
      expect(result).toBe(expected);
    });
  });

  describe("encodeForPrettifiedJSON", () => {
    it("should compact and encode prettified JSON", () => {
      const prettifiedJson = `{
  "name": "John",
  "age": 30,
  "city": "New York"
}`;
      const result = encodeForPrettifiedJSON(prettifiedJson);
      const compactJson = '{"name":"John","age":30,"city":"New York"}';
      const expected = encodeForText(compactJson);
      expect(result).toBe(expected);
    });

    it("should handle nested objects", () => {
      const prettifiedJson = `{
  "user": {
    "name": "Alice",
    "preferences": {
      "theme": "dark",
      "lang": "en"
    }
  }
}`;
      const result = encodeForPrettifiedJSON(prettifiedJson);
      const compactJson =
        '{"user":{"name":"Alice","preferences":{"theme":"dark","lang":"en"}}}';
      const expected = encodeForText(compactJson);
      expect(result).toBe(expected);
    });

    it("encodeForPrettifiedJSON and decodeForPrettifiedJSON does not round trip", () => {
      const encoded = encodeForPrettifiedJSON(`{
  "name": "Test",
  "data": {
    "values": [1, 2, 3],
    "enabled": true
  }
}`);
      const decoded = decodeForPrettifiedJSON(encoded);
      expect(decoded).toBe(`{
  "name": "Test",
  "data": {
    "values": [
      1,
      2,
      3
    ],
    "enabled": true
  }
}`);
    });

    it("encodeForPrettifiedJSON return encodeForText if the input is invalid JSON", () => {
      const encoded = encodeForPrettifiedJSON(`{`);
      expect(encoded).toBe(encodeForText(`{`));
    });
  });
});
