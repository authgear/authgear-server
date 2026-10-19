# Auth flow builder: editing steps and saving

Linear: DEV-3934 (edit steps and branches), with the basic publish of DEV-3936 and the error-on-node part of DEV-3935. Builds on the [read-only viewer](2026-10-14-01-viewer-design.md).

## Goal

Turn the viewer into an editor: an admin opens a flow, changes it on the canvas, and saves it to the project config, then tests it with a real sign-in.

Success: every setting of every step can be changed (through the form or YAML), each edit keeps the draft saveable, and saving a flow loaded from YAML without changes writes the same config back.

## Scope

In:

- Edit mode: draft, Discard, Save through `updateApp`, unsaved-changes blocker, conflict dialog
- Turning a generated flow into the custom `default` flow on first edit
- Adding steps on canvas lines, adding and removing options, removing steps
- A side-panel form for common settings, and "Edit as YAML" for one step
- Keeping `target_step` references valid on rename and remove
- Server validation errors on the node they belong to

Out:

- Drag to reorder or move steps, duplicating, undo
- Deleting a custom flow to fall back to the generated one (DEV-3933)
- Creating, renaming or deleting named flows (DEV-3937)
- Warnings beyond those listed here, and blocking schema-valid but rejected combinations up front (DEV-3935)
- Preview (DEV-3922), export and import (DEV-3920)

## Edit mode

The toolbar has **Edit** for the flow shown.

| Flow shown | Edit |
|---|---|
| Custom | The draft is a copy of the flow from the raw config, not the effective config. |
| Generated | A dialog says the flow becomes a custom flow named `default`, and that settings pages (Login Methods and others) stop changing it until the custom flow is deleted. Confirming copies the generated flow into the draft. Nothing is written until Save. |

While editing:

