import React, { useMemo, useCallback, useContext, useRef } from "react";
import cn from "classnames";
import { Flex, RadioGroup, Text } from "@radix-ui/themes";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
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
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { TextField } from "../../components/v2/TextField/TextField";
import { FormField } from "../../components/v2/FormField/FormField";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { Callout } from "../../components/v2/Callout/Callout";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
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
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  RedMessageBar_RemindConfigureSMSProviderInNonSMSProviderScreen,
  RedMessageBar_RemindConfigureSMTPInNonSMTPConfigurationScreen,
} from "../../RedMessageBar";

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

interface SecretConfigFormState {
  smsProviderConfigured: boolean;
  smtpConfigured: boolean;
}

interface FormState
  extends ConfigFormState,
    FeatureConfigFormState,
    SecretConfigFormState {}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  getIsDirty: () => boolean;
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

    if (!currentState.mfaGlobalGracePeriodEnabled) {
      config.authentication.secondary_authentication_grace_period = undefined;
    } else {
      config.authentication.secondary_authentication_grace_period = {
        enabled: currentState.mfaGlobalGracePeriodEnabled,
      };
    }

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
    <div className={styles.warningList}>
      {unreasonableTypes.map((t) => {
        return (
          <Callout
            key={t}
            className="w-full"
            type="info"
            showCloseButton={false}
            text={
              <FormattedMessage id={"MFAConfigurationScreen.unreasonable." + t} />
            }
          />
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
    const { isAuthgearOnce } = useSystemConfig();
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

      smsProviderConfigured,
      smtpConfigured,
    } = state;
    const { renderToString } = useContext(Context);
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onChangeMFAMode = useCallback(
      (value: string) => {
        const option = value as SecondaryAuthenticationMode;
        setState((prev) => ({
          ...prev,
          mfaMode: option,
          mfaGlobalGracePeriodEnabled:
            option !== "required" ? false : prev.mfaGlobalGracePeriodEnabled,
        }));
      },
      [setState]
    );

    const isSMSRequiredForSomeEnabledFeatures = useMemo(() => {
      return secondary
        .filter((c) => c.type === "oob_otp_sms")
        .some((c) => c.isChecked);
    }, [secondary]);

    const isSMTPRequiredForSomeEnabledFeatures = useMemo(() => {
      return secondary
        .filter((c) => c.type === "oob_otp_email")
        .some((c) => c.isChecked);
    }, [secondary]);

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
              <Text size="2" color={disabled ? "gray" : undefined}>
                <FormattedMessage id={secondaryAuthenticatorNameIds[type]} />
              </Text>
            ),
          };
        }),
      [secondary, featureDisabled]
    );

    const onChangeMFAGlobalGracePeriodEnabled = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          mfaGlobalGracePeriodEnabled: checked,
        }));
      },
      [setState]
    );

    const onChangeDeviceTokenEnabled = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          deviceTokenEnabled: checked,
        }));
      },
      [setState]
    );

    const onChangeRecoveryCodeEnabled = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          recoveryCodeEnabled: checked,
        }));
      },
      [setState]
    );

    const onChangeRecoveryCodeListEnabled = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          recoveryCodeListEnabled: checked,
        }));
      },
      [setState]
    );

    const onChangeNumRecoveryCode = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setState((prev) => ({
          ...prev,
          numRecoveryCode: parseIntegerAllowLeadingZeros(e.target.value),
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
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="MFAConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="MFAConfigurationScreen.description" />
          </Text>
        </div>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          {isAuthgearOnce &&
          isSMSRequiredForSomeEnabledFeatures &&
          !smsProviderConfigured ? (
            <RedMessageBar_RemindConfigureSMSProviderInNonSMSProviderScreen
              className={styles.widget}
            />
          ) : null}
          {isAuthgearOnce &&
          isSMTPRequiredForSomeEnabledFeatures &&
          !smtpConfigured ? (
            <RedMessageBar_RemindConfigureSMTPInNonSMTPConfigurationScreen
              className={styles.widget}
            />
          ) : null}
          <SettingsSectionCard
            className={styles.widget}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="MFAConfigurationScreen.policy.title" />
            }
          >
            <FormField
              size="2"
              labelSize="2"
              labelSpace="1"
              label={
                <FormattedMessage id="MFAConfigurationScreen.policy.mode.title" />
              }
            >
              <RadioGroup.Root
                value={mfaMode}
                onValueChange={onChangeMFAMode}
                className={styles.modeRadioGroup}
              >
                <Flex direction="column" gap="3">
                  {ALL_MFA_OPTIONS.map((option) => (
                    <div key={option} className={styles.modeRadioBlock}>
                      <Text
                        as="label"
                        size="2"
                        className={styles.modeRadioOption}
                      >
                        <Flex gap="2" align="start">
                          <RadioGroup.Item
                            value={option}
                            className={styles.modeRadioItem}
                          />
                          <div className={styles.modeRadioContent}>
                            <Text as="span" size="2">
                              <FormattedMessage
                                id={`MFAConfigurationScreen.policy.mode.${option}.label`}
                              />
                            </Text>
                            <Text as="p" size="1" color="gray">
                              <FormattedMessage
                                id={`MFAConfigurationScreen.policy.mode.${option}.description`}
                              />
                            </Text>
                          </div>
                        </Flex>
                      </Text>
                      {option === "required" && mfaMode === "required" ? (
                        <Flex
                          gap="2"
                          align="start"
                          className={styles.gracePeriodRow}
                        >
                          <span
                            className={styles.modeRadioItemSpacer}
                            aria-hidden={true}
                          />
                          <div className={styles.gracePeriodContent}>
                            <Toggle
                              checked={mfaGlobalGracePeriodEnabled}
                              onCheckedChange={
                                onChangeMFAGlobalGracePeriodEnabled
                              }
                              textWeight="medium"
                              text={
                                <FormattedMessage id="MFAConfigurationScreen.policy.enable-global-grace-period.title" />
                              }
                            />
                            <Text as="p" size="1" color="gray">
                              <FormattedMessage id="MFAConfigurationScreen.policy.enable-global-grace-period.description" />
                            </Text>
                          </div>
                        </Flex>
                      ) : null}
                    </div>
                  ))}
                </Flex>
              </RadioGroup.Root>
            </FormField>
            <Toggle
              checked={deviceTokenEnabled}
              onCheckedChange={onChangeDeviceTokenEnabled}
              textWeight="medium"
              text={
                <FormattedMessage id="MFAConfigurationScreen.policy.device-token.title" />
              }
            />
          </SettingsSectionCard>
          <SettingsSectionCard
            className={styles.widget}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="MFAConfigurationScreen.authenticator.title" />
            }
          >
            <Text
              as="p"
              size="2"
              color="gray"
              className={styles.sectionDescription}
            >
              <FormattedMessage id="MFAConfigurationScreen.authenticator.description" />
            </Text>
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
          </SettingsSectionCard>
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
          <SettingsSectionCard
            className={cn(
              styles.widget,
              isDirty && styles.settingsCardSaveBarClearance
            )}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="MFAConfigurationScreen.recovery-code.title" />
            }
          >
            <Text
              as="p"
              size="2"
              color="gray"
              className={styles.sectionDescription}
            >
              <FormattedMessage id="MFAConfigurationScreen.recovery-code.description" />
            </Text>
            <Toggle
              checked={recoveryCodeEnabled}
              onCheckedChange={onChangeRecoveryCodeEnabled}
              textWeight="medium"
              text={
                <FormattedMessage id="MFAConfigurationScreen.recovery-code.toggle.title" />
              }
            />
            {recoveryCodeEnabled ? (
              <>
                <TextField
                  size="2"
                  labelSize="2"
                  type="text"
                  parentJSONPointer="/authentication/recovery_code"
                  fieldName="count"
                  label={
                    <FormattedMessage id="MFAConfigurationScreen.recovery-code.input.title" />
                  }
                  value={numRecoveryCode?.toFixed(0) ?? ""}
                  onChange={onChangeNumRecoveryCode}
                />
                <Toggle
                  checked={recoveryCodeListEnabled}
                  onCheckedChange={onChangeRecoveryCodeListEnabled}
                  textWeight="medium"
                  text={
                    <FormattedMessage id="MFAConfigurationScreen.recovery-code.list.toggle.title" />
                  }
                />
              </>
            ) : null}
          </SettingsSectionCard>
        </ShowOnlyIfSIWEIsDisabled>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const MFAConfigurationScreen: React.VFC = function MFAConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const secretConfig = useAppAndSecretConfigQuery(appID);
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
      featureConfig: featureConfig.effectiveFeatureConfig ?? undefined,
      smsProviderConfigured:
        secretConfig.secretConfig?.smsProviderSecrets?.twilioCredentials !=
          null ||
        secretConfig.secretConfig?.smsProviderSecrets
          ?.customSMSProviderCredentials != null,
      smtpConfigured: secretConfig.secretConfig?.smtpSecret != null,
      ...configForm.state,
    };
  }, [
    featureConfig.effectiveFeatureConfig,
    configForm.state,
    secretConfig.secretConfig?.smsProviderSecrets?.twilioCredentials,
    secretConfig.secretConfig?.smsProviderSecrets?.customSMSProviderCredentials,
    secretConfig.secretConfig?.smtpSecret,
  ]);

  const form: FormModel = {
    isLoading:
      configForm.isLoading || featureConfig.isLoading || secretConfig.isLoading,
    isUpdating: configForm.isUpdating,
    getIsDirty: configForm.getIsDirty,
    loadError:
      configForm.loadError ?? featureConfig.loadError ?? secretConfig.loadError,
    updateError: configForm.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      configForm.setState(() => newState);
    },
    reload: () => {
      configForm.reload();
      featureConfig.refetch().finally(() => {});
      secretConfig.refetch().finally(() => {});
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
    <FormContainer form={form} hideFooterComponent={true}>
      <MFAConfigurationContent
        form={form}
        isLoginIDEmailEnabled={isLoginIDEmailEnabled}
        isLoginIDPhoneEnabled={isLoginIDPhoneEnabled}
      />
    </FormContainer>
  );
};

export default MFAConfigurationScreen;
