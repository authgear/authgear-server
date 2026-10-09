# Screen scaffold template

Use this template when creating a new portal settings screen. It follows `portal/src/graphql/portal/AnonymousUsersConfigurationScreen.tsx`.

## Template (data-driven settings screen)

```tsx
import React from "react";
import { useParams } from "react-router-dom";
import { Heading, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import FormContainer from "../../FormContainer";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";

const ExampleScreen: React.VFC = function ExampleScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useAppConfigForm({
    appID,
    constructFormState, // defined in this file, as in AnonymousUsersConfigurationScreen.tsx
    constructConfig,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <ScreenContent>
        <div>
          <Heading as="h1" size="5" weight="bold">
            <FormattedMessage id="ExampleScreen.title" />
          </Heading>
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="ExampleScreen.description" />
          </Text>
        </div>
        <SettingsSectionCard
          title={<FormattedMessage id="ExampleScreen.section.general" />}
        >
          {/* Form fields */}
        </SettingsSectionCard>
        <SaveFunctionBar />
      </ScreenContent>
    </FormContainer>
  );
};

export default ExampleScreen;
```

`constructFormState` and `constructConfig` are defined in the screen. Copy that pair from `AnonymousUsersConfigurationScreen.tsx` and change the fields.

## Variants

- Use `ScreenContent layout="list"` for list-heavy pages.
- Omit `FormContainer` and `SaveFunctionBar` when the page is read-only and has no submit flow.
- Add `NavBreadcrumb` (`portal/src/NavBreadcrumb.tsx`) when the page is not a top-level destination.
- Wrap with `ScreenLayoutScrollView` (`portal/src/ScreenLayoutScrollView.tsx`) when the screen needs the shared scroll body. `VerifyDomainScreen.tsx` does this.
- Keep `ShowLoading` / `ShowError` as the gate for async screens.
- For narrow vs full-width decisions, follow `layout-screen-content.md`.

## Checklist

- One `Heading as="h1"` per screen body.
- Title, description, and labels are in i18n (`FormattedMessage` / `renderToString`).
- Loading and error states render before main content.
- Sections use `SettingsSectionCard`, not a one-off card.
- A config form uses `FormContainer` and `SaveFunctionBar`.