- The toolbar reads "Editing &lt;name&gt; &lt;type&gt; flow" with **Discard** and **Save**.
- The flow type and name pickers are disabled while the draft has changes. Leaving the page asks first, with the portal's navigation blocker.
- Save replaces only this flow, matched by flow type and name, in `authentication_flow`. A flow from a generated one is appended. Every other flow and key is left as it was.
- A stale config shows the portal's conflict dialog.
- Save succeeds: the page leaves edit mode and shows the saved flow, badged Custom.
- Save fails: the page stays in edit mode. See [Errors](#errors).

## Adding steps

Every line on the canvas is one insertion point. A "+" appears on a line on hover and on keyboard focus.

| Line | Inserts at |
|---|---|
| Start → first step | The top of the flow |
| Step → next step | Between them |
| Option → its first nested step | The start of that option's steps |
| Option → the next outer step (no nested steps) | That option's steps |
| A branch's last step → the next outer step | The end of that branch |
| Last step → End | The end of the flow |

Where a choice's options meet, just before the step they lead into (or End), a second "+" appears while any of those lines is hovered: **Add a step for every option**. It inserts right after the choice, in the choice's own list, so every option runs the step after its own steps. This is how a flow shares steps across options instead of repeating them per option.

Lines into a hand-off node have no "+": `signup_login` options cannot have steps.

The "+" opens a list of the step types the flow type allows:

| Flow type | Step types |
|---|---|
| `signup`, `promote` | `identify`, `create_authenticator`, `verify`, `fill_form`, `view_recovery_code`, `prompt_create_passkey` |
| `login` | `identify`, `authenticate`, `check_account_status`, `terminate_other_sessions`, `change_password`, `prompt_create_passkey`, `fill_form` |
| `signup_login` | `identify` |
| `reauth` | `identify`, `authenticate` |
| `account_recovery` | `identify`, `select_destination`, `verify_account_recovery_code`, `reset_password` |

`fill_in_user_profile` is deprecated and is not offered; existing ones stay editable through YAML.

A new step is valid for the schema as inserted, and becomes the selected node:

| Step type | Starts with |
|---|---|
| `identify` | One option: the first identification the flow type allows. In `signup_login`, `login_flow: default`, plus `signup_flow: default` except for `select_account` and `passkey`. In `account_recovery`, a nested `select_destination` step, since those options need at least one step. |
| `authenticate`, `create_authenticator` | One option: the first authentication the flow type allows |
| `verify` | `target_step` naming the nearest earlier `identify` step |
| `change_password` | `target_step` naming the nearest earlier `authenticate` step |
| `fill_form` | A form with one optional field, `/given_name` |
| Others | No other fields |

When the step a new `target_step` names has no `name`, it is given a unique one, and the panel says so. When there is no earlier step of that type, the type is greyed out in the list, with the reason as its tooltip.

## Options

The side panel of a choice step (`identify`, `authenticate`, `create_authenticator`) has **Add option** under its options; the canvas cards have no edit controls besides the toolbar. The new option takes the first identification or authentication the flow type allows that the step does not already list.

`target_step` on an option is an advanced setting: only options that send codes read it (it names the earlier step whose email or phone the codes go to). It is not a column of the options table; a collapsed **Advanced** section under the table, shown only when some option sends codes, has a **Send codes to** select per code option, defaulting to "The email or phone the user enters". When closed, the section says how many are set. Switching an option to an authenticator that sends no codes removes its `target_step` and `verification_required`.

An option's authenticator is picked from two groups, **Primary** and **2FA**, since both have a password and codes sent to email or phone; a 2FA authenticator is tagged "2FA" wherever it shows.

`id_token` is not offered in `login` and `signup_login`: apps practically never send an ID token there. An existing `id_token` option still shows its value. `reauth` keeps it, as its only identification.

## Removing

The selected node's toolbar has **Remove**. Each option in the side panel has its own remove control.

- Removing a step removes every step nested under its options. When there are any, a confirmation says how many.
- Blocked, with the reason:
  - the flow's only step
  - a choice's only option
  - a step that a `target_step` names, or one with such a step nested under it, unless every step that names it goes too

## Lanes

The editor draws flows in lanes; the viewer's free lines are gone. A choice where any option has steps of its own is drawn as a box: the choice card on the left, then one row per option, labelled with the option, holding that option's steps. The rows' lines meet at a dot on the box's right edge, labelled "Every option", and one line leaves it: what follows the box runs for every option. A choice whose options have no steps stays a single card. In `account_recovery`, rows do not meet at the dot, so each still leads to its own section of Verify recovery code.

- A line that leaves an option has its "+" just past that option's output. Every "+" names what it adds to, as a tooltip and as its menu's heading: "Add a step for Email only", "Add a step for every option", or "Add a step here" at the top level.
- A "+" inside a row adds a step to that option only; the "+" on the line leaving the dot adds a step for every option.
- Dragging a card onto a "+" moves the step there, with any steps nested under its options; holding Alt copies it instead, without its name. While dragging, every "+" shows, and the one the card would land on is highlighted.
- Blocked, with a notice over the canvas saying why:
  - into the step itself
  - out of an option to after its choice, while another option of that choice has its own copy: a step of the same type that, for a choice, shares an identification or authenticator with it (a password step and a 2FA step are not copies); that option would run it twice
  - leaving an `account_recovery` option without steps
  - into an option in `signup_login`
- Dropping a card where it already is changes nothing.

## Side panel

In edit mode, the panel shows the step's description, its form, then "Edit as YAML". Labels use the canvas wording; hovering a label shows the config key it writes. Edits apply to the draft as they are made. A control writes only the field it owns, so fields the form does not show are kept.

Options are a table: one row per option, one column per setting the step type uses, with the column labels shown once ("Has account", "No account", "No account found", "Target step", "Verify"). A setting that does not apply to an option shows a dash. Each row ends with a remove icon.

Every step has **Name** (`name`), checked against `^[a-zA-Z_][a-zA-Z0-9_]*$`. A name another step in the flow already has is a warning, not an error: the runtime uses the later step, so `target_step` would resolve to it.

| Step type | Form |
|---|---|
| `identify` | Options. Each: identification (a select of those the flow type allows and the step does not already list). `signup_login` adds the login flow and signup flow it hands off to (selects of the flow names of that type); `select_account` and `passkey` always find an account, so they show only "Continues in login flow" and carry no `signup_flow`. `account_recovery` adds `on_failure`: "Show an error" or "Continue as if it was sent". |
| `authenticate` | Optional (login only). Options. Each: authentication, `target_step`. |
| `create_authenticator` | Options. Each: authentication, `target_step`, `verification_required` (email and phone codes only). |
| `verify`, `change_password` | `target_step` |
| `fill_form` | Form name, and fields. Each field: an attribute (standard or custom), Required. |
| `select_destination` | "Let the user choose where to send it" (`enumerate_destinations`), stating that anyone who knows a login ID then sees masked versions of the account's other emails and phones. A checkbox per allowed channel and form (`allowed_channels`). |
| Others | Name only |

The form has no bot protection control for now: it confused more than it helped. `bot_protection` stays as written, and "Edit as YAML" can change it.

A `target_step` select lists only steps the runtime accepts, named and earlier on the path to this step:

| Used by | Targets |
|---|---|
| `verify`, a `create_authenticator` option | an `identify` step |
| an `authenticate` option | an `identify` step whose options are not all `id_token` |
| `change_password` | an `authenticate` step |

When no candidate is named, the select offers to name one.

### Edit as YAML

The portal's `CodeEditor` shows the selected step as YAML, with everything nested under it. **Apply** replaces the step in the draft. It is refused, with the reason under the editor, when the YAML does not parse, is not a single step object, or has a `type` the flow type does not allow. It covers what the form does not: an `identify` option's `form` and `account_linking`, `fill_in_user_profile`, and any field added later.

## target_step references

`target_step` exists only in `signup`, `promote` and `login` flows, and names a step in the same flow. The server checks on save that the name exists; the runtime checks the step's type and that it ran earlier.

- **Rename:** every `target_step` naming the old name follows. When another step has the same old name, references are left alone and the panel says why.
- **Clearing a targeted step's name:** refused in the name field.
- **Removing:** see [Removing](#removing). The message lists the steps that name it, each selecting that step.
- **Warnings that do not block Save,** shown on the node with the reason in the panel:
  - a `target_step` naming no step (the save will be rejected)
  - a target of the wrong type, not earlier on the path, or an `identify` whose options are all `id_token` for an `authenticate` option

## Errors

- **Save:** an error whose pointer lies inside the edited flow (`/authentication_flow/<type>_flows/<index>/steps/...`) marks the node whose pointer is its longest prefix, with a red outline and the message at the top of its form. Every error also appears in the form's error bar.
- **Edit as YAML:** see [above](#edit-as-yaml).
- Nothing else can break the config: an edit either keeps the draft valid for the schema or is blocked.

## Testing

Unit tests for the pure edit functions on the flow config:

- Insert at each line kind, into the right config location
- Remove a step or an option, including each blocked case
- New-step defaults per step type and flow type
- Rename with `target_step` rewriting, including two steps with the same name
- Reference checks: which removals are blocked, and each warning
- The pointer to select after each edit
- Round trip: each fixture flow, and a hand-written flow with fields the portal does not know, saved without changes, gives a config equal to the input
- Save: only the edited flow is replaced or appended
- Error pointers map to the right node, or to none

Checks before the PR: typecheck, eslint, stylelint, prettier, jest, build, and `review-pr`.
