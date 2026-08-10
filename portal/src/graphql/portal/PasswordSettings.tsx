import React, { useMemo, useContext, useCallback, ReactElement } from "react";
import { produce } from "immer";
import cn from "classnames";
import { Checkbox, Select, Separator, Text } from "@radix-ui/themes";
import { IBasePickerStyles } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import {
  AuthenticatorPasswordConfig,
  PasswordPolicyFeatureConfig,
  isPasswordPolicyGuessableLevel,
  passwordPolicyGuessableLevels,
  PortalAPIAppConfig,
  AccountRecoveryCodeForm,
  AccountRecoveryCodeChannel,
  AccountRecoveryChannel,
} from "../../types";
import { TextField } from "../../components/v2/TextField/TextField";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { FormField } from "../../components/v2/FormField/FormField";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import CustomTagPicker from "../../CustomTagPicker";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { useTagPickerWithNewTags } from "../../hook/useInput";
import { fixTagPickerStyles } from "../../bugs";
import {
  ensurePositiveNumber,
  parseIntegerAllowLeadingZeros,
  parseNumber,
  tryProduce,
} from "../../util/input";
import { formatDuration, parseDuration } from "../../util/duration";
import styles from "./PasswordSettings.module.css";

const excludedKeywordsTagPickerStyles: Partial<IBasePickerStyles> = {
  ...fixTagPickerStyles,
  root: {
    width: "100%",
  },
};

export enum ResetPasswordWithEmailMethod {
  Link = "link",
  Code = "code",
}

export enum ResetPasswordWithPhoneMethod {
  SMS = "sms",
  Whatsapp = "whatsapp",
  WhatsappOrSMS = "whatsapp_or_sms",
}

export function getResetPasswordWithEmailMethod(
  config: PortalAPIAppConfig
): ResetPasswordWithEmailMethod {
  const channels = config.ui?.forgot_password?.email;
  if (
    channels != null &&
    channels.length > 0 &&
    channels[0].otp_form === AccountRecoveryCodeForm.Code
  ) {
    return ResetPasswordWithEmailMethod.Code;
  }
  return ResetPasswordWithEmailMethod.Link;
}

function compareAccountRecoveryChannels(
  channels1: AccountRecoveryChannel[],
  channels2: AccountRecoveryChannel[]
): boolean {
  if (channels1.length !== channels2.length) {
    return false;
  }
  for (const [idx, c1] of channels1.entries()) {
    const c2 = channels2[idx];
    if (c1.channel !== c2.channel || c1.otp_form !== c2.otp_form) {
      return false;
    }
  }
  return true;
}

export function getResetPasswordWithPhoneMethod(
  config: PortalAPIAppConfig
): ResetPasswordWithPhoneMethod {
  const channels = config.ui?.forgot_password?.phone;
  if (channels == null) {
    return ResetPasswordWithPhoneMethod.SMS;
  }
  if (
    compareAccountRecoveryChannels(channels, [
      {
        channel: AccountRecoveryCodeChannel.Whatsapp,
        otp_form: AccountRecoveryCodeForm.Code,
      },
      {
        channel: AccountRecoveryCodeChannel.SMS,
        otp_form: AccountRecoveryCodeForm.Code,
      },
    ])
  ) {
    return ResetPasswordWithPhoneMethod.WhatsappOrSMS;
  }

  if (
    compareAccountRecoveryChannels(channels, [
      {
        channel: AccountRecoveryCodeChannel.Whatsapp,
        otp_form: AccountRecoveryCodeForm.Code,
      },
    ])
  ) {
    return ResetPasswordWithPhoneMethod.Whatsapp;
  }

  return ResetPasswordWithPhoneMethod.SMS;
}

