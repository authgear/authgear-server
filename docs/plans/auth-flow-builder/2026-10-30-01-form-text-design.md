# Form text in the auth flow builder

## Goal

A developer sets everything an end-user reads on a form, in every language the project supports, from the form step in the builder: the form's title and description, and each field's label and placeholder. The form's fields also get their own settings panel. Today the text can only be changed by editing `translation.json` by hand, and a field can only be configured inline in a crowded list.

The typical moment: open the form node, pick Chinese, fill in the Chinese text.

Tracked in DEV-3918. Builds on the editor (`2026-10-15-01-editor-design.md`) in PR #6008. Needs DEV-3989 (#6029, per-form labels) and DEV-3990 (placeholders and the Authflow API), which ship first as separate server PRs.

## Scope

- In: the form panel's text and language dropdown, the field panel, link editing in checkbox labels, a name for every new form, saving text with the flow.
- Out:
  - Editing an attribute's own label, which shows on the settings page and in every form.
  - Pattern error messages: edited on the attribute (DEV-3913, #6026).
  - Custom Text: stays a raw JSON editor.
  - Text on a form attached to an identify option. Its fields stay edited inline in the identify step's panel as today; only its name is prefilled.
  - Any note that the text is for Auth UI only. Custom UIs receive the same text through the Authflow API (DEV-3990), so there is none.

## Server contract (DEV-3989, DEV-3990)

The builder only reads and writes these keys, per language, in `templates/{language}/translation.json`:

| Text | Key | When unset |
|---|---|---|
| Title | `v2.page.fill-form.{name}.title` | `v2.page.fill-form.default.title`, then the built-in title |
| Description | `v2.page.fill-form.{name}.description` | `v2.page.fill-form.default.description`, then the built-in description. An empty value hides it (see below) |
| Field label | `v2.page.fill-form.{name}.field-label-{pointer}` | The attribute's label key, then the label derived from the attribute name |
| Field placeholder | `v2.page.fill-form.{name}.field-placeholder-{pointer}` | No placeholder |

Auth UI shows all four; the Authflow API returns them as `form.title`, `form.description`, `fields[].label` and `fields[].placeholder`. A form attached to an identify option has no title, description or placeholder: its fields sit next to the login ID with labels inside the inputs.

Every key is per form name, so forms sharing a name share their text.

An empty named description must hide the description. Today it does not: the template treats an empty key as unset and shows the `default` description. DEV-3990 fixes this, in Auth UI and in the API (`form.description` absent).

## Form names

- Every new form gets a name: a new form step, and a form added to an identify option. The prefill is `form_1`, `form_2`, … the lowest number no form in the project uses. *(My call; the developer renames it.)*
- An existing form without a name gets one when the developer first types text into it, made from the title in the default language ("About you" → `about_you`), or `form_N` when there is no title. *(My call.)*
- The prefill skips names that still have keys in any language's `translation.json`, such as text added by hand in Custom Text, so a new form never picks up text it was not given. *(My call.)*
- Renaming a form carries its text: on Save, its keys are copied to the new name in every language. The old keys are removed when no other form in `authgear.yaml` keeps the old name. *(My call; without it, renaming silently loses every translation.)*
- Renaming onto a name that already has text, or another form's name, shows a hint under the name: "This name already has text. The form will share it." On Save the renamed form's own keys win where both are set. *(My call.)*
- A name change is the developer's own config edit, visible in the Form name field; hooks and Custom UIs see the new `form.name` after Save, like any rename. No confirmation and no breaking-change entry. *(My call.)*
- Deleting a form deletes its keys in every language on Save, unless another form in `authgear.yaml` still uses the name, including flows the builder does not list.

## Form panel

The step panel (400px) shown for a form step.

- **Language** dropdown at the top: the project's supported languages by name, the default language first and selected when the panel opens. A dot marks languages that have text set for this form. It controls every text box in both panels and the text on the canvas node.
- **Form name**, as today, prefilled for new forms.
- **Title** and **Description**. Empty boxes show the text in effect for this language as grey placeholder text, the built-in description worded for this flow's type. Clearing a box returns to that text.
- **Hide description**, a checkbox under Description, for the selected language. It saves an empty description. *(My call.)*
- A value the panel cannot edit safely, one with MessageFormat arguments such as `{FlowType, select, …}` or, in a checkbox label, markup other than `<a href>`, is shown read-only with a link to Custom Text. The panel never rewrites it unless the developer clears it. *(My call; opening and saving must not change existing text.)*
- **Fields**: one row per field, showing its label in the selected language and a Required badge, with the move up/down and remove controls and the Add field button that exist today. Clicking a row opens the field panel. **Add field** asks which attribute to add, with the same choices and unavailable reasons as today's inline select. A field's attribute is not changed afterwards; the developer removes it and adds another. The inline Required checkbox moves to the field panel. *(My call; the attribute is read-only in the field panel.)*

## Field panel

A second panel that slides out beside the form panel when a field is selected, from its row or from the field on the canvas node. The selected field is highlighted in both. Selecting another field swaps the panel's content; ✕ closes it.

| Setting | Shown for | Notes |
|---|---|---|
| Label | Every field | In the selected language. Empty shows the attribute's label in that language |
| Placeholder | `string`, `email`, `url`, `integer`, `number`, `phone_number` fields | Example answer, e.g. "e.g. Jane" |
| Required / Optional | Every field | Writes `user_profile.required` |
| Attribute | Every field | Read-only: the pointer and whether it is standard or custom, linking to its settings page |
| Type | Every field | Read-only: the field's type in words (Text, Email, Checkbox, Dropdown, Date, …) |

### Checkbox labels with links

A checkbox (`boolean`) field's Label is a rich-text box. Select words, click **Link**, paste a URL, Apply. Clicking an existing link edits or removes it. The developer never sees HTML. A preview of the checkbox with its label sits below.

- What it saves is plain text with `<a href="…">…</a>` and nothing else, the only markup the server keeps (it adds `target` and `rel` itself).
- Only `http` and `https` URLs are accepted.
- Other field types' labels are plain text.

## Saving

- Text edits are part of the editor's draft, with the flow's other edits, and survive a reload the same way.
- **Save** writes `authgear.yaml` first, then the changed `translation.json` files in one call. Each file is the latest saved file with only the edited keys changed, so text changed elsewhere since the editor loaded is kept. If that call fails, the flow is saved, the error shows, and the text edits stay pending for the next Save. *(My call.)*
- Text is stored escaped for MessageFormat, as Custom Text expects, and shown unescaped.
- Read mode shows the same panels with every control read-only.

## Left to the builder

- How the field panel slides in, its width, and how the canvas makes room for two panels.
- The rich-text editing component.
- How the dot per language is computed from the loaded translation files.
