import React, { useCallback, useContext } from "react";
import { Dropdown } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  PortalAPIAppConfig,
  SecondaryAuthenticationMode,
  secondaryAuthenticationModes,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import Toggle from "../../Toggle";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useDropdown } from "../../hook/useInput";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./MFAConfigurationScreen.module.css";

const ALL_MFA_OPTIONS: SecondaryAuthenticationMode[] = [
  ...secondaryAuthenticationModes,
];

interface FormState {
  mfaMode: SecondaryAuthenticationMode;
  deviceTokenDisabled: boolean;
  recoveryCodeEnabled: boolean;
  numRecoveryCode: number | undefined;
  recoveryCodeListEnabled: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    mfaMode:
      config.authentication?.secondary_authentication_mode ?? "if_exists",
    deviceTokenDisabled: config.authentication?.device_token?.disabled ?? false,
    recoveryCodeEnabled: !(
      config.authentication?.recovery_code?.disabled ?? false
    ),
    numRecoveryCode: config.authentication?.recovery_code?.count,
    recoveryCodeListEnabled:
      config.authentication?.recovery_code?.list_enabled ?? false,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.device_token ??= {};
    config.authentication.recovery_code ??= {};

    config.authentication.secondary_authentication_mode = currentState.mfaMode;
    config.authentication.device_token.disabled =
      currentState.deviceTokenDisabled;

    config.authentication.recovery_code.disabled =
      !currentState.recoveryCodeEnabled;
    config.authentication.recovery_code.count = currentState.numRecoveryCode;
    config.authentication.recovery_code.list_enabled =
      currentState.recoveryCodeListEnabled;

    clearEmptyObject(config);
  });
}

interface MFAConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
}

const MFAConfigurationContent: React.VFC<MFAConfigurationContentProps> =
  function MFAConfigurationContent(props) {
    const { state, setState } = props.form;
    const {
      mfaMode,
      deviceTokenDisabled,
      recoveryCodeEnabled,
      recoveryCodeListEnabled,
      numRecoveryCode,
    } = state;
    const { renderToString } = useContext(Context);

    const renderMFAMode = useCallback(
      (key: SecondaryAuthenticationMode) => {
        return renderToString("MFAConfigurationScreen.policy.mode." + key);
      },
      [renderToString]
    );

    const { options: mfaModeOptions, onChange: onChangeMFAMode } = useDropdown(
      ALL_MFA_OPTIONS,
      (option) => {
        setState((prev) => ({
          ...prev,
          mfaMode: option,
        }));
      },
      mfaMode,
      renderMFAMode
    );

    const onChangeDeviceTokenDisabled = useCallback(
      (_e, checked?: boolean) => {
        setState((prev) => ({
          ...prev,
          deviceTokenDisabled: checked ?? false,
        }));
      },
      [setState]
    );

    const onChangeRecoveryCodeEnabled = useCallback(
      (_e, checked?: boolean) => {
        setState((prev) => ({
          ...prev,
          recoveryCodeEnabled: checked ?? false,
        }));
      },
      [setState]
    );

    const onChangeRecoveryCodeListEnabled = useCallback(
      (_e, checked?: boolean) => {
        setState((prev) => ({
          ...prev,
          recoveryCodeListEnabled: checked ?? false,
        }));
      },
      [setState]
    );

    const onChangeNumRecoveryCode = useCallback(
      (_, value?: string) => {
        setState((prev) => ({
          ...prev,
          numRecoveryCode: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="MFAConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="MFAConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="MFAConfigurationScreen.policy.title" />
          </WidgetTitle>
          <Dropdown
            label={renderToString("MFAConfigurationScreen.policy.mode.title")}
            options={mfaModeOptions}
            selectedKey={mfaMode}
            onChange={onChangeMFAMode}
          />
          <Toggle
            label={
              <FormattedMessage id="MFAConfigurationScreen.policy.device-token.title" />
            }
            inlineLabel={false}
            checked={deviceTokenDisabled}
            onChange={onChangeDeviceTokenDisabled}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="MFAConfigurationScreen.recovery-code.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="MFAConfigurationScreen.recovery-code.description" />
          </WidgetDescription>
          <Toggle
            label={
              <FormattedMessage id="MFAConfigurationScreen.recovery-code.toggle.title" />
            }
            inlineLabel={false}
            checked={recoveryCodeEnabled}
            onChange={onChangeRecoveryCodeEnabled}
          />
          <FormTextField
            disabled={!recoveryCodeEnabled}
            parentJSONPointer="/authentication/recovery_code"
            fieldName="count"
            label={renderToString(
              "MFAConfigurationScreen.recovery-code.input.title"
            )}
            value={numRecoveryCode?.toFixed(0) ?? ""}
            onChange={onChangeNumRecoveryCode}
          />
          <Toggle
            disabled={!recoveryCodeEnabled}
            label={
              <FormattedMessage id="MFAConfigurationScreen.recovery-code.list.toggle.title" />
            }
            inlineLabel={false}
            checked={recoveryCodeListEnabled}
            onChange={onChangeRecoveryCodeListEnabled}
          />
        </Widget>
      </ScreenContent>
    );
  };

const MFAConfigurationScreen: React.VFC = function MFAConfigurationScreen() {
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
      <MFAConfigurationContent form={form} />
    </FormContainer>
  );
};

export default MFAConfigurationScreen;
