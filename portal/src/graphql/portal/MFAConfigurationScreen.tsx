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
  PrimaryAuthenticatorType,
  SecondaryAuthenticationMode,
  secondaryAuthenticationModes,
  SecondaryAuthenticatorType,
  secondaryAuthenticatorTypes,
  PortalAPIFeatureConfig,
  AuthenticatorPasswordConfig,
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
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useDropdown } from "../../hook/useInput";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import PriorityList, { PriorityListItem } from "../../PriorityList";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./MFAConfigurationScreen.module.css";
import PasswordSettings, {
  ResetPasswordWithEmailMethod,
  ResetPasswordWithPhoneMethod,
  getResetPasswordWithEmailMethod,
  getResetPasswordWithPhoneMethod,
  setUIForgotPasswordConfig,
} from "./PasswordSettings";
import { formatDuration, parseDuration } from "../../util/duration";

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

interface ConfigFormState {
  mfaMode: SecondaryAuthenticationMode;
  mfaGlobalGracePeriodEnabled: boolean;
  deviceTokenEnabled: boolean;
  recoveryCodeEnabled: boolean;
  numRecoveryCode: number | undefined;
  recoveryCodeListEnabled: boolean;
  primary: PrimaryAuthenticatorType[];
  secondary: AuthenticatorTypeFormState<SecondaryAuthenticatorType>[];

  forgotPasswordLinkValidPeriodSeconds: number | undefined;
  forgotPasswordCodeValidPeriodSeconds: number | undefined;
  authenticatorPasswordConfig: AuthenticatorPasswordConfig;

  resetPasswordWithEmailBy: ResetPasswordWithEmailMethod;
  resetPasswordWithPhoneBy: ResetPasswordWithPhoneMethod;
}

interface FeatureConfigFormState {
  featureConfig?: PortalAPIFeatureConfig;
}

interface FormState extends ConfigFormState, FeatureConfigFormState {}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
}

function constructForgotpasswordValidPeriods(config: PortalAPIAppConfig) {
  const forgotPasswordLinkValidPeriod =
    config.forgot_password?.valid_periods?.link;
  const forgotPasswordLinkValidPeriodSeconds = forgotPasswordLinkValidPeriod
    ? parseDuration(forgotPasswordLinkValidPeriod)
    : undefined;

  const forgotPasswordCodeValidPeriod =
    config.forgot_password?.valid_periods?.code;
  const forgotPasswordCodeValidPeriodSeconds = forgotPasswordCodeValidPeriod
    ? parseDuration(forgotPasswordCodeValidPeriod)
    : undefined;

  return {
    forgotPasswordLinkValidPeriodSeconds,
    forgotPasswordCodeValidPeriodSeconds,
  };
}

