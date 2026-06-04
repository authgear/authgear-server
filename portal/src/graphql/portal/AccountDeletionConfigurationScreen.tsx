import React, { useCallback, useRef } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import { Callout, Text } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../intl";
import { produce } from "immer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import { PortalAPIAppConfig } from "../../types";
import styles from "./AccountDeletionConfigurationScreen.module.css";
import { TextField } from "../../components/v2/TextField/TextField";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { checkIntegerInput } from "../../util/input";
import ExternalLink from "../../ExternalLink";
import { useFormContainerBaseContext } from "../../FormContainerBase";

interface FormState {
  scheduled_by_end_user_enabled: boolean;
  grace_period_days: string;
}

interface AccountDeletionConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const scheduled_by_end_user_enabled =
    config.account_deletion?.scheduled_by_end_user_enabled ?? false;
  const grace_period_days = String(
    config.account_deletion?.grace_period_days ?? 30
  );
  return {
    scheduled_by_end_user_enabled,
    grace_period_days,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.account_deletion ??= {};

    config.account_deletion.scheduled_by_end_user_enabled =
      currentState.scheduled_by_end_user_enabled;

    const n = parseInt(currentState.grace_period_days, 10);
    if (!isNaN(n)) {
      config.account_deletion.grace_period_days = n;
    }
  });
}

const AccountDeletionConfigurationContent: React.VFC<AccountDeletionConfigurationContentProps> =
  function AccountDeletionConfigurationContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const { scheduled_by_end_user_enabled, grace_period_days } = state;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onChangeGracePeriod = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        if (checkIntegerInput(value)) {
          setState((prev) => ({
            ...prev,
            grace_period_days: value,
          }));
        }
      },
      [setState]
    );

    const onChangeEnabled = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          scheduled_by_end_user_enabled: checked,
        }));
      },
      [setState]
    );

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={styles.contentWidthAnchor}
          aria-hidden
        />
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AccountDeletionConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="AccountDeletionConfigurationScreen.description" />
          </Text>
        </div>

        <div
          className={cn(
            styles.widget,
            "border border-[var(--gray-5)] rounded-lg p-6 flex gap-8 bg-white",
            isDirty && styles.settingsCardSaveBarClearance
          )}
        >
          <Text as="p" size="3" weight="medium" className="shrink-0 w-[200px]">
            <FormattedMessage id="AccountDeletionConfigurationScreen.deletion-schedule.title" />
          </Text>
          <div className="flex-1 flex flex-col gap-4 min-w-0">
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="AccountDeletionConfigurationScreen.grace-period.label" />
              }
              hint={
                <FormattedMessage id="AccountDeletionConfigurationScreen.grace-period.description" />
              }
              value={grace_period_days}
              onChange={onChangeGracePeriod}
              parentJSONPointer="/account_deletion"
              fieldName="grace_period_days"
            />
            <Toggle
              checked={scheduled_by_end_user_enabled}
              onCheckedChange={onChangeEnabled}
              text={
                <FormattedMessage id="AccountDeletionConfigurationScreen.scheduled-by-end-user.label" />
              }
            />
            <Callout.Root color="blue" variant="surface" size="1">
              <Callout.Icon>
                <InfoCircledIcon />
              </Callout.Icon>
              <Callout.Text>
                <FormattedMessage
                  id="AccountDeletionConfigurationScreen.apple-app-store.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    ExternalLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://developer.apple.com/app-store/review/guidelines/#5.1.1">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              </Callout.Text>
            </Callout.Root>
          </div>
        </div>

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const AccountDeletionConfigurationScreen: React.VFC =
  function AccountDeletionConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} hideFooterComponent={true}>
        <AccountDeletionConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AccountDeletionConfigurationScreen;
