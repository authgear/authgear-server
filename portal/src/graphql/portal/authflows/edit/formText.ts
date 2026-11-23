import {
  AuthenticationFlowFormConfig,
  AuthenticationFlowStepConfig,
  CustomAttributesAttributeConfig,
  PortalAPIAppConfig,
} from "../../../../types";
import { unescapeMessageFormatText } from "../../../../util/messageFormat";
import { isAllowedLinkURL } from "./linkLabel";

// The text a form shows, kept per form name in each language's
// translation.json. See docs/specs/user-profile/profile-filling.md.

export type FormTextPart =
  | "title"
  | "description"
  | { label: string }
  | { placeholder: string };

const KEY_PREFIX = "v2.page.fill-form.";
// The built-in form text; any form without its own text uses it.
const DEFAULT_FORM_NAME = "default";

// The form a text key belongs to; null for other keys.
function formNameOfKey(key: string): string | null {
  if (!key.startsWith(KEY_PREFIX)) {
    return null;
  }
  const rest = key.slice(KEY_PREFIX.length);
  const dot = rest.indexOf(".");
  return dot > 0 ? rest.slice(0, dot) : null;
}

function renamedKey(key: string, from: string, to: string): string {
  return `${KEY_PREFIX}${to}.${key.slice(`${KEY_PREFIX}${from}.`.length)}`;
}

export function formTextKey(formName: string, part: FormTextPart): string {
  const prefix = `${KEY_PREFIX}${formName}.`;
  if (part === "title" || part === "description") {
    return prefix + part;
  }
  if ("label" in part) {
    return `${prefix}field-label-${part.label}`;
  }
  return `${prefix}field-placeholder-${part.placeholder}`;
}

// The field types Auth UI shows a placeholder on.
export const PLACEHOLDER_FIELD_TYPES: readonly string[] = [
  "string",
  "email",
  "url",
  "integer",
  "number",
  "phone_number",
];

// A form field's type, as the Authflow API reports it (fill_form_data in
// docs/specs/authentication-flow-api-reference.md).
export function formFieldType(
  pointer: string,
  customAttributes: CustomAttributesAttributeConfig[]
): string {
  const custom = customAttributes.find((a) => a.pointer === pointer);
  if (custom != null) {
    return custom.type;
  }
  switch (pointer) {
    case "/birthdate":
      return "date";
    case "/zoneinfo":
      return "timezone";
    case "/locale":
      return "enum";
    case "/profile":
    case "/picture":
    case "/website":
      return "url";
    case "/address":
      return "address";
    default:
      return "string";
  }
}

function isFormKey(key: string, formName: string): boolean {
  return key.startsWith(`${KEY_PREFIX}${formName}.`);
}

function collectForms(
  steps: AuthenticationFlowStepConfig[] | undefined,
  out: AuthenticationFlowFormConfig[]
): void {
  for (const step of steps ?? []) {
    if (step.form != null) {
      out.push(step.form);
    }
    for (const option of step.one_of ?? []) {
      if (option.form != null) {
        out.push(option.form);
      }
      collectForms(option.steps, out);
    }
  }
}

// Every form name in authgear.yaml, in every flow type.
export function usedFormNames(raw: PortalAPIAppConfig): Set<string> {
  const forms: AuthenticationFlowFormConfig[] = [];
  for (const flows of Object.values(raw.authentication_flow ?? {})) {
    for (const flow of flows ?? []) {
      collectForms(flow.steps, forms);
    }
  }
  return new Set(
    forms.flatMap((f) => (f.name != null && f.name !== "" ? [f.name] : []))
  );
}

export function hasFormKeys(
  json: Record<string, unknown>,
  formName: string
): boolean {
  return Object.keys(json).some((k) => isFormKey(k, formName));
}

// A name for a new form: one no form uses and no text is kept under, so the
// form starts with no text.
export function nextFormName(
  used: Set<string>,
  files: Record<string, Record<string, unknown> | null>
): string {
  for (let n = 1; ; n++) {
    const name = `form_${n}`;
    if (
      !used.has(name) &&
      !Object.values(files).some((f) => f != null && hasFormKeys(f, name))
    ) {
      return name;
    }
  }
}

// Whether an unquoted { remains: a MessageFormat argument.
function hasArgument(pattern: string): boolean {
  let quoted = false;
  for (let i = 0; i < pattern.length; i++) {
    const ch = pattern[i];
    if (ch === "'") {
      if (pattern[i + 1] === "'") {
        i += 1;
        continue;
      }
      quoted = !quoted;
      continue;
    }
    if (!quoted && (ch === "{" || ch === "}")) {
      return true;
    }
  }
  return false;
}

const TAG = /<\/?[a-zA-Z][^>]*>/g;
const LINK_TAG = /^(?:<a href="([^"<>]*)">|<\/a>)$/;

// Whether the panel can show and save a translation value without changing
// what it renders: no MessageFormat arguments, and no markup but links where
// links are allowed.
export function isEditableText(value: string, allowLinks: boolean): boolean {
  if (hasArgument(value)) {
    return false;
  }
  const tags = unescapeMessageFormatText(value).match(TAG) ?? [];
  return tags.every((t) => {
    const m = LINK_TAG.exec(t);
    // Undefined for a closing tag.
    const href = m?.[1];
    // Only links the builder would make: http(s) URLs.
    return allowLinks && m != null && (href == null || isAllowedLinkURL(href));
  });
}

