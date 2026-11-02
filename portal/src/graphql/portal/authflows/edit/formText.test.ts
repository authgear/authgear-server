import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../../types";
import {
  filesWithRenames,
  formFieldType,
  formTextKey,
  isEditableText,
  nextFormName,
  planTextWrites,
  readEffectiveText,
  readOwnText,
  renameEdits,
  subtractEdits,
  withoutEdit,
  usedFormNames,
} from "./formText";

describe("formTextKey", () => {
  it("names each part of a form", () => {
    expect(formTextKey("about", "title")).toEqual(
      "v2.page.fill-form.about.title"
    );
    expect(formTextKey("about", "description")).toEqual(
      "v2.page.fill-form.about.description"
    );
    expect(formTextKey("about", { label: "/given_name" })).toEqual(
      "v2.page.fill-form.about.field-label-/given_name"
    );
    expect(formTextKey("about", { placeholder: "/given_name" })).toEqual(
      "v2.page.fill-form.about.field-placeholder-/given_name"
    );
  });
});

describe("usedFormNames", () => {
  it("collects forms from steps, options, nested steps and every flow type", () => {
    const raw: PortalAPIAppConfig = {
      authentication_flow: {
        signup_flows: [
          {
            name: "default",
            steps: [
              { type: "fill_form", form: { name: "a" } },
              {
                type: "identify",
                one_of: [{ identification: "email", form: { name: "b" } }],
              },
            ],
          },
        ],
        login_flows: [
          {
            name: "default",
            steps: [
              {
                type: "identify",
                one_of: [
                  {
                    identification: "email",
                    steps: [{ type: "fill_form", form: { name: "c" } }],
                  },
                ],
              },
            ],
          },
        ],
        promote_flows: [
          {
            name: "default",
            steps: [{ type: "fill_form", form: { name: "d" } }],
          },
        ],
      },
    } as PortalAPIAppConfig;
    expect(usedFormNames(raw)).toEqual(new Set(["a", "b", "c", "d"]));
  });
});

describe("nextFormName", () => {
  it("skips names in use and names that still have text", () => {
    expect(
      nextFormName(new Set(["form_1"]), {
        en: { "v2.page.fill-form.form_2.title": "x" },
      })
    ).toEqual("form_3");
  });
});

describe("isEditableText", () => {
  it("accepts plain and quoted text", () => {
    expect(isEditableText("Hello", false)).toBe(true);
    expect(isEditableText("'{'literal'}'", false)).toBe(true);
    expect(isEditableText("It''s", false)).toBe(true);
  });

  it("refuses MessageFormat arguments", () => {
    expect(isEditableText("{FlowType, select, other {x}}", false)).toBe(false);
  });

  it("accepts links only where links are allowed", () => {
    const linked = 'See <a href="https://x">T</a>';
    expect(isEditableText(linked, true)).toBe(true);
    expect(isEditableText(linked, false)).toBe(false);
    expect(isEditableText("<b>x</b>", true)).toBe(false);
  });
});

describe("planTextWrites", () => {
  it("changes only edited keys, on the latest file", () => {
    expect(
      planTextWrites(
        { en: { k2: "kept" }, ja: { k3: "x" } },
        { en: { k1: "new" } },
        [],
        new Set()
      )
    ).toEqual({ en: { k1: "new", k2: "kept" } });
  });

  it("deletes a key edited to null", () => {
    expect(
      planTextWrites(
        { en: { k1: "x", k2: "y" } },
        { en: { k1: null } },
        [],
        new Set()
      )
    ).toEqual({ en: { k2: "y" } });
  });

  it("moves a renamed form's text", () => {
    expect(
      planTextWrites(
        { en: { "v2.page.fill-form.a.title": "T", other: "o" } },
        {},
        [{ kind: "rename", from: "a", to: "b" }],
        new Set(["b"])
      )
    ).toEqual({ en: { "v2.page.fill-form.b.title": "T", other: "o" } });
  });

  it("copies a renamed form's text when another form keeps the old name", () => {
    expect(
      planTextWrites(
        { en: { "v2.page.fill-form.a.title": "T" } },
        {},
        [{ kind: "rename", from: "a", to: "b" }],
        new Set(["a", "b"])
      )
    ).toEqual({
      en: {
        "v2.page.fill-form.a.title": "T",
        "v2.page.fill-form.b.title": "T",
      },
    });
  });

  it("lets the renamed form's text win over text under the new name", () => {
    expect(
      planTextWrites(
        {
          en: {
            "v2.page.fill-form.a.title": "Mine",
            "v2.page.fill-form.b.title": "Theirs",
            "v2.page.fill-form.b.description": "Theirs too",
          },
        },
        {},
        [{ kind: "rename", from: "a", to: "b" }],
        new Set(["b"])
      )
    ).toEqual({
      en: {
        "v2.page.fill-form.b.title": "Mine",
        "v2.page.fill-form.b.description": "Theirs too",
      },
    });
  });

  it("keeps a deleted form's text another form still uses", () => {
    expect(
      planTextWrites(
        { en: { "v2.page.fill-form.a.title": "T" } },
        {},
        [{ kind: "delete", name: "a" }],
        new Set(["a"])
      )
    ).toEqual({});
  });

  it("removes a deleted form's text in every language", () => {
    expect(
      planTextWrites(
        {
          en: { "v2.page.fill-form.a.title": "T", k: "v" },
          ja: { "v2.page.fill-form.a.field-label-/x": "L" },
        },
        {},
        [{ kind: "delete", name: "a" }],
        new Set()
      )
    ).toEqual({ en: { k: "v" }, ja: {} });
  });

  it("does not touch a name that only shares a prefix", () => {
    expect(
      planTextWrites(
        { en: { "v2.page.fill-form.ab.title": "T" } },
        {},
        [{ kind: "delete", name: "a" }],
        new Set()
      )
    ).toEqual({});
  });

  it("starts a language with no file yet", () => {
    expect(
      planTextWrites({ ja: null }, { ja: { k: "v" } }, [], new Set())
    ).toEqual({ ja: { k: "v" } });
  });
});