export function setUIForgotPasswordConfig(
  config: PortalAPIAppConfig,
  options: {
    resetPasswordWithEmailBy: ResetPasswordWithEmailMethod;
    resetPasswordWithPhoneBy: ResetPasswordWithPhoneMethod;
  }
): void {
  const { resetPasswordWithEmailBy, resetPasswordWithPhoneBy } = options;
  config.ui ??= {};
  config.ui.forgot_password ??= {};
  switch (resetPasswordWithEmailBy) {
    case ResetPasswordWithEmailMethod.Code:
      config.ui.forgot_password.email = [
        {
          channel: AccountRecoveryCodeChannel.Email,
          otp_form: AccountRecoveryCodeForm.Code,
        },
      ];
      break;
    case ResetPasswordWithEmailMethod.Link:
      config.ui.forgot_password.email = [
        {
          channel: AccountRecoveryCodeChannel.Email,
          otp_form: AccountRecoveryCodeForm.Link,
        },
      ];
      break;
  }

  switch (resetPasswordWithPhoneBy) {
    case ResetPasswordWithPhoneMethod.SMS:
      config.ui.forgot_password.phone = [
        {
          channel: AccountRecoveryCodeChannel.SMS,
          otp_form: AccountRecoveryCodeForm.Code,
        },
      ];
      break;
    case ResetPasswordWithPhoneMethod.Whatsapp:
      config.ui.forgot_password.phone = [
        {
          channel: AccountRecoveryCodeChannel.Whatsapp,
          otp_form: AccountRecoveryCodeForm.Code,
        },
      ];
      break;
    case ResetPasswordWithPhoneMethod.WhatsappOrSMS:
      config.ui.forgot_password.phone = [
        {
          channel: AccountRecoveryCodeChannel.Whatsapp,
          otp_form: AccountRecoveryCodeForm.Code,
        },
        {
          channel: AccountRecoveryCodeChannel.SMS,
          otp_form: AccountRecoveryCodeForm.Code,
        },
      ];
      break;
  }
}

export interface State {
  forgotPasswordLinkValidPeriodSeconds: number | undefined;
  forgotPasswordCodeValidPeriodSeconds: number | undefined;
  resetPasswordWithEmailBy: ResetPasswordWithEmailMethod;
  resetPasswordWithPhoneBy: ResetPasswordWithPhoneMethod;
  authenticatorPasswordConfig: AuthenticatorPasswordConfig;
  passwordPolicyFeatureConfig?: PasswordPolicyFeatureConfig;
}

export interface PasswordSettingsProps<T extends State> extends State {
  className?: string;
  isLoginIDEmailEnabled: boolean;
  isLoginIDPhoneEnabled: boolean;
  setState: (fn: (state: T) => T) => void;
}

function usePasswordNumberOnChange<T extends State>(
  setState: PasswordSettingsProps<T>["setState"],
  key: "min_length" | "history_days" | "history_size"
) {
  return useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy[key] =
            parseIntegerAllowLeadingZeros(value);
        })
      );
    },
    [setState, key]
  );
}

function usePasswordCheckboxOnChange<T extends State>(
  setState: PasswordSettingsProps<T>["setState"],
  key:
    | "uppercase_required"
    | "lowercase_required"
    | "alphabet_required"
    | "digit_required"
    | "symbol_required"
) {
  return useCallback(
    (checked: boolean | "indeterminate") => {
      if (checked === "indeterminate") {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy[key] = checked;
        })
      );
    },
    [setState, key]
  );
}

function PolicyCheckbox(props: {
  label: React.ReactNode;
  checked: boolean | undefined;
  onCheckedChange: (checked: boolean | "indeterminate") => void;
}): ReactElement {
  const { label, checked, onCheckedChange } = props;
  return (
    <label className={styles.checkboxRow}>
      <Checkbox checked={checked ?? false} onCheckedChange={onCheckedChange} />
      <Text size="2">{label}</Text>
    </label>
  );
}

