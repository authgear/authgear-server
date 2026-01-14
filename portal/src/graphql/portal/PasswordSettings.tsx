import React, { useMemo, useContext, useCallback, ReactElement } from "react";
import { produce } from "immer";
import {
  Checkbox,
  Dropdown,
  IDropdownOption,
  IDropdownProps,
  Label,
} from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import {
  AuthenticatorPasswordConfig,
  PasswordPolicyFeatureConfig,
  PasswordPolicyConfig,
  isPasswordPolicyGuessableLevel,
  passwordPolicyGuessableLevels,
  PortalAPIAppConfig,
  AccountRecoveryCodeForm,
  AccountRecoveryCodeChannel,
  AccountRecoveryChannel,
} from "../../types";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetSubtitle from "../../WidgetSubtitle";
import WidgetDescription from "../../WidgetDescription";
import HorizontalDivider from "../../HorizontalDivider";
import TextField from "../../TextField";
import Toggle from "../../Toggle";
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

function useResetPasswordWithEmailDropdown<T extends State>(
  setState: PasswordSettingsProps<T>["setState"]
): {
  options: IDropdownOption<ResetPasswordWithEmailMethod>[];
  onChange: IDropdownProps["onChange"];
} {
  const { renderToString } = useContext(Context);
  const options: IDropdownOption<ResetPasswordWithEmailMethod>[] = useMemo(
    () => [
      {
        key: ResetPasswordWithEmailMethod.Link,
        text: renderToString(
          "PasswordSettings.resetPasswordWithEmail.options.link"
        ),
      },
      {
        key: ResetPasswordWithEmailMethod.Code,
        text: renderToString(
          "PasswordSettings.resetPasswordWithEmail.options.code"
        ),
      },
    ],
    [renderToString]
  );

  const onChange = useCallback(
    (_: unknown, option?: IDropdownOption<ResetPasswordWithEmailMethod>) => {
      const key = option?.key as ResetPasswordWithEmailMethod;
      setState((prev) =>
        produce(prev, (prev) => {
          prev.resetPasswordWithEmailBy = key;
        })
      );
    },
    [setState]
  );

  return {
    options,
    onChange,
  };
}

function useResetPasswordWithPhoneDropdown<T extends State>(
  setState: PasswordSettingsProps<T>["setState"]
): {
  options: IDropdownOption<ResetPasswordWithPhoneMethod>[];
  onChange: IDropdownProps["onChange"];
} {
  const { renderToString } = useContext(Context);
  const options: IDropdownOption<ResetPasswordWithPhoneMethod>[] = useMemo(
    () => [
      {
        key: ResetPasswordWithPhoneMethod.SMS,
        text: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.sms"
        ),
      },
      {
        key: ResetPasswordWithPhoneMethod.Whatsapp,
        text: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.whatsapp"
        ),
      },
      {
        key: ResetPasswordWithPhoneMethod.WhatsappOrSMS,
        text: renderToString(
          "PasswordSettings.resetPasswordWithPhone.options.whatsappOrSMS"
        ),
      },
    ],
    [renderToString]
  );

  const onChange = useCallback(
    (_: unknown, option?: IDropdownOption<ResetPasswordWithPhoneMethod>) => {
      const key = option?.key as ResetPasswordWithPhoneMethod;
      setState((prev) =>
        produce(prev, (prev) => {
          prev.resetPasswordWithPhoneBy = key;
        })
      );
    },
    [setState]
  );

  return {
    options,
    onChange,
  };
}

