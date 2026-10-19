# Auth flows list and New flow

## Goal

"Auth Flows" lists the project's flows: the flows generated from Login methods and the custom ones. From it, a developer opens, edits, creates, renames, duplicates and deletes them. A generated flow is edited directly; saving it makes it custom. Creating a flow starts from scratch, from a template, or (later) from an import, for one of three types: Signup & login, Signup, or Login.

Builds on the editor (`2026-10-15-01-editor-design.md`). Tracked in DEV-3944, which took over most of DEV-3937.

## Scope

- In: the list page, the New flow dialog, the editor page at its own URL, the list actions.
- Out:
  - Promote, reauth and account recovery flows: not listed or created yet.
  - Import. It shows in the dialog as "Coming soon" (DEV-3920).
  - Generating flows from a prompt (DEV-3943).
  - Assigning flows to applications (DEV-3916).

## Pages

| URL | Page |
|---|---|
| `…/configuration/authentication/flows` | The list |
| `…/configuration/authentication/flows/<type>/<name>` | The editor for one flow |

## Signup & login flows

Every flow stands on its own: a signup_login flow is its identify step, whose options hand off to signup and login flows by name (`login_flow`, `signup_flow`). Those flows are listed, edited, renamed and deleted separately. A new signup_login flow hands off to the `default` signup and login flows; its options' "Has account" / "No account" settings repoint them.

## The list

Sections in this order: **Signup & login**, **Signup**, **Login**. Promote, reauth and account recovery flows are not listed for now; they stay in `authgear.yaml` untouched. A section with no flows is left out. Each section starts with its `default` flow: the custom one, or the generated one when there is none.

Each row: the flow's name, linking to the editor, and a **⋯** menu. Under the name, when anything names the flow, a grey line says what: "Used by group shared_2fa · app Admin portal · Signup & login flow members" (flow groups, applications' allowed flows, signup_login hand-offs). Nothing is shown for a flow nothing names, or for `default`. The only tag is **Edited**, on a `default` flow changed from the one built from Login methods (its tooltip points to Reset). A generated row's menu has Open and Duplicate only, since it is not in `authgear.yaml` until edited and saved.

### Actions

Each action saves to `authgear.yaml` at once, built on the last saved config, after a confirmation where noted.

| Action | Behaviour |
|---|---|
| Open | Goes to the editor. |
| Rename | Not offered for the `default` flow. Asks for a name. References follow: `login_flow` / `signup_flow` hand-offs, flow groups (`ui.authentication_flow.groups`) and applications' allowed flows (`x_authentication_flow_allowlist`). |
| Duplicate | Asks for a name; copies the flow. A copied signup_login flow keeps its hand-offs. |
| Delete | Confirmation. Refused, with the reason, while a signup_login flow hands off to it: there is no safe stand-in. Flow groups and applications are cleaned up instead, and the confirmation says how: a group falls back to the `default` flow of that type, and an application no longer lists the flow as allowed. |
| Reset | Offered instead of Delete for an edited `default` flow. Confirmation: the flow goes back to the one built from Login methods and follows them again; edits are lost. It removes the custom `default`, so the generated one runs; hand-offs to `default` keep working, so it is never refused. |

A name is valid when it matches `^[a-zA-Z_][a-zA-Z0-9_]*$`, is not `default`, and no flow of that type has it. `default` is reserved: it is the generated flow, or the custom flow saved from editing it.

## New flow

A dialog from the list's **New flow** button. Nothing is saved until the editor's Save.

1. **Type.** Cards: Signup & login ("One page: existing users log in, new users sign up"), Signup ("Create an account"), Login ("Sign in to an existing account").
2. **Start.**
   - **From the default flow** (first, selected): a copy of the flow generated from Login methods. A copied signup_login flow still hands off to the `default` signup and login flows.
   - **From scratch:** the smallest flow that saves. Signup: an identify step with the project's first login method. Login: identify, then authenticate. Signup & login: an identify step whose option hands off to the `default` signup and login flows.
   - **Use a template:** step 3.
   - **Import:** disabled, "Coming soon".
3. **Template.** The starters: Email + password, Passwordless email code, Phone + SMS code, Email + password + 2FA. Each card shows its steps in one line ("Identify → Verify → Create password"). A starter needing a login method or authenticator the project has not enabled is disabled, with the reason. One link under the templates goes to Login methods.
4. **Name.** Prefilled with an unused name: the type plus a number (`signup_login_2`, `login_2`). `default` is refused.

**Create** opens the editor on the new flow as an unsaved draft. Discard on a never-saved flow returns to the list. Reloading before the first save loses the draft and shows "Flow not found" with a link to the list.

## The editor page

- Header: a back link to the list, the flow's name and type, and Edit / Save / Discard. No badges, and no type or name pickers.
- A generated `default` opens like any flow, and Edit starts at once: the page never says generated or custom. An edited `default` has **Reset** next to Edit, as in the list. A URL naming a flow that is neither custom nor a generated `default` shows "Flow not found".
- A hand-off card's Open goes to that flow's editor.

## Errors

A failed save shows the existing error bar and leaves the list as it was. Server validation decides: a template that the client allows but the server rejects shows the server's message.

## Testing

- Rows: a generated default only for a type with no custom default, every flow on its own row, empty sections left out.
- Actions: rename with hand-off references following a renamed signup or login flow, add, delete, the delete refusal for a flow a signup_login flow hands off to, name validation within a type.
- Starters: availability for a project missing a login method or authenticator; every starter uses only step types and options the editor allows.
