import React, { useMemo, useCallback, useContext } from "react";
import {
  Dropdown,
  Text,
  useTheme,
  MessageBar,
  MessageBarType,
} from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  PortalAPIAppConfig,
  PortalAPIFeatureConfig,
  PrimaryAuthenticatorType,
  SecondaryAuthenticationMode,
  secondaryAuthenticationModes,
  SecondaryAuthenticatorType,
  secondaryAuthenticatorTypes,
} from "../../types";
import { swap } from "../../OrderButtons";
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
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PriorityList, { PriorityListItem } from "../../PriorityList";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./MFAConfigurationScreen.module.css";

interface AuthenticatorTypeFormState<T> {
  isChecked: boolean;
  isDisabled: boolean;
  type: T;
}

const ALL_MFA_OPTIONS: SecondaryAuthenticationMode[] = [
  ...secondaryAuthenticationModes,
];

const secondaryAuthenticatorNameIds = {
  totp: "AuthenticatorType.secondary.totp",
  oob_otp_email: "AuthenticatorType.secondary.oob-otp-email",
  oob_otp_sms: "AuthenticatorType.secondary.oob-otp-phone",
  password: "AuthenticatorType.secondary.password",
};

interface FormState {
  mfaMode: SecondaryAuthenticationMode;
  deviceTokenEnabled: boolean;
  recoveryCodeEnabled: boolean;
  numRecoveryCode: number | undefined;
  recoveryCodeListEnabled: boolean;
  primary: PrimaryAuthenticatorType[];
  secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[] = (
    config.authentication?.secondary_authenticators ?? []
  ).map((t) => ({
    isChecked: true,
    isDisabled: false,
    type: t,
  }));
  for (const type of secondaryAuthenticatorTypes) {
    if (!secondary.some((t) => t.type === type)) {
      secondary.push({
        isChecked: false,
        isDisabled: false,
        type,
      });
    }
  }

  return {
    mfaMode:
      config.authentication?.secondary_authentication_mode ?? "if_exists",
    deviceTokenEnabled: !(
      config.authentication?.device_token?.disabled ?? false
    ),
    recoveryCodeEnabled: !(
      config.authentication?.recovery_code?.disabled ?? false
    ),
    numRecoveryCode: config.authentication?.recovery_code?.count,
    recoveryCodeListEnabled:
      config.authentication?.recovery_code?.list_enabled ?? false,
    primary: config.authentication?.primary_authenticators ?? [],
    secondary,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    function filterEnabled<T extends string>(
      s: AuthenticatorTypeFormState<T>[]
    ) {
      return s.filter((t) => t.isChecked).map((t) => t.type);
    }

    config.authentication ??= {};
    config.authentication.device_token ??= {};
    config.authentication.recovery_code ??= {};

    config.authentication.secondary_authenticators = filterEnabled(
      currentState.secondary
    );

    config.authentication.secondary_authentication_mode = currentState.mfaMode;
    config.authentication.device_token.disabled =
      !currentState.deviceTokenEnabled;

    config.authentication.recovery_code.disabled =
      !currentState.recoveryCodeEnabled;
    config.authentication.recovery_code.count = currentState.numRecoveryCode;
    config.authentication.recovery_code.list_enabled =
      currentState.recoveryCodeListEnabled;

    clearEmptyObject(config);
  });
}

interface UnreasonableWarningProps {
  primary: PrimaryAuthenticatorType[];
  secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[];
}

function UnreasonableWarning(props: UnreasonableWarningProps) {
  const { primary, secondary } = props;

  const unreasonableTypes = useMemo(() => {
    const out: PrimaryAuthenticatorType[] = [];
    for (const p of primary) {
      for (const s of secondary) {
        if (s.type === p && s.isChecked) {
          out.push(p);
        }
      }
    }
    return out;
  }, [primary, secondary]);

  if (unreasonableTypes.length <= 0) {
    return null;
  }

  return (
    <div>
      {unreasonableTypes.map((t) => {
        return (
          <MessageBar key={t} messageBarType={MessageBarType.info}>
            <FormattedMessage id={"MFAConfigurationScreen.unreasonable." + t} />
          </MessageBar>
        );
      })}
    </div>
  );
}

interface MFAConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const MFAConfigurationContent: React.VFC<MFAConfigurationContentProps> =
  function MFAConfigurationContent(props) {
    const { featureConfig } = props;
    const { state, setState } = props.form;
    const {
      mfaMode,
      deviceTokenEnabled,
      recoveryCodeEnabled,
      recoveryCodeListEnabled,
      numRecoveryCode,
      primary,
      secondary,
    } = state;
    const { renderToString } = useContext(Context);
    const {
      semanticColors: { disabledText },
    } = useTheme();

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

    const featureDisabled = useMemo(() => {
      return (
        featureConfig?.authentication?.secondary_authenticators?.oob_otp_sms
          ?.disabled ?? false
      );
    }, [featureConfig]);

    const secondaryItems: PriorityListItem[] = useMemo(
      () =>
        secondary.map(({ type, isChecked, isDisabled }) => {
          const disabled =
            isDisabled || (type === "oob_otp_sms" && featureDisabled);
          return {
            disabled,
            key: type,
            checked: isChecked,
            content: (
              <div>
                <Text
                  variant="small"
                  styles={{
                    root: {
                      color: disabled ? disabledText : undefined,
                    },
                  }}
                >
                  <FormattedMessage id={secondaryAuthenticatorNameIds[type]} />
                </Text>
              </div>
            ),
          };
        }),
      [secondary, featureDisabled, disabledText]
    );

    const onChangeDeviceTokenEnabled = useCallback(
      (_e, checked?: boolean) => {
        setState((prev) => ({
          ...prev,
          deviceTokenEnabled: checked ?? false,
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

    const onChangeSecondaryAuthenticatorChecked = useCallback(
      (key: string, checked: boolean) => {
        setState((state) =>
          produce(state, (state) => {
            const t = state.secondary.find((t) => t.type === key);
            if (t != null) {
              t.isChecked = checked;
            }
          })
        );
      },
      [setState]
    );

    const onSwapSecondaryAuthenticator = useCallback(
      (index1: number, index2: number) => {
        setState((prev) => ({
          ...prev,
          secondary: swap(prev.secondary, index1, index2),
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
            checked={deviceTokenEnabled}
            onChange={onChangeDeviceTokenEnabled}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="MFAConfigurationScreen.authenticator.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="MFAConfigurationScreen.authenticator.description" />
          </WidgetDescription>
          {featureDisabled ? (
            <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
          ) : null}
          <PriorityList
            items={secondaryItems}
            checkedColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.activate"
            )}
            keyColumnLabel={renderToString(
              "AuthenticatorConfigurationScreen.columns.authenticator"
            )}
            onChangeChecked={onChangeSecondaryAuthenticatorChecked}
            onSwap={onSwapSecondaryAuthenticator}
          />
          <UnreasonableWarning primary={primary} secondary={secondary} />
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

  const featureConfig = useAppFeatureConfigQuery(appID);

  if (form.isLoading || featureConfig.loading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (featureConfig.error != null) {
    return (
      <ShowError error={featureConfig.error} onRetry={featureConfig.refetch} />
    );
  }

  return (
    <FormContainer form={form}>
      <MFAConfigurationContent
        form={form}
        featureConfig={featureConfig.effectiveFeatureConfig ?? undefined}
      />
    </FormContainer>
  );
};

export default MFAConfigurationScreen;
