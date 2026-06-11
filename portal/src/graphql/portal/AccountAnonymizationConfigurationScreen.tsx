import React, { useCallback, useRef } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
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
import styles from "./AccountAnonymizationConfigurationScreen.module.css";
import { TextField } from "../../components/v2/TextField/TextField";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { checkIntegerInput } from "../../util/input";
import { useFormContainerBaseContext } from "../../FormContainerBase";

interface FormState {
  grace_period_days: string;
}

interface AccountAnonymizationConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const grace_period_days = String(
    config.account_anonymization?.grace_period_days ?? 30
  );
  return {
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
    config.account_anonymization ??= {};

    const n = parseInt(currentState.grace_period_days, 10);
    if (!isNaN(n)) {
      config.account_anonymization.grace_period_days = n;
    }
  });
}

const AccountAnonymizationConfigurationContent: React.VFC<AccountAnonymizationConfigurationContentProps> =
  function AccountAnonymizationConfigurationContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const { grace_period_days } = state;
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

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AccountAnonymizationConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="AccountAnonymizationConfigurationScreen.description" />
          </Text>
        </div>

        <SettingsSectionCard
          className={cn(
            styles.widget,
            isDirty && styles.settingsCardSaveBarClearance
          )}
          contentClassName="gap-4"
          title={
            <FormattedMessage id="AccountAnonymizationConfigurationScreen.anonymization-schedule.title" />
          }
        >
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="AccountAnonymizationConfigurationScreen.grace-period.label" />
              }
              hint={
                <FormattedMessage id="AccountAnonymizationConfigurationScreen.grace-period.description" />
              }
              value={grace_period_days}
              onChange={onChangeGracePeriod}
              parentJSONPointer="/account_anonymization"
              fieldName="grace_period_days"
            />
        </SettingsSectionCard>

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const AccountAnonymizationConfigurationScreen: React.VFC =
  function AccountAnonymizationConfigurationScreen() {
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
        <AccountAnonymizationConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AccountAnonymizationConfigurationScreen;
