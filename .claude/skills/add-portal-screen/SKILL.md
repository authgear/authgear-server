---
name: add-portal-screen
description: Add screens or components to the portal frontend. Use when the user asks to create a new portal page, screen, tab, or UI component.
argument-hint: "<feature name or description>"
---

Follow this guide when adding screens or components to the portal.

## Step 1: Locate existing patterns

Before writing any code, read at least one existing screen from `portal/src/screens/` that is similar to the feature being built. Reading the full file is preferred so you understand the complete pattern.

**Good reference screens:**
- Config-form screen with tabs: `portal/src/screens/fraud-protection/FraudProtectionConfigurationScreen.tsx`
- Screen with sub-tab files: `portal/src/screens/api-resources/APIResourceDetailsScreen.tsx`

Also read any components under `portal/src/components/` in the same feature area that exist already.

---

## Step 2: File structure

### Screens

New screens go in:
```
portal/src/screens/<feature-name>/
  <FeatureName>Screen.tsx
  <FeatureName>Screen.module.css
```

- Use **kebab-case** for the directory name (e.g., `fraud-protection`, `api-resources`).
- Use **PascalCase** for file names (e.g., `FraudProtectionConfigurationScreen.tsx`).
- Sub-tabs of a screen that are complex enough to deserve their own file live in the **same screen directory**, not in `components/`.

### Components

Reusable or extracted sub-components go in:
```
portal/src/components/<feature-name>/
  <ComponentName>.tsx
  <ComponentName>.module.css
```

- A component that is only used by one screen is still placed in `components/` when it is large enough to deserve its own file.
- Shared low-level UI primitives (buttons, inputs, etc.) go in `portal/src/components/common/`.

### CSS modules

Every component or screen that has its own styles gets its own `.module.css` file. Name it identically to the `.tsx` file (e.g., `FraudProtectionOverviewTab.module.css`).

When two sibling components share many CSS classes (e.g., two card variants), they may both import from a single shared `.module.css` file. Name it after the dominant component (e.g., `OverviewMetricCard.module.css` shared by `OverviewMetricCard.tsx` and `OverviewEnforcementCard.tsx`).

---

## Step 3: Screen anatomy

A screen file always has two components:

### Outer component (data shell)

Handles routing params, loading/error states, and form setup. Renders nothing except loading spinners and error views until data is ready. Then renders `FormContainer` wrapping the inner content.

```tsx
const MyFeatureScreen: React.VFC = function MyFeatureScreen() {
  const { appID } = useParams() as { appID: string };

  // Data hooks
  const form = useAppConfigForm({ appID, constructFormState, constructConfig });
  const featureConfig = useAppFeatureConfigQuery(appID);

  // Tab navigation (if the screen has tabs)
  const { selectedKey, onLinkClick, onChangeKey } =
    usePivotNavigation<MyTab>(["overview", "settings"]);

  // Loading / error guards — render nothing else until resolved
  if (form.isLoading || featureConfig.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }
  if (featureConfig.loadError) {
    return <ShowError error={featureConfig.loadError} onRetry={featureConfig.reload} />;
  }

  return (
    <FormContainer form={form} canSave={isModifiable} ...>
      <MyFeatureContent
        form={form}
        featureConfig={featureConfig.effectiveFeatureConfig?.my_feature}
        selectedKey={selectedKey}
        onLinkClick={onLinkClick}
        onChangeKey={onChangeKey}
      />
    </FormContainer>
  );
};
```

### Inner content component

Receives the form model and config as props. Handles user interactions and renders the full UI. Callbacks (e.g., `onChange` handlers) are defined here using `useCallback`.

```tsx
interface MyFeatureContentProps {
  form: AppConfigFormModel<FormState>;
  featureConfig?: MyFeatureConfig;
  selectedKey: MyTab;
  onLinkClick: (item?: PivotItem) => void;
  onChangeKey: (key: MyTab) => void;
}

const MyFeatureContent: React.VFC<MyFeatureContentProps> =
  function MyFeatureContent(props) {
    const { form, featureConfig, selectedKey, onLinkClick, onChangeKey } = props;
    const { state, setState } = form;

    const onSomeFieldChange = useCallback(
      (_event, value?: string) => {
        setState((current) => ({ ...current, someField: value ?? "" }));
      },
      [setState]
    );

    return (
      <ScreenContent layout="list">
        <ScreenTitle className={styles.widget}>...</ScreenTitle>
        {/* ... */}
      </ScreenContent>
    );
  };
```