// Text edits by language, then key; null removes the key.
export type TextEdits = Record<string, Record<string, string | null>>;

export type FormNameOp =
  | { kind: "rename"; from: string; to: string }
  | { kind: "delete"; name: string };

function applyOp(
  file: Record<string, unknown>,
  op: FormNameOp,
  usedAfter: Set<string>
): void {
  const name = op.kind === "rename" ? op.from : op.name;
  for (const key of Object.keys(file)) {
    if (!isFormKey(key, name)) {
      continue;
    }
    if (op.kind === "rename") {
      file[renamedKey(key, name, op.to)] = file[key];
    }
    if (!usedAfter.has(name)) {
      delete file[key];
    }
  }
}

// The translation files to write on Save: each changed language's latest
// file with the name ops, then the edits, applied. Unchanged files are left
// out. A language without a file starts from an empty one.
export function planTextWrites(
  latest: Record<string, Record<string, unknown> | null>,
  edits: TextEdits,
  ops: FormNameOp[],
  usedAfter: Set<string>
): Record<string, Record<string, unknown>> {
  const out: Record<string, Record<string, unknown>> = {};
  const languages = new Set([...Object.keys(latest), ...Object.keys(edits)]);
  for (const language of languages) {
    const before = latest[language] ?? {};
    const file: Record<string, unknown> = { ...before };
    for (const op of ops) {
      applyOp(file, op, usedAfter);
    }
    for (const [key, value] of Object.entries(edits[language] ?? {})) {
      // Text of a form that is gone, e.g. edited and then deleted.
      const name = formNameOfKey(key);
      if (name != null && name !== DEFAULT_FORM_NAME && !usedAfter.has(name)) {
        continue;
      }
      if (value == null) {
        delete file[key];
      } else {
        file[key] = value;
      }
    }
    const changed =
      latest[language] == null
        ? Object.keys(edits[language] ?? {}).length > 0
        : JSON.stringify(file) !== JSON.stringify(before);
    if (changed) {
      out[language] = file;
    }
  }
  return out;
}

type Files = Record<string, Record<string, unknown> | null>;

// Pending edits with a renamed form's keys moved to its new name.
export function renameEdits(
  edits: TextEdits,
  from: string,
  to: string
): TextEdits {
  const out: TextEdits = {};
  for (const [language, byKey] of Object.entries(edits)) {
    out[language] = {};
    for (const [key, value] of Object.entries(byKey)) {
      out[language][
        formNameOfKey(key) === from ? renamedKey(key, from, to) : key
      ] = value;
    }
  }
  return out;
}

// The saved files as the renames so far would leave them, to show a
// renamed form's text before Save. Old keys stay: another form may use them.
export function filesWithRenames(files: Files, ops: FormNameOp[]): Files {
  const renames = ops.filter(
    (o): o is Extract<FormNameOp, { kind: "rename" }> => o.kind === "rename"
  );
  if (renames.length === 0) {
    return files;
  }
  const out: Files = {};
  for (const [language, file] of Object.entries(files)) {
    if (file == null) {
      out[language] = null;
      continue;
    }
    const next = { ...file };
    for (const op of renames) {
      for (const key of Object.keys(next)) {
        if (formNameOfKey(key) === op.from) {
          next[renamedKey(key, op.from, op.to)] = next[key];
        }
      }
    }
    out[language] = next;
  }
  return out;
}

// The project's own value of a key, pending edits applied; undefined when
// unset or pending removal.
export function readOwnText(
  edits: TextEdits,
  files: Files,
  language: string,
  key: string
): string | undefined {
  const edited = edits[language] as Record<string, string | null> | undefined;
  if (edited != null && key in edited) {
    return edited[key] ?? undefined;
  }
  const value = files[language]?.[key];
  return typeof value === "string" ? value : undefined;
}

// The value in effect: the own value, else the built-in one. A pending
// removal is unset, not the saved text it removes.
export function readEffectiveText(
  edits: TextEdits,
  files: Files,
  effective: Record<string, Record<string, unknown>>,
  language: string,
  key: string
): string | undefined {
  const edited = edits[language] as Record<string, string | null> | undefined;
  if (edited != null && key in edited) {
    return edited[key] ?? undefined;
  }
  const own = readOwnText(edits, files, language, key);
  if (own != null) {
    return own;
  }
  const value = (effective[language] as Record<string, unknown> | undefined)?.[
    key
  ];
  return typeof value === "string" ? value : undefined;
}

export function withoutEdit(
  edits: TextEdits,
  language: string,
  key: string
): TextEdits {
  const byKey = { ...edits[language] };
  delete byKey[key];
  return { ...edits, [language]: byKey };
}

// The edits made since snapshot: those it saved are dropped, later changes
// to the same keys kept.
export function subtractEdits(
  current: TextEdits,
  snapshot: TextEdits
): TextEdits {
  const out: TextEdits = {};
  for (const [language, byKey] of Object.entries(current)) {
    const saved = snapshot[language] as
      | Record<string, string | null>
      | undefined;
    const rest: Record<string, string | null> = {};
    for (const [key, value] of Object.entries(byKey)) {
      if (saved == null || !(key in saved) || saved[key] !== value) {
        rest[key] = value;
      }
    }
    if (Object.keys(rest).length > 0) {
      out[language] = rest;
    }
  }
  return out;
}