// eslint-disable-next-line complexity
function constructFormState(config: PortalAPIAppConfig): ConfigFormState {
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

  const {
    forgotPasswordCodeValidPeriodSeconds,
    forgotPasswordLinkValidPeriodSeconds,
  } = constructForgotpasswordValidPeriods(config);

  return {
    mfaMode:
      config.authentication?.secondary_authentication_mode ?? "if_exists",
    deviceTokenEnabled: !(
      config.authentication?.device_token?.disabled ?? false
    ),
    mfaGlobalGracePeriodEnabled:
      config.authentication?.secondary_authentication_grace_period?.enabled ??
      false,
    recoveryCodeEnabled: !(
      config.authentication?.recovery_code?.disabled ?? false
    ),
    numRecoveryCode: config.authentication?.recovery_code?.count,
    recoveryCodeListEnabled:
      config.authentication?.recovery_code?.list_enabled ?? false,
    primary: config.authentication?.primary_authenticators ?? [],
    secondary,
    authenticatorPasswordConfig: {
      force_change: config.authenticator?.password?.force_change,
      policy: {
        min_length: config.authenticator?.password?.policy?.min_length ?? 8,
        uppercase_required:
          config.authenticator?.password?.policy?.uppercase_required ?? false,
        lowercase_required:
          config.authenticator?.password?.policy?.lowercase_required ?? false,
        alphabet_required:
          config.authenticator?.password?.policy?.alphabet_required ?? false,
        digit_required:
          config.authenticator?.password?.policy?.digit_required ?? false,
        symbol_required:
          config.authenticator?.password?.policy?.symbol_required ?? false,
        minimum_guessable_level:
          config.authenticator?.password?.policy?.minimum_guessable_level ??
          (0 as const),
        excluded_keywords:
          config.authenticator?.password?.policy?.excluded_keywords ?? [],
        history_size: config.authenticator?.password?.policy?.history_size ?? 0,
        history_days: config.authenticator?.password?.policy?.history_days ?? 0,
      },
      expiry: {
        force_change: {
          enabled:
            config.authenticator?.password?.expiry?.force_change?.enabled,
          duration_since_last_update:
            config.authenticator?.password?.expiry?.force_change
              ?.duration_since_last_update,
        },
      },
    },
    forgotPasswordLinkValidPeriodSeconds,
    forgotPasswordCodeValidPeriodSeconds,
    resetPasswordWithEmailBy: getResetPasswordWithEmailMethod(config),
    resetPasswordWithPhoneBy: getResetPasswordWithPhoneMethod(config),
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState,
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
    config.authenticator ??= {};
    config.forgot_password ??= {};

    config.authentication.secondary_authenticators = filterEnabled(
      currentState.secondary
    );

    config.authentication.secondary_authentication_mode = currentState.mfaMode;
    config.authentication.secondary_authentication_grace_period = {
      enabled: currentState.mfaGlobalGracePeriodEnabled,
    };
    config.authentication.device_token.disabled =
      !currentState.deviceTokenEnabled;

    config.authentication.recovery_code.disabled =
      !currentState.recoveryCodeEnabled;
    config.authentication.recovery_code.count = currentState.numRecoveryCode;
    config.authentication.recovery_code.list_enabled =
      currentState.recoveryCodeListEnabled;
    // currentState.authenticatorPasswordConfig may contain deprecated fields generated by SetDefaults.
    // So we explicitly save force_change and policy here only.
    config.authenticator.password = {
      force_change: currentState.authenticatorPasswordConfig.force_change,
      policy: currentState.authenticatorPasswordConfig.policy,
      expiry: currentState.authenticatorPasswordConfig.expiry,
    };

    if (currentState.forgotPasswordLinkValidPeriodSeconds != null) {
      config.forgot_password.valid_periods ??= {};
      config.forgot_password.valid_periods.link = formatDuration(
        currentState.forgotPasswordLinkValidPeriodSeconds,
        "s"
      );
    }
    if (currentState.forgotPasswordCodeValidPeriodSeconds != null) {
      config.forgot_password.valid_periods ??= {};
      config.forgot_password.valid_periods.code = formatDuration(
        currentState.forgotPasswordCodeValidPeriodSeconds,
        "s"
      );
    }

    setUIForgotPasswordConfig(config, currentState);

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
  form: FormModel;
  isLoginIDPhoneEnabled: boolean;
  isLoginIDEmailEnabled: boolean;
}

const MFAConfigurationContent: React.VFC<MFAConfigurationContentProps> =
  function MFAConfigurationContent(props) {
    const { isLoginIDEmailEnabled, isLoginIDPhoneEnabled } = props;
    const { state, setState } = props.form;
    const {
      mfaMode,
      mfaGlobalGracePeriodEnabled,
      deviceTokenEnabled,
      recoveryCodeEnabled,
      recoveryCodeListEnabled,
      numRecoveryCode,
      primary,
      secondary,
      featureConfig,
      forgotPasswordLinkValidPeriodSeconds,
      forgotPasswordCodeValidPeriodSeconds,
      resetPasswordWithEmailBy,
      resetPasswordWithPhoneBy,
      authenticatorPasswordConfig,
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
          mfaGlobalGracePeriodEnabled:
            option !== "required" ? false : mfaGlobalGracePeriodEnabled,
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

    const showPasswordSettings = useMemo(() => {
      return secondary.some((a) => a.type === "password" && a.isChecked);
    }, [secondary]);

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

    const onChangeMFAGlobalGracePeriodEnabled = useCallback(
      (_e, checked?: boolean) => {
        setState((prev) => ({
          ...prev,
          mfaGlobalGracePeriodEnabled: checked ?? false,
        }));
      },
      [setState]
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
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
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
            {mfaMode === "required" ? (
              <div>
                <Toggle
                  label={
                    <FormattedMessage id="MFAConfigurationScreen.policy.enable-global-grace-period.title" />
                  }
                  inlineLabel={false}
                  checked={mfaGlobalGracePeriodEnabled}
                  onChange={onChangeMFAGlobalGracePeriodEnabled}
                />
                <WidgetDescription>
                  <FormattedMessage id="MFAConfigurationScreen.policy.enable-global-grace-period.description" />
                </WidgetDescription>
              </div>
            ) : null}
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
          {showPasswordSettings ? (
            <PasswordSettings
              className={styles.widget}
              forgotPasswordLinkValidPeriodSeconds={
                forgotPasswordLinkValidPeriodSeconds
              }
              forgotPasswordCodeValidPeriodSeconds={
                forgotPasswordCodeValidPeriodSeconds
              }
              resetPasswordWithEmailBy={resetPasswordWithEmailBy}
              resetPasswordWithPhoneBy={resetPasswordWithPhoneBy}
              authenticatorPasswordConfig={authenticatorPasswordConfig}
              passwordPolicyFeatureConfig={
                featureConfig?.authenticator?.password?.policy
              }
              isLoginIDPhoneEnabled={isLoginIDPhoneEnabled}
              isLoginIDEmailEnabled={isLoginIDEmailEnabled}
              setState={setState}
            />
          ) : null}
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
        </ShowOnlyIfSIWEIsDisabled>
      </ScreenContent>
    );
  };

const MFAConfigurationScreen: React.VFC = function MFAConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const featureConfig = useAppFeatureConfigQuery(appID);
  const configForm = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  const isLoginIDEmailEnabled = useMemo(() => {
    return (
      configForm.effectiveConfig.identity?.login_id?.keys?.find(
        (k) => k.type === "email"
      ) != null
    );
  }, [configForm.effectiveConfig.identity?.login_id?.keys]);

  const isLoginIDPhoneEnabled = useMemo(() => {
    return (
      configForm.effectiveConfig.identity?.login_id?.keys?.find(
        (k) => k.type === "phone"
      ) != null
    );
  }, [configForm.effectiveConfig.identity?.login_id?.keys]);

  const state = useMemo<FormState>(() => {
    return {
      featureConfig: featureConfig.effectiveFeatureConfig,
      ...configForm.state,
    };
  }, [featureConfig.effectiveFeatureConfig, configForm.state]);

  const form: FormModel = {
    isLoading: configForm.isLoading || featureConfig.loading,
    isUpdating: configForm.isUpdating,
    isDirty: configForm.isDirty,
    loadError: configForm.loadError ?? featureConfig.error,
    updateError: configForm.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      configForm.setState(() => newState);
    },
    reload: () => {
      configForm.reload();
      featureConfig.refetch().finally(() => {});
    },
    reset: () => {
      configForm.reset();
    },
    save: async () => {
      await configForm.save();
    },
  };

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <MFAConfigurationContent
        form={form}
        isLoginIDEmailEnabled={isLoginIDEmailEnabled}
        isLoginIDPhoneEnabled={isLoginIDPhoneEnabled}
      />
    </FormContainer>
  );
};

export default MFAConfigurationScreen;