### Form state helpers

Define `constructFormState` and `constructConfig` as plain functions (not inside components) near the top of the file, after the `FormState` interface:

```tsx
interface FormState {
  enabled: boolean;
  someField: string;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    enabled: config.my_feature?.enabled ?? false,
    someField: config.my_feature?.some_field ?? "",
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.my_feature ??= {};
    draft.my_feature.enabled = currentState.enabled;
    draft.my_feature.some_field = currentState.someField || undefined;
    clearEmptyObject(draft);
  });
}
```

---

## Step 4: Register the screen in AppRoot

After creating the screen file, register it as a lazy-loaded route in `portal/src/AppRoot.tsx`.

### Add the lazy import (near other similar imports)

```tsx
const MyFeatureScreen = lazy(
  async () => import("./screens/my-feature/MyFeatureScreen")
);
```

### Add the route (inside the JSX route tree)

Find the appropriate `<Route>` section and add:

```tsx
<Route path="my-feature" element={<MyFeatureScreen />} />
```

---

## Step 5: Breaking down large components

When a component grows large, extract sub-components into `portal/src/components/<feature>/`. Good candidates for extraction:

- **A tab's content** when the tab has its own state or data-fetching logic
- **A repeated card/row pattern** rendered multiple times with different props
- **A self-contained panel** (e.g., a sidebar widget with its own toggle state)

### Extraction rules

1. **Each extracted component gets its own `.tsx` and `.module.css`.**
2. **Do not pass raw `form` or `setState` into sub-components.** Instead, pass the specific values and typed callback functions they need.
3. **Keep state as close to its consumer as possible.** If state is only used inside a sub-component (e.g., `showAll` for a list toggle), define it inside that sub-component — do not lift it up.
4. **A sub-component that is only rendered conditionally** (e.g., only when a tab is active) does not need an `isActive` prop — it is simply not rendered when inactive, so queries inside it should use `skip: !enabled` (or similar) rather than checking tab state.

### Example: extracting a tab

Before (all in one component):
```tsx
{selectedKey === "overview" ? (
  <section>
    {/* 200 lines of overview JSX */}
  </section>
) : null}
```

After:
```tsx
// In components/my-feature/MyFeatureOverviewTab.tsx
export interface MyFeatureOverviewTabProps {
  enabled: boolean;
  onChangeToSettings: () => void;
}

const MyFeatureOverviewTab: React.VFC<MyFeatureOverviewTabProps> = ...

// In the screen:
{selectedKey === "overview" ? (
  <MyFeatureOverviewTab enabled={state.enabled} onChangeToSettings={() => onChangeKey("settings")} />
) : null}
```

---

## Step 6: GraphQL queries

The portal has two separate GraphQL APIs with separate code-generation configs:

| API | Schema | Query files | Generated output | Use for |
|-----|--------|------------|-----------------|---------|
| Admin API | `portal/src/graphql/adminapi/schema.graphql` | `portal/src/graphql/adminapi/query/*.graphql` | `*.generated.ts` next to the `.graphql` file | App-level data (users, audit logs, fraud protection, etc.) |
| Portal API | `portal/src/graphql/portal/schema.graphql` | `portal/src/graphql/portal/query/*.graphql` | `*.generated.ts` next to the `.graphql` file | Portal-level data (subscriptions, feature config, app list, etc.) |

### Writing a query

1. Read the relevant schema file to understand available types and fields.
2. Create a `.graphql` file in the appropriate `query/` directory:

```graphql
# portal/src/graphql/adminapi/query/myFeatureQuery.graphql
query myFeatureQuery($appID: ID!, $rangeFrom: DateTime, $rangeTo: DateTime) {
  myFeature(appID: $appID, rangeFrom: $rangeFrom, rangeTo: $rangeTo) {
    totalCount
    someField
    nestedItems {
      id
      value
    }
  }
}
```

