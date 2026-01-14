import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import { MessageBar } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import { produce } from "immer";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { PortalAPIAppConfig } from "../../types";
import styles from "./AccountDeletionConfigurationScreen.module.css";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import FormTextField from "../../FormTextField";
import Toggle from "../../Toggle";
import { checkIntegerInput } from "../../util/input";

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

    const { renderToString } = useContext(Context);

    const onChangeGracePeriod = useCallback(
      (_e, value?: string) => {
        if (value != null) {
          if (checkIntegerInput(value)) {
            setState((prev) => {
              return {
                ...prev,
                grace_period_days: value,
              };
            });
          }
        }
      },
      [setState]
    );

    const onChangeEnabled = useCallback(
      (_e, checked?: boolean) => {
        if (checked != null) {
          setState((prev) => {
            return {
              ...prev,
              scheduled_by_end_user_enabled: checked,
            };
          });
        }
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="AccountDeletionConfigurationScreen.title" />
        </ScreenTitle>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AccountDeletionConfigurationScreen.deletion-schedule.title" />
          </WidgetTitle>
          <FormTextField
            parentJSONPointer="/account_deletion"
            fieldName="grace_period_days"
            label={renderToString(
              "AccountDeletionConfigurationScreen.grace-period.label"
            )}
            description={renderToString(
              "AccountDeletionConfigurationScreen.grace-period.description"
            )}
            value={grace_period_days}
            onChange={onChangeGracePeriod}
          />
          <Toggle
            checked={scheduled_by_end_user_enabled}
            onChange={onChangeEnabled}
            label={renderToString(
              "AccountDeletionConfigurationScreen.scheduled-by-end-user.label"
            )}
            inlineLabel={true}
          />
          <MessageBar>
            <FormattedMessage id="AccountDeletionConfigurationScreen.apple-app-store.description" />
          </MessageBar>
        </Widget>
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
      <FormContainer form={form}>
        <AccountDeletionConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AccountDeletionConfigurationScreen;
