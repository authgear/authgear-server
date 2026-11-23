# Form Text and Field Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A form step's title, description, and each field's label and placeholder are edited per language in the builder, in a form panel plus a second field panel, and saved with the flow.

**Architecture:** Two pure modules carry the rules: `edit/formText.ts` (keys, names, which text is safe to edit, what to write on Save) and `edit/linkLabel.ts` (checkbox labels with links). A `useFormText` hook loads `translation.json` for every supported language, holds pending text edits, and writes them after the config is saved. A `FormTextContext` gives the panels and the canvas the selected language and the text API.

**Tech Stack:** React 18, Radix Themes, immer, Jest (`@jest/globals`), Apollo (`useAppTemplatesQuery`, `useUpdateAppTemplatesMutation`).

**Spec:** `docs/plans/auth-flow-builder/2026-10-30-01-form-text-design.md` (DEV-3918). Builds on `2026-10-15-01-editor-design.md` and `2026-10-19-01-flow-list-design.md`. Server keys from DEV-3989 (#6053) and DEV-3990 (#6033).

## Global Constraints

- Branch `form-text`, stacked on `auth-flow-list` (PR #6008), worktree `~/.herdr/worktrees/authgear-server/auth-flow-editor`. Do not rebase or push #5995.
- **No commits until fung approves the UI.** Steps below edit and verify only; one commit at the end (Task 9), message ending `ref DEV-3918`. No browser verification unless asked; typecheck and lint are the gate.
- Keys, exactly:
  - `v2.page.fill-form.{name}.title`
  - `v2.page.fill-form.{name}.description` (empty string = hidden)
  - `v2.page.fill-form.{name}.field-label-{key}` (`{key}` is the pointer for a user profile field)
  - `v2.page.fill-form.{name}.field-placeholder-{key}`
- Form names match `^[a-zA-Z_][a-zA-Z0-9_]*$` and are never `default`.
- Placeholder only for field types `string`, `email`, `url`, `integer`, `number`, `phone_number`.
- Forms attached to an identify option get no text editing and no field panel; only their name is prefilled.
- Text is stored with `escapeMessageFormatText` and shown with `unescapeMessageFormatText` (`portal/src/util/messageFormat.ts`).
- All UI text through `locale-data/en.json` with static IDs; language names via `Locales.{tag}` (as `CustomTextConfigurationScreen`). Follow the `update-portal-ui` skill.
- Never mention Auth0 in code, comments, commits or PR text.
- Checks, in `portal/`: `npx jest src/graphql/portal/authflows`, `npm run typecheck`, `npx eslint <touched>`, `npx stylelint <touched css>`, `npx prettier --check <touched>`.

## Review Focus

- A translation value with MessageFormat arguments (`{FlowType, select, …}`) or, in a checkbox label, markup other than `<a href>` must survive opening the panel and saving byte for byte. Pinned in Task 1 (`isEditableText`) and Task 3 (unedited keys untouched).
- Two forms in different flows sharing a name: deleting or renaming one must not delete text the other still uses. Pinned in Task 1 (`planTextWrites`).
- `translation.json` changed in Custom Text after the editor loaded: Save keeps those keys. Pinned in Task 1 (writes are merged onto the latest file).
- A language with no `translation.json` yet: the first edit creates the file. Pinned in Task 1.
- A link URL that is `javascript:` or relative is refused; label text containing `<` without a link is stored as text, not markup. Pinned in Task 2.

---

### Task 1: Form text rules

**Files:**
- Create: `portal/src/graphql/portal/authflows/edit/formText.ts`
- Test: `portal/src/graphql/portal/authflows/edit/formText.test.ts`

**Interfaces:**
- Produces:
  - `type FormTextPart = "title" | "description" | { label: string } | { placeholder: string }` (pointer inside)
  - `formTextKey(formName: string, part: FormTextPart): string`
  - `PLACEHOLDER_FIELD_TYPES: readonly string[]`
  - `usedFormNames(raw: PortalAPIAppConfig): Set<string>`: every `form.name` in every flow type, steps, nested steps and identify options.
  - `hasFormKeys(json: Record<string, unknown>, name: string): boolean`
  - `nextFormName(used: Set<string>, files: Record<string, Record<string, unknown>>): string`: lowest `form_N`, N ≥ 1, not in `used` and with no keys in any file.
  - `formNameFromTitle(title: string, used: Set<string>): string | null`: lowercase, non-`[a-z0-9_]` runs → `_`, trimmed of `_`, prefixed `_` if it starts with a digit; null if empty or `default`; `_2`, `_3`… appended when taken.
  - `isEditableText(value: string, allowLinks: boolean): boolean`: false when the unescaped value has a MessageFormat argument, or HTML other than `<a href>` (any HTML when `!allowLinks`).
  - `type TextEdits = Record<string, Record<string, string | null>>` (language → key → value; null deletes)
  - `type FormNameOp = { kind: "rename"; from: string; to: string } | { kind: "delete"; name: string }`
  - `planTextWrites(latest: Record<string, Record<string, unknown> | null>, edits: TextEdits, ops: FormNameOp[], usedAfter: Set<string>): Record<string, Record<string, unknown>>`: the changed files only, each the latest file with ops then edits applied. Rename copies `from` keys to `to` (the renamed form's keys win), then deletes `from` keys unless `usedAfter` has `from`. Delete removes the name's keys unless `usedAfter` has it. A null latest file starts as `{}`.

- [ ] **Step 0: Branch.** In the worktree, `git switch -c form-text` from `auth-flow-list`.
- [ ] **Step 1: Write the failing tests** in `formText.test.ts`:
  - `formTextKey("about", { label: "/given_name" })` → `"v2.page.fill-form.about.field-label-/given_name"`; title, description, placeholder likewise.
  - `usedFormNames` on a raw config with a fill_form step named `a`, an identify option form `b`, a nested step form `c` in a login flow and a promote flow form `d` → `{a,b,c,d}`.
  - `nextFormName(new Set(["form_1"]), { en: { "v2.page.fill-form.form_2.title": "x" } })` → `"form_3"`.
  - `formNameFromTitle("About you", new Set())` → `"about_you"`; `"2FA setup"` → `"_2fa_setup"`; `"About you"` with `about_you` used → `"about_you_2"`; `"关于你"` → `null`; `"Default"` → `null`.
  - `isEditableText("Hello", false)` true; `"{FlowType, select, other {x}}"` false; `'See <a href="https://x">T</a>'` true with links, false without; `"<b>x</b>"` false; `"'{'literal'}'"` (escaped braces) true.
  - `planTextWrites`:
    - edit only `en.k1`; latest `en` has `k2` → `{ en: { k1, k2 } }`, and no other language in the result.
    - null value deletes the key.
    - rename `a`→`b`, `a` unused after → `a` keys moved; `a` still used → copied, `a` kept.
    - delete `a` with `a` used by another form → file unchanged and not in the result.
    - latest `ja` is null and an edit targets `ja` → `{ ja: { … } }`.
- [ ] **Step 2: Run** `npx jest src/graphql/portal/authflows/edit/formText.test.ts`. Expected: FAIL, module not found.
- [ ] **Step 3: Implement** the exports above in `formText.ts`. MessageFormat detection: after ICU apostrophe quoting is removed (`''` is a literal quote; `'{…}'` is quoted text), any remaining `{` means an argument.
- [ ] **Step 4: Run** the same command. Expected: PASS.

### Task 2: Checkbox labels with links

**Files:**
- Create: `portal/src/graphql/portal/authflows/edit/linkLabel.ts`
- Test: `portal/src/graphql/portal/authflows/edit/linkLabel.test.ts`

**Interfaces:**
- Produces:
  - `type LabelSegment = { text: string } | { text: string; href: string }`
  - `parseLinkLabel(html: string): LabelSegment[] | null`: null when it contains anything but text and `<a href>`.
  - `serializeLinkLabel(segments: LabelSegment[]): string`: text HTML-escaped, links as `<a href="…">…</a>`, adjacent text merged, empty segments dropped.
  - `isAllowedLinkURL(url: string): boolean`: absolute `http:` or `https:` only.

- [ ] **Step 1: Write the failing tests**: round-trip `'I agree to the <a href="https://e.com/t">Terms</a>.'`; `parseLinkLabel('<a href="x" onclick="y">T</a>')` → null (extra attribute); `<b>` → null; `serializeLinkLabel([{ text: "a < b" }])` → `"a &lt; b"`; `isAllowedLinkURL("javascript:alert(1)")` false, `"/terms"` false, `"https://e.com"` true.
- [ ] **Step 2: Run** `npx jest src/graphql/portal/authflows/edit/linkLabel.test.ts`. Expected: FAIL.
- [ ] **Step 3: Implement** with `DOMParser` (jsdom in Jest), walking only text and `A` nodes.
- [ ] **Step 4: Run** it. Expected: PASS.

### Task 3: Loading and saving form text

**Files:**
- Create: `portal/src/graphql/portal/authflows/edit/useFormText.ts`
- Create: `portal/src/graphql/portal/authflows/edit/FormTextContext.tsx`
- Modify: `portal/src/graphql/portal/authflows/edit/draftStorage.ts` (+ its test)

**Interfaces:**
- Consumes: Task 1 exports.
- Produces:
  - `useFormText(appID: string, languages: string[]): FormText` where
    ```ts
    interface FormText {
      isLoading: boolean; loadError: unknown; updateError: unknown;
      // The project's own value; undefined when unset. Pending edits applied.
      ownText(language: string, key: string): string | undefined;
      // What the end-user sees now: own value, else the merged built-in file's.
      effectiveText(language: string, key: string): string | undefined;
      setText(language: string, key: string, value: string | null): void;
      recordNameOp(op: FormNameOp): void;
      hasText(language: string, formName: string): boolean; // for the dot
      edits: TextEdits; ops: FormNameOp[];
      restore(edits: TextEdits, ops: FormNameOp[]): void;
      getIsDirty(): boolean;
      reset(): void;
      // Refetches the latest files, plans writes, sends one mutation.
      save(usedAfter: Set<string>): Promise<void>;
    }
    ```
  - `FormTextContext` value: `{ text: FormText; language: string; setLanguage(l: string): void; defaultLanguage: string }`, plus `useFormTextContext()`.
  - `StoredDraft` gains `text?: { edits: TextEdits; ops: FormNameOp[] }`; `storeDraft(appID, type, name, flow, now, text?)`.
- Values pass through `escapeMessageFormatText` in `setText` and `unescapeMessageFormatText` in the readers. `effectiveText` reads `effectiveData`, `ownText` reads `nullableValue`, of `RESOURCE_TRANSLATION_JSON` per language.

- [ ] **Step 1: Extend `draftStorage.test.ts`**: a draft stored with `text` reads back with it; one stored without reads back with `text` undefined; a stored `text` that is not an object reads back undefined.
- [ ] **Step 2: Run** `npx jest src/graphql/portal/authflows/edit/draftStorage.test.ts`. Expected: FAIL.
- [ ] **Step 3: Implement** the draft change, then `useFormText` and `FormTextContext`. `save` refetches with `fetchPolicy: "network-only"`, calls `planTextWrites`, sends the changed files in one `updateAppTemplates` call, and clears edits and ops on success.
- [ ] **Step 4: Run** the draft test and `npm run typecheck`. Expected: PASS, no errors.

### Task 4: Form names

**Files:**
- Modify: `portal/src/graphql/portal/authflows/edit/rules.ts` (`newStep`), `edit/useFlowEditorValue.ts:143`, `edit/StepForm.tsx` (`OptionForms` near line 1185, `FormNameField`)
- Test: `portal/src/graphql/portal/authflows/edit/rules.test.ts`

**Interfaces:**
- Consumes: `nextFormName`, `usedFormNames`, `formNameFromTitle`, `FormText.recordNameOp`.
- Produces: `newStep(flowType, type, targetStep, formPointer, formName: string | null = null)`; a fill_form step gets `form.name = formName` when given.

- [ ] **Step 1: Write the failing test** in `rules.test.ts`: `newStep("signup", "fill_form", null, "/given_name", "form_1").form` equals `{ name: "form_1", fields: [{ user_profile: { pointer: "/given_name" } }] }`.
- [ ] **Step 2: Run** `npx jest src/graphql/portal/authflows/edit/rules.test.ts`. Expected: FAIL.
- [ ] **Step 3: Implement.** `useFlowEditorValue` passes `nextFormName(usedFormNames(raw with drafts), files)`; `OptionForms` sets the same on a new option form. `FormNameField` commit records `{ kind: "rename", from, to }` when the old name was non-empty. Removing a fill_form step, or an option form, records `{ kind: "delete", name }`. Renaming onto a used name or a name with keys shows the hint `AuthFlowsScreen.edit.form.form-name.shared-text`.
- [ ] **Step 4: Run** the test and typecheck. Expected: PASS.

### Task 5: Form panel

**Files:**
- Modify: `portal/src/graphql/portal/authflows/edit/StepForm.tsx` (`FormEditor`, `FormFieldsField`), `StepForm.module.css`
- Create: `portal/src/graphql/portal/authflows/edit/FormTextFields.tsx`

**Interfaces:**
- Consumes: `useFormTextContext`, `formTextKey`, `isEditableText`, `formNameFromTitle`.
- Produces: `FormEditor` prop `onSelectField?: (index: number) => void` and `selectedField?: number | null`; `<FormLanguageSelect />`; `<FormTextFields formName flowType onNameNeeded />`.

Only for a fill_form step (`option` undefined). In order: Language select (Radix `Select`, project languages by `Locales.{tag}`, default first, a dot on languages where `hasText`), Form name, Title, Description, Hide description checkbox, Fields.

- Empty Title/Description show `effectiveText` as placeholder; the built-in description is formatted with `FlowType` = the flow's type, using the MessageFormat library `src/intl.tsx` already uses.
- A value failing `isEditableText` renders as read-only text with a link to Custom Text (`/project/{appID}/branding/custom-text`).
- Typing into a form without a name first sets the name from `formNameFromTitle` (or `nextFormName`), via the same update as `FormNameField`.
- Field rows: label in the selected language (`ownText` label key, else `formFieldLabel`), a Required badge, existing move/remove buttons; the row is a button calling `onSelectField`. The inline attribute `Select` and Required checkbox are removed from rows; Add field opens a `Select` of addable attributes, with today's unavailable reasons.

- [ ] **Step 1: Implement** the above.
- [ ] **Step 2: Run** `npm run typecheck` and `npx eslint src/graphql/portal/authflows/edit/StepForm.tsx src/graphql/portal/authflows/edit/FormTextFields.tsx`. Expected: no errors.

### Task 6: Field panel and canvas selection

**Files:**
- Create: `portal/src/graphql/portal/authflows/FieldDetailsPanel.tsx`, `FieldDetailsPanel.module.css`
- Create: `portal/src/graphql/portal/authflows/edit/LinkLabelEditor.tsx`, `LinkLabelEditor.module.css`
- Modify: `portal/src/graphql/AuthFlowsScreen.tsx` (selection state near line 316, panel render near line 970), `authflows/sketches.tsx` (`FormFields`, line 153), `authflows/StepDetailsPanel.tsx`

**Interfaces:**
- Consumes: Tasks 1–3, `FormEditor.onSelectField`.
- Produces: selection `{ flowKey; stepID; notice?; field?: number }`; `FormFields` prop `onSelectField?: (i: number) => void` and `selectedField?: number`; `<FieldDetailsPanel step flowType field index onClose readOnly />`.

- Field panel rows per the design table: Label (plain `TextField`, or `LinkLabelEditor` for `boolean`), Placeholder (only `PLACEHOLDER_FIELD_TYPES`), Required/Optional (`SegmentedControl`, writes `user_profile.required`), Attribute (read-only, links to the standard/custom attributes page), Type (read-only words, new IDs `AuthFlowsScreen.field.type.{type}`).
- `LinkLabelEditor`: contentEditable limited to text and links, a Link button opening a popover with a URL field, Apply disabled unless `isAllowedLinkURL`; a link click shows Edit/Remove; output through `serializeLinkLabel`; a checkbox preview below.
- The field panel renders beside the step panel only when the selected step is fill_form; selecting another step clears `field`. Canvas field clicks select the step and field; canvas labels use the selected language.
- Read mode: both panels, every control disabled.

- [ ] **Step 1: Implement** the above.
- [ ] **Step 2: Run** typecheck, eslint and stylelint on the touched files. Expected: no errors.

### Task 7: Save, drafts, leave warning

**Files:**
- Modify: `portal/src/graphql/AuthFlowsScreen.tsx` (lines ~580–700)

**Interfaces:**
- Consumes: `useFormText`, `FormTextContext`, `usedFormNames`, `storeDraft(..., text)`.

- Wrap the editor in `FormTextContext.Provider`; languages from `effectiveAppConfig.localization.supported_languages`, default `fallback_language`.
- Dirty = `form.getIsDirty() || text.getIsDirty()` everywhere it is used (`useLeavePageWarning`, Save button, `getIsDirtyRef`).
- `onSaveDrafts`: `await save(ignoreConflict)` only if the config is dirty, then `await text.save(usedFormNames(latest raw))`; a text failure shows in the error bar and keeps edits pending. `forgetStored` only when both are clean.
- The stored draft includes `{ edits, ops }`; Restore calls `text.restore`; Discard calls `text.reset()`.

- [ ] **Step 1: Implement** the above.
- [ ] **Step 2: Run** `npm run typecheck` and `npx jest src/graphql/portal/authflows`. Expected: no errors, all pass.

### Task 8: Copy

**Files:**
- Modify: `portal/src/locale-data/en.json`

Add every new ID used in Tasks 4–7; remove `AuthFlowsScreen.edit.form.form-name.text-keys` (its hint is replaced) and any ID no longer referenced (grep `portal/src` for each). Copy: "Language", "Title", "Description", "Hide description", "Label", "Placeholder", "e.g. Jane" as the placeholder box's hint, "Required", "Optional", "Attribute", "Type", "Link", "Apply", "Remove link", "This text uses features the builder can't edit. Change it in Custom Text.", "This name already has text. The form will share it."

- [ ] **Step 1: Edit** `en.json`.
- [ ] **Step 2: Run** `npm run typecheck` and grep each removed ID across `portal/src`. Expected: no references.

### Task 9: Full checks, review, commit

- [ ] **Step 1: Run** in `portal/`: `npm run typecheck && npm run eslint && npm run stylelint && npm run prettier && npm run test`. Expected: all pass.
- [ ] **Step 2: Ask fung** to look at the UI; apply changes until approved.
- [ ] **Step 3: Run the `review-pr` skill**; fix Bugs/Security/Performance/Code quality findings.
- [ ] **Step 4: On fung's "commit"**: commit on `form-text` with `[Portal] Edit form text and field settings in the auth flow builder` and `ref DEV-3918`, push to `fung`, open a draft PR based on `auth-flow-list` (stacked on #6008).