function usePasswordNumberOnChange<T extends State>(
  setState: PasswordSettingsProps<T>["setState"],
  key: "min_length" | "history_days" | "history_size"
) {
  return useCallback(
    (_, value) => {
      if (value == null) {
        return;
      }
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
  key: keyof PasswordPolicyConfig
) {
  return useCallback(
    (_, value) => {
      if (value == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy[key] = value;
        })
      );
    },
    [setState, key]
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
    (_e, checked) => {
      if (checked == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.force_change = checked;
        })
      );
    },
    [setState]
  );

  const onChangeLinkExpirySeconds = useCallback(
    (_e, value: string | undefined) => {
      if (value == null) {
        return;
      }
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
    (_e, value: string | undefined) => {
      if (value == null) {
        return;
      }
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
    (_e, value: string | undefined) => {
      if (value == null) {
        return;
      }
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
      key: level,
      text: renderToString(
        `PasswordPolicyConfigurationScreen.min-guessable-level.${level}`
      ),
    }));
  }, [renderToString]);
  const onChangeMinimumGuessableLevel = useCallback(
    (_, option) => {
      const key = option?.key;
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
    (_, checked) => {
      if (checked == null) {
        return;
      }
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
    (_, checked) => {
      if (checked == null) {
        return;
      }
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

  const resetPasswordWithEmailDropdown =
    useResetPasswordWithEmailDropdown(setState);
  const resetPasswordWithPhoneDropdown =
    useResetPasswordWithPhoneDropdown(setState);

  return (
    <Widget className={className}>
      <WidgetTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.title" />
      </WidgetTitle>
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.resetPassword.title" />
      </WidgetSubtitle>
      <Dropdown
        label={renderToString("PasswordSettings.resetPasswordWithEmail.label")}
        options={resetPasswordWithEmailDropdown.options}
        selectedKey={resetPasswordWithEmailBy}
        onChange={resetPasswordWithEmailDropdown.onChange}
        disabled={!isLoginIDEmailEnabled}
      />
      <Dropdown
        label={renderToString("PasswordSettings.resetPasswordWithPhone.label")}
        options={resetPasswordWithPhoneDropdown.options}
        selectedKey={resetPasswordWithPhoneBy}
        onChange={resetPasswordWithPhoneDropdown.onChange}
        disabled={!isLoginIDPhoneEnabled}
      />
      <TextField
        type="text"
        label={renderToString(
          "PasswordSettings.reset-link-valid-duration.label"
        )}
        value={forgotPasswordLinkValidPeriodSeconds?.toFixed(0) ?? ""}
        onChange={onChangeLinkExpirySeconds}
        disabled={!isLoginIDEmailEnabled}
      />
      <TextField
        type="text"
        label={renderToString(
          "PasswordSettings.reset-code-valid-duration.label"
        )}
        value={forgotPasswordCodeValidPeriodSeconds?.toFixed(0) ?? ""}
        onChange={onChangeCodeExpirySeconds}
        disabled={!(isLoginIDEmailEnabled || isLoginIDPhoneEnabled)}
      />
      <HorizontalDivider />
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.requirements" />
      </WidgetSubtitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.description" />
      </WidgetDescription>
      <Toggle
        checked={authenticatorPasswordConfig.force_change}
        inlineLabel={true}
        label={
          <FormattedMessage id="PasswordPolicyConfigurationScreen.force-change.label" />
        }
        onChange={onChangeForceChange}
      />
      <TextField
        type="text"
        label={renderToString(
          "PasswordPolicyConfigurationScreen.min-length.label"
        )}
        value={authenticatorPasswordConfig.policy?.min_length?.toFixed(0) ?? ""}
        onChange={onChangeMinLength}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-digit.label"
        )}
        checked={authenticatorPasswordConfig.policy?.digit_required}
        onChange={onChangeDigitRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-lowercase.label"
        )}
        checked={authenticatorPasswordConfig.policy?.lowercase_required}
        onChange={onChangeLowercaseRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-uppercase.label"
        )}
        checked={authenticatorPasswordConfig.policy?.uppercase_required}
        onChange={onChangeUppercaseRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-alphabet.label"
        )}
        checked={authenticatorPasswordConfig.policy?.alphabet_required}
        onChange={onChangeAlphabetRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-symbol.label"
        )}
        checked={authenticatorPasswordConfig.policy?.symbol_required}
        onChange={onChangeSymbolRequired}
      />
      <HorizontalDivider />
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.requirements.advanced" />
      </WidgetSubtitle>
      {anyAdvancedPolicyDisabled ? (
        <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
      ) : null}
      <Dropdown
        label={renderToString(
          "PasswordPolicyConfigurationScreen.min-guessable-level.label"
        )}
        disabled={
          passwordPolicyFeatureConfig?.minimum_guessable_level?.disabled
        }
        options={minGuessableLevelOptions}
        selectedKey={
          authenticatorPasswordConfig.policy?.minimum_guessable_level
        }
        onChange={onChangeMinimumGuessableLevel}
      />
      <Toggle
        disabled={passwordPolicyFeatureConfig?.history?.disabled}
        checked={isPreventPasswordReuseEnabled}
        inlineLabel={true}
        label={
          <FormattedMessage id="PasswordPolicyConfigurationScreen.prevent-reuse.label" />
        }
        onChange={onChangePreventReuseEnabled}
      />
      <TextField
        type="text"
        disabled={
          (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
          !isPreventPasswordReuseEnabled
        }
        label={renderToString(
          "PasswordPolicyConfigurationScreen.history-days.label"
        )}
        value={
          authenticatorPasswordConfig.policy?.history_days?.toFixed(0) ?? ""
        }
        onChange={onChangeHistoryDays}
      />
      <TextField
        type="text"
        disabled={
          (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
          !isPreventPasswordReuseEnabled
        }
        label={renderToString(
          "PasswordPolicyConfigurationScreen.history-size.label"
        )}
        value={
          authenticatorPasswordConfig.policy?.history_size?.toFixed(0) ?? ""
        }
        onChange={onChangeHistorySize}
      />
      <HorizontalDivider />
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry" />
      </WidgetSubtitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry.description" />
      </WidgetDescription>
      <Toggle
        checked={isPasswordExpiryForceChangeEnabled}
        inlineLabel={true}
        label={
          <FormattedMessage id="LoginMethodConfigurationScreen.password.expiry.enable-force-change.label" />
        }
        onChange={onChangeExpiryForceChangeEnabled}
      />
      <TextField
        type="number"
        min={0}
        disabled={!isPasswordExpiryForceChangeEnabled}
        label={renderToString(
          "LoginMethodConfigurationScreen.password.expiry.force-change-since-last-update.label"
        )}
        value={passwordExpiryForceChangeDays?.toFixed(0) ?? ""}
        onChange={onChangeExpiryForceChangeDays}
        onBlur={onBlurExpiryForceChangeDays}
      />
      <HorizontalDivider />
      <div>
        <Label
          disabled={passwordPolicyFeatureConfig?.excluded_keywords?.disabled}
        >
          <FormattedMessage id="PasswordPolicyConfigurationScreen.excluded-keywords.label" />
        </Label>
        <CustomTagPicker
          styles={fixTagPickerStyles}
          inputProps={{
            "aria-label": renderToString(
              "PasswordPolicyConfigurationScreen.excluded-keywords.label"
            ),
          }}
          disabled={passwordPolicyFeatureConfig?.excluded_keywords?.disabled}
          selectedItems={excludedKeywords}
          onChange={onChangeExcludedKeywords}
          onResolveSuggestions={onResolveSuggestionsExcludedKeywords}
          onAdd={onAddExcludedKeywords}
        />
      </div>
    </Widget>
  );
}
