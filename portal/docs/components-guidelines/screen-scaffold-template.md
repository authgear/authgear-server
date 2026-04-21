# Screen scaffold template

Use this template when creating a new portal screen so structure and behavior stay consistent.

## Template (data-driven settings screen)

```tsx
import React, { useMemo } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import NavBreadcrumb from "../../NavBreadcrumb";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import FormContainer from "../../FormContainer";

const ExampleScreen: React.VFC = function ExampleScreen() {
  const { appID } = useParams() as { appID: string };

  // Replace this with real query/form hooks.
  const isLoading = false;
  const error: unknown = null;
  const retry = () => {};

  const breadcrumbItems = useMemo(
    () => [
      {
        to: "~/settings",
        label: <FormattedMessage id="SettingsScreen.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="ExampleScreen.title" />,
      },
    ],
    []
  );

  if (isLoading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={retry} />;
  }

  return (
    <ScreenLayoutScrollView>
      <ScreenContent>
        <NavBreadcrumb items={breadcrumbItems} />
        <ScreenTitle>
          <FormattedMessage id="ExampleScreen.title" />
        </ScreenTitle>
        <ScreenDescription>
          <FormattedMessage id="ExampleScreen.description" />
        </ScreenDescription>
        <FormContainer>
          <Widget>
            <WidgetTitle>
              <FormattedMessage id="ExampleScreen.section.general" />
            </WidgetTitle>
            {/* Form fields or feature content here */}
          </Widget>
        </FormContainer>
      </ScreenContent>
    </ScreenLayoutScrollView>
  );
};

export default ExampleScreen;
```

## Variants

- Use `ScreenContent layout="list"` for list-heavy pages.
- Omit `FormContainer` when the page is read-only and has no submit flow.
- Omit `NavBreadcrumb` for top-level/root pages when breadcrumb adds no value.
- Keep `ShowLoading` / `ShowError` at the top-level gate for async screens.
- For narrow vs full-width decisions, follow `docs/portal/layout-screen-content.md` width policy.

## Checklist

- One `ScreenTitle` per screen body.
- Title/description/labels are in i18n (`FormattedMessage` / `renderToString`).
- Loading and error states are handled before rendering main content.
- Shared layout components are used (`ScreenLayoutScrollView`, `ScreenContent`).
- No ad-hoc replacements for standard screen skeleton unless there is a clear exception.