export default function PasswordSettings<T extends State>(
  props: PasswordSettingsProps<T>
): ReactElement {
  const {
    className,
    authenticatorPasswordConfig,
    forgotPasswordLinkValidPeriodSeconds,
    forgotPasswordCodeValidPeriodSeconds,
    resetPasswordWithEmailBy,
    resetPasswordWithPhoneBy,
    passwordPolicyFeatureConfig,
    isLoginIDEmailEnabled,
    isLoginIDPhoneEnabled,
    setState,
  } = props;

  const { renderToString } = useContext(Context);

  const anyAdvancedPolicyDisabled =
    (passwordPolicyFeatureConfig?.minimum_guessable_level?.disabled ?? false) ||
    (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
    (passwordPolicyFeatureConfig?.excluded_keywords?.disabled ?? false);

  const isPreventPasswordReuseEnabled =
    (authenticatorPasswordConfig.policy?.history_days != null &&
      authenticatorPasswordConfig.policy.history_days > 0) ||
    (authenticatorPasswordConfig.policy?.history_size != null &&
      authenticatorPasswordConfig.policy.history_size > 0);

  const isPasswordExpiryForceChangeEnabled =
    authenticatorPasswordConfig.expiry?.force_change?.enabled === true;

  const passwordExpiryForceChangeDays = useMemo(() => {
    const duration =
      authenticatorPasswordConfig.expiry?.force_change
        ?.duration_since_last_update;
    const secondsPerDay = 24 * 60 * 60;
    const days = duration ? parseDuration(duration) / secondsPerDay : undefined;

    return days;
  }, [
    authenticatorPasswordConfig.expiry?.force_change
      ?.duration_since_last_update,
  ]);

  const onChangeForceChange = useCallback(
    (checked: boolean) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.force_change = checked;
        })
      );
    },
    [setState]
  );

  const onChangeLinkExpirySeconds = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((s) =>
        produce(s, (s) => {
          s.forgotPasswordLinkValidPeriodSeconds = tryProduce(
            s.forgotPasswordLinkValidPeriodSeconds,
            () => {
              const num = parseNumber(value);
              return num == null ? undefined : ensurePositiveNumber(num);
            }
          );
        })
      );
    },
    [setState]
  );

  const onChangeCodeExpirySeconds = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((s) =>
        produce(s, (s) => {
          s.forgotPasswordCodeValidPeriodSeconds = tryProduce(
            s.forgotPasswordCodeValidPeriodSeconds,
            () => {
              const num = parseNumber(value);
              return num == null ? undefined : ensurePositiveNumber(num);
            }
          );
        })
      );
    },
    [setState]
  );

  const onChangeExpiryForceChangeDays = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((s) =>
        produce(s, (s) => {
          s.authenticatorPasswordConfig.expiry ??= {};
          s.authenticatorPasswordConfig.expiry.force_change ??= {};
          s.authenticatorPasswordConfig.expiry.force_change.duration_since_last_update =
            tryProduce(
              s.authenticatorPasswordConfig.expiry.force_change
                .duration_since_last_update,
              () => {
                const num = parseNumber(value);
                return num == null ? undefined : formatDuration(num * 24, "h");
              }
            );
        })
      );
    },
    [setState]
  );

  const onBlurExpiryForceChangeDays = useCallback(() => {
    setState((s) =>
      produce(s, (s) => {
        if (!passwordExpiryForceChangeDays) {
          s.authenticatorPasswordConfig.expiry = undefined;
        }
      })
    );
  }, [passwordExpiryForceChangeDays, setState]);

  const onChangeMinLength = usePasswordNumberOnChange(setState, "min_length");
  const onChangeDigitRequired = usePasswordCheckboxOnChange(
    setState,
    "digit_required"
  );
  const onChangeLowercaseRequired = usePasswordCheckboxOnChange(
    setState,
    "lowercase_required"
  );
  const onChangeUppercaseRequired = usePasswordCheckboxOnChange(
    setState,
    "uppercase_required"
  );
  const onChangeAlphabetRequired = usePasswordCheckboxOnChange(
    setState,
    "alphabet_required"
  );
  const onChangeSymbolRequired = usePasswordCheckboxOnChange(
    setState,
    "symbol_required"
  );
  const onChangeHistoryDays = usePasswordNumberOnChange(
    setState,
    "history_days"
  );
  const onChangeHistorySize = usePasswordNumberOnChange(
    setState,
    "history_size"
  );

  const minGuessableLevelOptions = useMemo(() => {
    return passwordPolicyGuessableLevels.map((level) => ({
      value: String(level),
      label: renderToString(
        `PasswordPolicyConfigurationScreen.min-guessable-level.${level}`
      ),
    }));
  }, [renderToString]);

  const onChangeMinimumGuessableLevel = useCallback(
    (value: string) => {
      const key = Number(value);
      if (!isPasswordPolicyGuessableLevel(key)) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy.minimum_guessable_level = key;
        })
      );
    },
    [setState]
  );

  const onChangePreventReuseEnabled = useCallback(
    (checked: boolean) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          if (checked) {
            prev.authenticatorPasswordConfig.policy.history_days = 90;
            prev.authenticatorPasswordConfig.policy.history_size = 3;
          } else {
            prev.authenticatorPasswordConfig.policy.history_days = 0;
            prev.authenticatorPasswordConfig.policy.history_size = 0;
          }
        })
      );
    },
    [setState]
  );

  const onChangeExpiryForceChangeEnabled = useCallback(
    (checked: boolean) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.expiry ??= {};
          prev.authenticatorPasswordConfig.expiry.force_change ??= {};

          if (checked) {
            prev.authenticatorPasswordConfig.expiry.force_change.enabled = true;
            prev.authenticatorPasswordConfig.expiry.force_change.duration_since_last_update =
              formatDuration(90 * 24, "h");
          } else {
            prev.authenticatorPasswordConfig.expiry = undefined;
          }
        })
      );
    },
    [setState]
  );

  const valueForExcludedKeywords = useMemo(() => {
    return authenticatorPasswordConfig.policy?.excluded_keywords ?? [];
  }, [authenticatorPasswordConfig.policy?.excluded_keywords]);
  const updateExcludedKeywords = useCallback(
    (value: string[]) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy.excluded_keywords = value;
        })
      );
    },
    [setState]
  );
  const {
    selectedItems: excludedKeywords,
    onChange: onChangeExcludedKeywords,
    onResolveSuggestions: onResolveSuggestionsExcludedKeywords,
    onAdd: onAddExcludedKeywords,
  } = useTagPickerWithNewTags(valueForExcludedKeywords, updateExcludedKeywords);

  const resetPasswordWithEmailOptions = useMemo(
    () => [
      {
        value: ResetPasswordWithEmailMethod.Link,
        label: renderToString(
          "PasswordSettings.resetPasswordWithEmail.options.link"
        ),
      },
      {
        value: ResetPasswordWithEmailMethod.Code,
        label: renderToString(
          "PasswordSettings.resetPasswordWithEmail.options.code"
        ),
      },
    ],
    [renderToString]
  );

  const resetPasswordWithPhoneOptions = useMemo(
    () => [
      {
        value: ResetPasswordWithPhoneMethod.SMS,
        label: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.sms"
        ),
      },
      {
        value: ResetPasswordWithPhoneMethod.Whatsapp,
        label: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.whatsapp"
        ),
      },
      {
        value: ResetPasswordWithPhoneMethod.WhatsappOrSMS,
        label: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.whatsappOrSMS"
        ),
      },
    ],
    [renderToString]
  );

  const onChangeResetPasswordWithEmail = useCallback(
    (value: string) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.resetPasswordWithEmailBy = value as ResetPasswordWithEmailMethod;
        })
      );
    },
    [setState]
  );

  const onChangeResetPasswordWithPhone = useCallback(
    (value: string) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.resetPasswordWithPhoneBy = value as ResetPasswordWithPhoneMethod;
        })
      );
    },
    [setState]
  );

  return (
    <SettingsSectionCard
      className={className}
      contentClassName="gap-4"
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.password.title" />
      }
    >
      <div className={styles.section}>
        <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="LoginMethodConfigurationScreen.password.resetPassword.title" />
        </Text>
        <FormField
          size="2"
          labelSize="2"
          labelSpace="1"
          label={
            <FormattedMessage id="PasswordSettings.resetPasswordWithEmail.label" />
          }
        >
          <Select.Root
            value={resetPasswordWithEmailBy}
            onValueChange={onChangeResetPasswordWithEmail}
            disabled={!isLoginIDEmailEnabled}
            size="2"
          >
            <Select.Trigger
              variant="surface"
              className={styles.selectTrigger}
            />
            <Select.Content>
              {resetPasswordWithEmailOptions.map((opt) => (
                <Select.Item key={opt.value} value={opt.value}>
                  {opt.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
        <FormField
          size="2"
          labelSize="2"
          labelSpace="1"
          label={
            <FormattedMessage id="PasswordSettings.resetPasswordWithPhone.label" />
          }
        >
          <Select.Root
            value={resetPasswordWithPhoneBy}
            onValueChange={onChangeResetPasswordWithPhone}
            disabled={!isLoginIDPhoneEnabled}
            size="2"
          >
            <Select.Trigger
              variant="surface"
              className={styles.selectTrigger}
            />
            <Select.Content>
              {resetPasswordWithPhoneOptions.map((opt) => (
                <Select.Item key={opt.value} value={opt.value}>
                  {opt.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
        <TextField
          size="2"
          labelSize="2"
          type="text"
          label={
            <FormattedMessage id="PasswordSettings.reset-link-valid-duration.label" />
          }
          value={forgotPasswordLinkValidPeriodSeconds?.toFixed(0) ?? ""}
          onChange={onChangeLinkExpirySeconds}
          disabled={!isLoginIDEmailEnabled}
        />
        <TextField
          size="2"
          labelSize="2"
          type="text"
          label={
            <FormattedMessage id="PasswordSettings.reset-code-valid-duration.label" />
          }
          value={forgotPasswordCodeValidPeriodSeconds?.toFixed(0) ?? ""}
          onChange={onChangeCodeExpirySeconds}
          disabled={!(isLoginIDEmailEnabled || isLoginIDPhoneEnabled)}
        />
      </div>

      <Separator size="4" className={styles.separator} />

      <div className={styles.section}>
        <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="LoginMethodConfigurationScreen.password.requirements" />
        </Text>
        <Text
          as="p"
          size="2"
          color="gray"
          className={styles.sectionDescription}
        >
          <FormattedMessage id="LoginMethodConfigurationScreen.password.description" />
        </Text>
        <Toggle
          checked={authenticatorPasswordConfig.force_change ?? false}
          onCheckedChange={onChangeForceChange}
          textWeight="medium"
          text={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.force-change.label" />
          }
        />
        <TextField
          size="2"
          labelSize="2"
          type="text"
          label={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.min-length.label" />
          }
          value={
            authenticatorPasswordConfig.policy?.min_length?.toFixed(0) ?? ""
          }
          onChange={onChangeMinLength}
        />
        <div className={styles.checkboxGroup}>
          <PolicyCheckbox
            label={
              <FormattedMessage id="PasswordPolicyConfigurationScreen.require-digit.label" />
            }
            checked={authenticatorPasswordConfig.policy?.digit_required}
            onCheckedChange={onChangeDigitRequired}
          />
          <PolicyCheckbox
            label={
              <FormattedMessage id="PasswordPolicyConfigurationScreen.require-lowercase.label" />
            }
            checked={authenticatorPasswordConfig.policy?.lowercase_required}
            onCheckedChange={onChangeLowercaseRequired}
          />
          <PolicyCheckbox
            label={
              <FormattedMessage id="PasswordPolicyConfigurationScreen.require-uppercase.label" />
            }
            checked={authenticatorPasswordConfig.policy?.uppercase_required}
            onCheckedChange={onChangeUppercaseRequired}
          />
          <PolicyCheckbox
            label={
              <FormattedMessage id="PasswordPolicyConfigurationScreen.require-alphabet.label" />
            }
            checked={authenticatorPasswordConfig.policy?.alphabet_required}
            onCheckedChange={onChangeAlphabetRequired}
          />
          <PolicyCheckbox
            label={
              <FormattedMessage id="PasswordPolicyConfigurationScreen.require-symbol.label" />
            }
            checked={authenticatorPasswordConfig.policy?.symbol_required}
            onCheckedChange={onChangeSymbolRequired}
          />
        </div>
      </div>

      <Separator size="4" className={styles.separator} />

      <div className={styles.section}>
        <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="LoginMethodConfigurationScreen.password.requirements.advanced" />
        </Text>
        {anyAdvancedPolicyDisabled ? (
          <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
        ) : null}
        <FormField
          size="2"
          labelSize="2"
          labelSpace="1"
          label={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.min-guessable-level.label" />
          }
        >
          <Select.Root
            value={String(
              authenticatorPasswordConfig.policy?.minimum_guessable_level ?? 0
            )}
            onValueChange={onChangeMinimumGuessableLevel}
            disabled={
              passwordPolicyFeatureConfig?.minimum_guessable_level?.disabled
            }
            size="2"
          >
            <Select.Trigger
              variant="surface"
              className={styles.selectTrigger}
            />
            <Select.Content>
              {minGuessableLevelOptions.map((opt) => (
                <Select.Item key={opt.value} value={opt.value}>
                  {opt.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </FormField>
        <Toggle
          disabled={passwordPolicyFeatureConfig?.history?.disabled}
          checked={isPreventPasswordReuseEnabled}
          onCheckedChange={onChangePreventReuseEnabled}
          textWeight="medium"
          text={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.prevent-reuse.label" />
          }
        />
        <TextField
          size="2"
          labelSize="2"
          type="text"
          disabled={
            (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
            !isPreventPasswordReuseEnabled
          }
          label={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.history-days.label" />
          }
          value={
            authenticatorPasswordConfig.policy?.history_days?.toFixed(0) ?? ""
          }
          onChange={onChangeHistoryDays}
        />
        <TextField
          size="2"
          labelSize="2"
          type="text"
          disabled={
            (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
            !isPreventPasswordReuseEnabled
          }
          label={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.history-size.label" />
          }
          value={
            authenticatorPasswordConfig.policy?.history_size?.toFixed(0) ?? ""
          }
          onChange={onChangeHistorySize}
        />
      </div>

      <Separator size="4" className={styles.separator} />

      <div className={styles.section}>
        <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry" />
        </Text>
        <Text
          as="p"
          size="2"
          color="gray"
          className={styles.sectionDescription}
        >
          <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry.description" />
        </Text>
        <Toggle
          checked={isPasswordExpiryForceChangeEnabled}
          onCheckedChange={onChangeExpiryForceChangeEnabled}
          textWeight="medium"
          text={
            <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry.enable-force-change.label" />
          }
        />
        <TextField
          size="2"
          labelSize="2"
          type="number"
          disabled={!isPasswordExpiryForceChangeEnabled}
          label={
            <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry.force-change-since-last-update.label" />
          }
          value={passwordExpiryForceChangeDays?.toFixed(0) ?? ""}
          onChange={onChangeExpiryForceChangeDays}
          onBlur={onBlurExpiryForceChangeDays}
        />
      </div>

      <Separator size="4" className={styles.separator} />

      <div
        className={cn(
          passwordPolicyFeatureConfig?.excluded_keywords?.disabled &&
            styles.sectionDisabled
        )}
      >
        <FormField
          size="2"
          labelSize="2"
          labelSpace="1"
          label={
            <FormattedMessage id="PasswordPolicyConfigurationScreen.excluded-keywords.label" />
          }
        >
          <div className={styles.excludedKeywordsTagPicker}>
            <CustomTagPicker
              styles={excludedKeywordsTagPickerStyles}
              inputProps={{
                "aria-label": renderToString(
                  "PasswordPolicyConfigurationScreen.excluded-keywords.label"
                ),
              }}
              disabled={
                passwordPolicyFeatureConfig?.excluded_keywords?.disabled
              }
              selectedItems={excludedKeywords}
              onChange={onChangeExcludedKeywords}
              onResolveSuggestions={onResolveSuggestionsExcludedKeywords}
              onAdd={onAddExcludedKeywords}
            />
          </div>
        </FormField>
      </div>
    </SettingsSectionCard>
  );
}
