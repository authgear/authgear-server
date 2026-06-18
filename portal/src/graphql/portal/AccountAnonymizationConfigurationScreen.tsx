import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
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
import styles from "./AccountAnonymizationConfigurationScreen.module.css";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import FormTextField from "../../FormTextField";
import { checkIntegerInput } from "../../util/input";

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

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="AccountAnonymizationConfigurationScreen.title" />
        </ScreenTitle>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AccountAnonymizationConfigurationScreen.anonymization-schedule.title" />
          </WidgetTitle>
          <FormTextField
            parentJSONPointer="/account_anonymization"
            fieldName="grace_period_days"
            label={renderToString(
              "AccountAnonymizationConfigurationScreen.grace-period.label"
            )}
            description={renderToString(
              "AccountAnonymizationConfigurationScreen.grace-period.description"
            )}
            value={grace_period_days}
            onChange={onChangeGracePeriod}
          />
        </Widget>
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
      <FormContainer form={form}>
        <AccountAnonymizationConfigurationContent form={form} />
      </FormContainer>
    );
  };

export default AccountAnonymizationConfigurationScreen;