3. Run code generation from `portal/`:
```
npm run gentype
```

This produces a `myFeatureQuery.generated.ts` file next to the `.graphql` file containing typed hooks, variables types, and result types.

### Using the generated hook

**Always import from the `.generated.ts` file, never inline `gql` in `.ts`/`.tsx` files.**

```tsx
// ✅ Correct
import { useMyFeatureQueryQuery } from "../../graphql/adminapi/query/myFeatureQuery.generated";

// ❌ Wrong — do not write inline gql documents
const MY_QUERY = gql`query myFeatureQuery { ... }`;
```

The generated hook is a standard Apollo hook. Use Apollo options directly:

```tsx
const {
  data,
  loading,
  error,
  refetch,
} = useMyFeatureQueryQuery({
  skip: !enabled,
  variables: { appID, rangeFrom, rangeTo },
});

const result = data?.myFeature ?? null;

const onRetry = useCallback(() => {
  void refetch();
}, [refetch]);
```

### When to use which API

- **Admin API**: data tied to a specific app (audit logs, user data, app config overrides, fraud protection stats). The query variables typically include `appID`.
- **Portal API**: portal-level data (subscription info, feature flags, app list). The existing hooks like `useAppFeatureConfigQuery` use this API.

---

## Step 7: CSS module conventions

Use Tailwind utility classes via `@apply`. Never write raw CSS values when a Tailwind class exists.

```css
/* ✅ Good */
.metricCard {
  @apply flex flex-col rounded-lg border border-[#edebe9] bg-white px-[14px] py-4;
}

/* ❌ Avoid raw CSS when Tailwind equivalent exists */
.metricCard {
  display: flex;
  flex-direction: column;
  border-radius: 0.5rem;
  padding: 1rem;
}
```

For icon/badge variants that share a base style, group them in one selector then override per-variant:

```css
/* Shared base */
.metricIcon,
.metricIconSuccess,
.metricIconWarning,
.metricIconBlocked {
  @apply inline-flex h-8 w-8 flex-none items-center justify-center rounded-md text-[16px];
}

/* Per-variant color */
.metricIcon        { @apply bg-[#edf6ff] text-[#176df3]; }
.metricIconSuccess { @apply bg-[#eef6ef] text-[#16a34a]; }
.metricIconWarning { @apply bg-[#fff4ce] text-[#d97706]; }
.metricIconBlocked { @apply bg-[#fef2f2] text-[#dc2626]; }
```

Map variants via a lookup object in the component rather than building class strings dynamically:

```tsx
const iconVariantClass: Record<MetricIconVariant, string> = {
  default: styles.metricIcon,
  success: styles.metricIconSuccess,
  warning: styles.metricIconWarning,
  blocked: styles.metricIconBlocked,
};
```

---

## Step 8: Quality checks

After any changes to portal files, always run both checks and fix all errors before finishing:

```
cd portal
npm run typecheck
npm run eslint
```

Common issues to watch for:

- **`@typescript-eslint/no-unnecessary-condition`**: A condition that is always true/false because TypeScript can prove it at the call site. Usually caused by checking a prop value that is already guaranteed by context (e.g., `isActive={selectedKey === "overview"}` inside a branch where `selectedKey === "overview"` is already true).
- **Unused imports**: Remove any imports that are no longer referenced after refactoring.
- **Missing `useCallback`/`useMemo`**: Inline arrow functions passed as props to components should be wrapped in `useCallback` to avoid unnecessary re-renders.

---

## Quick reference

| Task | Command |
|------|---------|
| Generate TypeScript types from `.graphql` files | `cd portal && npm run gentype` |
| Type-check all TypeScript | `cd portal && npm run typecheck` |
| Lint all TypeScript/TSX | `cd portal && npm run eslint` |
| Auto-fix lint issues | `cd portal && npm run eslint:format` |
| Check CSS | `cd portal && npm run stylelint` |
