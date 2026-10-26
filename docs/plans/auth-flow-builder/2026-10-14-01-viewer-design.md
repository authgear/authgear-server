# Auth flow builder: read-only flow viewer

Linear: DEV-3912 (builder), DEV-3930 (canvas rules), DEV-3933 (load real flows). Project "Form filling in auth flow".

## Goal

A portal page that loads a project's auth flows and draws the one the runtime will run, as a left-to-right canvas. This is permanent code behind the flag `portal.auth_flow_builder_enabled`, and the base that editing (DEV-3934), saving (DEV-3936) and preview (DEV-3922) build on.

Success: an admin picks a flow type and sees its effective flow as a canvas of page sketches. The canvas matches the config step for step, with branches in config order.

## Scope

In:

- Route, nav item and flag gating
- Loading custom and generated flows, and resolving which flow applies
- Flow type and name pickers, Custom or Generated badge
- Canvas with auto layout, page sketches, pageless pills, Start and End
- Read-only side panel with step details and YAML
- Links between a `signup_login` option and the flow it hands off to

Out:

- Editing, the add-node toolbar, Render (DEV-3922), Export (DEV-3920), Publish (DEV-3936)
- "Start from my current settings" and deleting a custom flow (DEV-3933, need saving)
- Creating, renaming or deleting named flows (DEV-3937)
- Validation beyond the warnings listed under [Edge cases](#edge-cases) (DEV-3935)
- Real AuthUI renders inside nodes (DEV-3922)

## Dependencies

- authgear-server #5994: the `portal.auth_flow_builder_enabled` flag
- authgear-server #5995: `App.generatedAuthenticationFlows`

Branch from `main` after both merge, or stack on #5995.

## Page

| Item | Value |
|---|---|
| Route | `/project/:appID/configuration/authentication/flows` |
| Screen | `AuthFlowsScreen`, a lazily loaded route |
| Nav | "Auth Flows" under Authentication, after "2FA" |
| Gating | The nav item and route exist only when `effectiveFeatureConfig.portal.auth_flow_builder_enabled` is true. Otherwise the URL redirects to Getting Started, like other feature-gated pages (`RequireAppFeature`). |
| Width | Full width, no centred content column |

## Data

One query, `AuthFlowsScreenQuery`, on the App node:

- `effectiveAppConfig`: custom flows under `authentication_flow`, plus the identity config needed for sketches (OAuth providers, login ID keys, custom attributes)
- `generatedAuthenticationFlows`

Both come from the same config in one request, so they cannot disagree. The page reads no secrets and does not use `useAppAndSecretConfigQuery`.

### Resolving a flow

For flow type `T` and name `N`:

1. If `authentication_flow.<T>_flows` has a flow named `N`, show it with the "Custom" badge.
2. Otherwise, if `N` is `default`, show the generated flow of type `T` with the "Generated from your settings" badge. The badge links to Login Methods.

This matches the runtime (`pkg/lib/authenticationflow/declarative/utils_common.go`). Resolution is per type, so a project can have a custom signup flow and a generated login flow.

## Toolbar

| Control | Behaviour |
|---|---|
| Type picker | Sign up or log in / Sign up / Log in / Promote anonymous user / Reauthenticate / Account recovery |
| Name picker | `default` plus the names of any custom flows of the selected type. Hidden when `default` is the only one. |
| Badge | "Custom" or "Generated from your settings" |

The type and name live in the URL: `?type=<signup_login|signup|login|promote|reauth|account_recovery>&name=<name>`. A missing or unknown value falls back to `signup_login` and `default`.

## Model and layout

There are three pure functions, none of which use React.

### `flowToTree(flow)`

- A flow is a **sequence** of steps.
- A step whose `one_of` has two or more options is a **choice**. Each option is a **branch** holding its nested `steps` as a sequence, which may be empty.
- A step whose `one_of` has exactly one option is not a choice. That option's nested steps are inlined directly after the step.
- A step whose `one_of` is present but empty is a choice with no branches (see [Edge cases](#edge-cases)).
- `optional: true` is kept as a flag on the step.
- `target_step` is kept as step data for the side panel. It is not an edge.
- Every step in the config is drawn as its own node, even when several options repeat the same steps (the generator repeats a chain per identification). The canvas mirrors the config one-to-one, because the builder will edit it: a merged node would hide which copy an edit changes.
- Every step knows the identification of the identify option it follows (Email, Phone), so a `verify` step can say where its code went.
- A `device_token` option is not a branch. The user never chooses it: it lets a remembered device skip the step, and adds "Remember this device" to the step's page. Branches keep their original option index.

### `treeToGraph(tree)`

- One node per step, plus a Start node.
- A choice node has one output handle per option, in config order. Option edges carry no label: each handle sits beside its option's row, which already names it.
- Each handle connects to the first step of its branch. If the branch is empty, it connects to the join.
- **Join:** the last node of every branch connects to the step after the choice, or to End if there is none.
- No step gets a bypass edge, optional or not: its condition line already says it may be skipped.
- **Hand-off:** an option with `login_flow` or `signup_flow` (a `signup_login` identify option) ends the flow. It connects to one flow node per flow it names, instead of continuing. An option naming both has two outputs, "Has account" to the login flow and "No account" to the signup flow, because AuthUI logs the user in when an account with that login ID or identity exists and signs them up otherwise. Options naming the same two flows form a group: they share one "Has account" / "No account" pair, from the group's first option, and are drawn together in one tinted box at that option's position.
- **End** is drawn only when some path reaches it. A `signup_login` flow, whose every option hands off, has none.

### `layoutGraph(tree, sizes, terminals)`

- The layout is recursive and runs left to right, with a 120px gap between columns.
- A sequence places its nodes in a row.
- A choice stacks its branches vertically in config order and centres the choice node on the stack. The join goes to the right of the widest branch.
- The terminals (End, or the flows the tree hands off to) stack in the last column, in the order options first name them.
- **Lanes:** an edge that skips columns gets a reserved horizontal lane, so no line runs under a card. That is the edge of an empty branch beside non-empty ones, given a lane in the branch stack. When every branch of a choice is empty, no lane is needed.
- **Edge routing:** edges are curves. An edge to the next column is one S-curve. An edge that skips columns curves into its lane inside the gap after its source, runs straight along the lane, and curves into its target inside the gap before it, so it never crosses a card.
- `sizes` are the node sizes measured by `@xyflow/react` after the first render. The canvas renders hidden, measures, lays out, shows the result and fits the view.
- Positions are never stored. They are recomputed from the tree on every render.

`@xyflow/react` provides pan, zoom, fit view, a minimap and custom nodes. Dagre and elk are not used, because they reorder nodes to reduce crossings and branch order carries meaning: it is the order of choices on the page.

## Nodes

All nodes use Radix and theme tokens. Every label comes from `en.json`.

### Page node

A card about 240px wide. The header has a title, plus the step `name` in small mono text when one is set. The title is the step's config `type` in plain words, sentence case (`create_authenticator` is "Create authenticator", `prompt_create_passkey` is "Prompt to create passkey"), so the canvas matches the YAML one-to-one. The body is a greyscale wireframe of the page, built from config:

| Step | Sketch |
|---|---|
| `identify` | Per option: a login ID input labelled Email, Phone or Username; one "Continue with X" button per configured OAuth provider; a passkey button; a "Continue as…" row for `select_account`; username and password inputs for LDAP. A form attached to an option shows its fields under that option. |
| `authenticate` | Per option: a password input; OTP boxes ("Code sent to email" or "Code sent to phone"); authenticator-app code boxes; a passkey button; a "Use a recovery code" link. With `device_token`, a "Remember this device" checkbox after the options, with no "or" and no output, noted "Next time, this device skips this step." |
| `create_authenticator` | Per option, a heading describing the method ("Create a password", "Create a passkey", "Get codes by email", "Use an authenticator app"; a phone code names its channel from `authenticator.oob_otp.sms.phone_otp_mode`: "Get codes by SMS", "by WhatsApp" or "by WhatsApp or SMS"), then its fields: a new password input; a passkey button; for a code method, the email or phone input (left out with `target_step`, when the code goes to that step's login ID) and code boxes; for an authenticator app, a QR code and code boxes. |
| `verify` | Code boxes, labelled by the login ID its branch collected: "Enter the code sent to the email", or for a phone, by its channel from `phone_otp_mode` ("…sent by WhatsApp or SMS"). |
| `verify_account_recovery_code` | One section per channel and form the flow's `select_destination` steps can send, headed by what it sends ("Reset link by email"): a link shows "Link sent" and Resend, as AuthUI's link-sent page does; a code shows code boxes. Several are drawn as bordered sub-nodes. Each section has its own input on the card's left edge, and each `select_destination` connects to the sections for what it sends, so the email branch lands on the link and the phone branch on the code. A `select_destination` with no `allowed_channels` counts as every channel, as at runtime. |
| `fill_form`, `fill_in_user_profile` | Each field with its attribute label, a required marker, and a control that matches the attribute type. |
| `change_password`, `reset_password` | New password and confirm inputs. |
| `select_destination` | Only with `enumerate_destinations`, when it is a page: "Choose where to send it", two masked sample destinations as radio rows, and Send. Without it, it is a pill (see [Pageless node](#pageless-node)). |
| `view_recovery_code` | A grid of masked codes. |
| `prompt_create_passkey` | A passkey prompt with "Skip". |
| `terminate_other_sessions` | "Logged in on another device", a Continue button and a "No, don't log out" link, matching AuthUI's confirmation page. |

A step the runtime may skip states when it runs, in a short accent-coloured line with a branch icon at the top of the card or pill, above the title, with a divider under it. Hovering the line shows the full explanation.

| Line | Steps | Full condition |
|---|---|---|
| Only if the user has set up 2FA | `optional: true`, every option a second factor | the user has set up one of the step's 2FA methods |
| Only if the user has set up one of these | any other `optional: true` step | the user has set up one of the step's methods |
| Only if the password must change | `change_password` | the password just used has expired, or fails the policy while `authenticator.password.force_change` is on |
| Only if the app allows one session | `terminate_other_sessions` | the app allows a single session (`max_concurrent_session: 1`) and the user already has one there |
| Only if the user has no passkey | `prompt_create_passkey` | the user signed in with something that can use a passkey and has none yet |


On a choice node, each option's output handle sits beside that option's row in the sketch. When a step has more than one option, each option is drawn as its own sub-node (a bordered, rounded box holding its sketch and its handle), with an "or" between them, since they are alternatives on one page. Inputs and code boxes in a sketch have a solid panel background, so they read as fields.

### Identify group

An `identify` step offering `select_account` alongside other options is drawn as the two pages AuthUI shows, stacked inside one container labelled with the step title and name. It is still one step: clicking anywhere in it opens that step's panel.

| Part | Shows |
|---|---|
| "Continue as …" page (top, dashed) | Caption "If signed in" (tooltip: shown on its own page, before the login page, when the user already has a session); a Continue button whose output leaves the group's right edge. |
| "Use another account" (between the pages) | A small action node, with an arrow in from the page above and out to the page below. |
| Page below (solid) | Caption "Otherwise"; every other option, separated by "or". |

The step is entered between the two pages: a dashed line forks up to the "Continue as" page and a solid one down to the page below.

### Flow node

A flow the current flow hands off to: "Log in flow" or "Sign up flow", its name, and "Open flow →", which switches the canvas to it.

### Pageless node

A small pill. It is used for `check_account_status`, `select_destination` without `enumerate_destinations`, and an `identify` step whose only option is `id_token`. An `identify` by `id_token` adds a muted line, "From the app's ID token", whose tooltip explains that the app sends the signed-in user's ID token and the flow identifies them from it, with no page. A `select_destination` pill adds what it sends, one entry per allowed channel ("Reset link by email · Code by SMS"), and a line saying where: "To the email entered", "To the phone number entered", or "To the email or phone number entered" outside an identify option. Its tooltip says the user does not see the step.

### Start and End

Small circles.

### Unknown node

A plain card showing the raw `type`, for step types this portal version does not know. An unknown option inside a known step renders as a plain row with its raw value.

## Canvas chrome

- The canvas has a fixed height (the window less the page header), so fitting the flow centres it.
- The minimap is hidden at rest and fades in while panning or zooming.
- The library's attribution is hidden.
- Backspace and Delete do nothing: the canvas is read-only.

## Side panel

Clicking a node opens a read-only side panel with:

- a description of what the step does and when the user sees it, for every step type the portal knows
- the step's name as a "Name" row
- the step's settings in plain words, including `optional` and what `target_step` refers to
- the step's YAML, serialized with `js-yaml`

Clicking the canvas background closes it.

## Edge cases

| Case | Display |
|---|---|
| `one_of` with no options (for example, account recovery on a username-only project) | Choice node with a warning callout, "No options. Users can't recover their account" (wording per flow type), and no outgoing edge |
| `signup_login` option referencing a flow name that does not exist | Flow node with a red border and "Flow not found" instead of the link |
| Unknown step type or option | Unknown node or plain row; never a crash |
| Unknown `type` or `name` URL parameter | Falls back to `signup_login` / `default` |
| Query loading or failed | `ShowLoading`, or `ShowError` with retry |

## Testing

- Jest unit tests for `flowToTree`, `treeToGraph` and `layoutGraph`, with fixtures checked into the portal as typed TypeScript objects, shared with Storybook:
  - Generated flows for email + Google + passkey + optional 2FA, dumped from Go with `GenerateAuthenticationFlowConfig` (#5995)
  - Generated flows for a username-only project (empty account recovery `one_of`)
  - The custom signup flow with `fill_form` used by local project cousin-4bx3ay
  - Synthetic flows written inline in tests, for branch ordering
- What the tests assert:
  - branches appear in config order, each above the next
  - every branch ends at the join
  - the join is to the right of every branch
  - an optional step has no bypass edge
  - a single-option `one_of` is inlined
  - an empty `one_of` has no outgoing edge
  - an unknown step type yields an Unknown node
- Assertions are on structure and relative positions, not pixels.
- Storybook stories rendering the canvas for each fixture, plus one synthetic flow containing every step type.
- Manual visual check by the reviewer:
  - cousin-4bx3ay: custom signup with `fill_form`, generated login
  - a username-only project: the empty account recovery warning
- CI parity: portal `typecheck`, `eslint`, `stylelint`, `prettier`, `test`, `gentype`, `build`.

## Deviation from DEV-3930

DEV-3930 suggested dagre or elk for layout. This design uses its own layout instead, so that branch order follows config order and the layout can be unit-tested deterministically.
