# Auth flows list and New flow

## Goal

"Auth flows" opens on a list of the project's flows. From it, a developer creates, opens, renames, duplicates and deletes flows, and chooses which one is in use. Creating a flow starts from scratch, from a template, or (later) from an import, for one of three types: Signup & login, Signup, or Login.

Builds on the editor (`2026-10-15-01-editor-design.md`). Tracked in DEV-3944, which took over most of DEV-3937.

## Scope

- In: the list page, the New flow dialog, the editor page at its own URL, Signup & login bundles, the list actions.
- Out:
  - Creating promote, reauth and account recovery flows. Existing ones are listed and can be opened and edited.
  - Import. It shows in the dialog as "Coming soon" (DEV-3920).
  - Generating flows from a prompt (DEV-3943).
  - Assigning flows to applications (DEV-3916).

## Pages

| URL | Page |
|---|---|
| `…/configuration/authentication/flows` | The list |
| `…/configuration/authentication/flows/<type>/<name>` | The editor for one flow, or one bundle |

## Bundles

A Signup & login flow is a bundle: one `signup_login_flows` entry and the `signup_flows` and `login_flows` entries it hands off to, all sharing one name. Example: the bundle `members` is the signup_login flow `members`, whose options name `login_flow: members` and `signup_flow: members`.

- The list shows a bundle as one row, and the editor opens it with three tabs.
- A bundle's actions (rename, duplicate, delete, set as default) apply to all three flows.
- A signup_login flow whose options name flows of other names is listed on its own, not as a bundle.

## The list

Sections in this order: **Signup & login**, **Signup**, **Login**, then **Other flows** (promote, reauth, account recovery) when the project has any.

Each row:

- The flow's name.
- **In use** when it is named `default`, otherwise **Not in use**. Only the `default` flow of each type runs in AuthUI until DEV-3916 lets applications pick others.
- **Generated** when it comes from Login methods (no custom `default` of that type).
- A bundle row adds "Sign up: `<name>` · Log in: `<name>`". Its signup and login flows also appear in their own sections, tagged "Part of `<name>`", since AuthUI's separate sign-up and log-in pages run them on their own.
- A **⋯** menu with the actions below.

When a type has no custom `default`, its section lists the generated flow as "default · Generated · In use". Its menu has **Open** (read-only) and **Duplicate**; duplicating is how a developer starts editing it.

### Actions

Each action saves to `authgear.yaml` at once, built on the last saved config, after a confirmation where noted.

| Action | Behaviour |
|---|---|
| Open | Goes to the editor. |
| Set as default | Swaps names with the current custom `default` of that type (for a bundle, of all three types) and rewrites every `login_flow` / `signup_flow` that named either. With no custom `default`, the flow becomes `default` and the generated flow stops running. Confirmation names both flows. |
| Rename | Asks for a name. `login_flow` / `signup_flow` references follow. Renaming to `default` is refused: use Set as default. |
| Duplicate | Asks for a name; copies the flow, or all three for a bundle. The copy is not in use. |
| Delete | Confirmation. Deleting the in-use `default` warns that the generated flow takes over. Refused, with the reason, for a signup or login flow that a signup_login flow still names. |

A name is valid when it matches `^[a-zA-Z_][a-zA-Z0-9_]*$` and no flow of that type has it; for a bundle, no flow of any of the three types.

## New flow

A dialog from the list's **New flow** button. Nothing is saved until the editor's Save.

1. **Type.** Cards: Signup & login ("One page: existing users log in, new users sign up"), Signup ("Create an account"), Login ("Sign in to an existing account").
2. **Start.**
   - **From scratch:** the smallest flow that saves. Signup: an identify step with the project's first login method. Login: identify, then authenticate. Signup & login: both, and an identify step whose options hand off to them.
   - **Use a template:** step 3.
   - **Import:** disabled, "Coming soon".
3. **Template.**
   - First: "From your Login methods settings", a copy of the generated flow.
   - Then the starters: Email + password, Passwordless email code, Phone + SMS code, Email + password + 2FA. Each card shows its steps in one line ("Identify → Verify → Create password").
   - A starter needing a login method or authenticator the project has not enabled is disabled, with the reason. One link under the templates goes to Login methods.
4. **Name.** Prefilled with an unused name: the type plus a number (`signup_login_2`, `login_2`). `default` is allowed only when the type has no custom `default`, with a note that the flow is in use once saved.

**Create** opens the editor on the new flow as an unsaved draft. Discard on a never-saved flow returns to the list. Reloading before the first save loses the draft and shows "Flow not found" with a link to the list.

## The editor page

- Header: a back link to "Auth flows", the flow's name and type, the In use / Generated badges, and Edit / Save / Discard. No type or name pickers.
- A bundle has three tabs: **Choose** (the signup_login flow), **Sign up**, **Log in**. The tab is in the URL (`?part=signup`, `?part=login`). A hand-off card's Open switches tabs. One draft covers the three flows and Save writes them together.

## Errors

A failed save shows the existing error bar and leaves the list as it was. Server validation decides: a template that the client allows but the server rejects shows the server's message.

## Testing

- Rows: bundles grouped by shared name, a signup_login flow with other names left unbundled, generated rows only without a custom `default`, "Part of" tags.
- Actions: the name swap and its reference rewrites, bundle-wide rename and duplicate, the delete refusal for a named-by-a-bundle flow, name validation.
- Starters: availability for a project missing a login method or authenticator; every starter uses only step types and options the editor allows.