describe("formFieldType", () => {
  const custom = [
    {
      id: "1",
      pointer: "/x_terms",
      type: "boolean" as const,
      access_control: {},
    },
    {
      id: "2",
      pointer: "/x_country",
      type: "country_code" as const,
      access_control: {},
    },
  ] as never[];

  it("follows the server's types for standard attributes", () => {
    expect(formFieldType("/given_name", [])).toEqual("string");
    expect(formFieldType("/gender", [])).toEqual("string");
    expect(formFieldType("/birthdate", [])).toEqual("date");
    expect(formFieldType("/zoneinfo", [])).toEqual("timezone");
    expect(formFieldType("/locale", [])).toEqual("enum");
    expect(formFieldType("/website", [])).toEqual("url");
    expect(formFieldType("/address", [])).toEqual("address");
  });

  it("uses a custom attribute's type", () => {
    expect(formFieldType("/x_terms", custom)).toEqual("boolean");
    expect(formFieldType("/x_country", custom)).toEqual("country_code");
  });
});

describe("planTextWrites with pending edits", () => {
  it("drops edits under a name no form uses after the save", () => {
    expect(
      planTextWrites(
        { en: { "v2.page.fill-form.a.title": "Old" } },
        { en: { "v2.page.fill-form.a.title": "New" } },
        [{ kind: "delete", name: "a" }],
        new Set()
      )
    ).toEqual({ en: {} });
  });

  it("keeps edits to the default form's text", () => {
    expect(
      planTextWrites(
        {},
        { en: { "v2.page.fill-form.default.title": "Hi" } },
        [],
        new Set()
      )
    ).toEqual({ en: { "v2.page.fill-form.default.title": "Hi" } });
  });
});

describe("renameEdits", () => {
  it("moves pending edits to the new name", () => {
    expect(
      renameEdits(
        { en: { "v2.page.fill-form.a.title": "New", other: "x" } },
        "a",
        "b"
      )
    ).toEqual({ en: { "v2.page.fill-form.b.title": "New", other: "x" } });
  });
});

describe("filesWithRenames", () => {
  it("shows a renamed form's saved text under its new name", () => {
    expect(
      filesWithRenames({ en: { "v2.page.fill-form.a.title": "T" }, ja: null }, [
        { kind: "rename", from: "a", to: "b" },
      ])
    ).toEqual({
      en: {
        "v2.page.fill-form.a.title": "T",
        "v2.page.fill-form.b.title": "T",
      },
      ja: null,
    });
  });
});

describe("readOwnText and readEffectiveText", () => {
  const files = { en: { k: "Saved" } };
  const effective = { en: { k: "Saved", d: "Built-in" } };

  it("prefer a pending edit", () => {
    expect(readOwnText({ en: { k: "New" } }, files, "en", "k")).toEqual("New");
  });

  it("treat a pending removal as unset, not as the saved text", () => {
    expect(readOwnText({ en: { k: null } }, files, "en", "k")).toBeUndefined();
    expect(
      readEffectiveText({ en: { k: null } }, files, effective, "en", "k")
    ).toBeUndefined();
  });

  it("fall back to the built-in text", () => {
    expect(readEffectiveText({}, files, effective, "en", "d")).toEqual(
      "Built-in"
    );
  });
});

describe("withoutEdit", () => {
  it("drops one pending edit", () => {
    expect(withoutEdit({ en: { k: "", j: "x" } }, "en", "k")).toEqual({
      en: { j: "x" },
    });
  });
});

describe("subtractEdits", () => {
  it("keeps edits made after the snapshot", () => {
    expect(
      subtractEdits(
        { en: { a: "1", b: "2" }, ja: { c: "3" } },
        { en: { a: "1", b: "old" } }
      )
    ).toEqual({ en: { b: "2" }, ja: { c: "3" } });
  });
});

describe("isEditableText and link URLs", () => {
  it("refuses a stored link that is not http(s)", () => {
    const script = ["javascript", "x()"].join(":");
    expect(isEditableText(`<a href="${script}">T</a>`, true)).toBe(false);
    expect(isEditableText('<a href="/terms">T</a>', true)).toBe(false);
  });
});
