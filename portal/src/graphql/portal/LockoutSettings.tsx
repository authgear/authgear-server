/* eslint-disable @typescript-eslint/no-unnecessary-type-parameters */

import React, { ReactElement, useMemo, useContext, useCallback } from "react";
import cn from "classnames";
import { Checkbox, Flex, RadioGroup, Text } from "@radix-ui/themes";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { TextField } from "../../components/v2/TextField/TextField";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { AuthenticationLockoutType } from "../../types";
import { produce } from "immer";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./LockoutSettings.module.css";
import {
  ErrorParseRule,
  makeValidationErrorCustomMessageIDRule,
} from "../../error/parse";

export interface State {
  isEnabled: boolean;
  maxAttempts?: number;
  historyDurationMins?: number;
  minimumDurationMins?: number;
  maximumDurationMins?: number;
  backoffFactorRaw?: string;
  lockoutType: AuthenticationLockoutType;
  isEnabledForPassword: boolean;
  isEnabledForTOTP: boolean;
  isEnabledForOOBOTP: boolean;
  isEnabledForRecoveryCode: boolean;
}

export interface LockoutSettingsProps<T extends State> extends State {
  className?: string;
  setState: (fn: (state: T) => T) => void;
}

interface LabeledCheckboxProps {
  label?: React.ReactNode;
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
}

function LabeledCheckbox(props: LabeledCheckboxProps) {
  const { label, checked, onCheckedChange } = props;
  const handleCheckedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      if (checked === "indeterminate") {
        return;
      }
      onCheckedChange?.(checked);
    },
    [onCheckedChange]
  );
  return (
    <label className={styles.checkboxRow}>
      <Checkbox
        checked={checked ?? false}
        onCheckedChange={handleCheckedChange}
      />
      <Text size="2">{label}</Text>
    </label>
  );
}

function useIntegerOnChange<T extends State>(
  setState: LockoutSettingsProps<T>["setState"],
  key:
    | "maxAttempts"
    | "historyDurationMins"
    | "minimumDurationMins"
    | "maximumDurationMins"
) {
  return useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setState((prev) =>
        produce(prev, (prev) => {
          prev[key] = parseIntegerAllowLeadingZeros(value);
        })
      );
    },
    [setState, key]
  );
}

function useDecimalOnChange<T extends State>(
  setState: LockoutSettingsProps<T>["setState"],
  key: "backoffFactorRaw"
) {
  return useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      const newNumber = Number(value);
      if (!Number.isFinite(newNumber)) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev[key] = value;
        })
      );
    },
    [setState, key]
  );
}

function useBooleanOnChange<T extends State>(
  setState: LockoutSettingsProps<T>["setState"],
  key:
    | "isEnabled"
    | "isEnabledForPassword"
    | "isEnabledForTOTP"
    | "isEnabledForOOBOTP"
    | "isEnabledForRecoveryCode"
) {
  return useCallback(
    (value: boolean) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev[key] = value;
        })
      );
    },
    [setState, key]
  );
}

function formatOptionalHour(hour: number): {
  isDisplayed: "true" | "false";
  value: number;
} {
  if (hour < 1) {
    return {
      isDisplayed: "false",
      value: hour,
    };
  }
  return { isDisplayed: "true", value: Number(hour.toFixed(2)) };
}

function LockoutThresholdSection<T extends State>(props: {
  className?: string;
  state: T;
  onMaxAttemptsChange: React.ChangeEventHandler<HTMLInputElement>;
  onHistoryDurationMinsChange: React.ChangeEventHandler<HTMLInputElement>;
}) {
  const { renderToString } = useContext(MessageContext);
  const { className, state, onHistoryDurationMinsChange, onMaxAttemptsChange } =
    props;
  const overallDescValues = useMemo(() => {
    const hours = formatOptionalHour((state.historyDurationMins ?? 0) / 60);
    return {
      attempts: state.maxAttempts ?? 0,
      resetIntervalMins: state.historyDurationMins ?? 0,
      resetIntervalHoursDisplayed: hours.isDisplayed,
      resetIntervalHours: hours.value,
    };
  }, [state.historyDurationMins, state.maxAttempts]);

  return (
    <SettingsSectionCard
      className={className}
      contentClassName="gap-4"
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.title" />
      }
      description={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.description" />
      }
    >
      <TextField
        size="2"
        labelSize="2"
        fieldName="max_attempts"
        parentJSONPointer="/authentication/lockout"
        type="text"
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.threshold.failedAttempts.title"
        )}
        value={state.maxAttempts?.toFixed(0) ?? ""}
        onChange={onMaxAttemptsChange}
        hint={
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.failedAttempts.description" />
        }
      />
      <TextField
        size="2"
        labelSize="2"
        fieldName="history_duration"
        parentJSONPointer="/authentication/lockout"
        type="text"
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.threshold.resetAfter.title"
        )}
        value={state.historyDurationMins?.toFixed(0) ?? ""}
        onChange={onHistoryDurationMinsChange}
        hint={
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.resetAfter.description" />
        }
      />
      <div className={styles.descriptionBox}>
        <Text size="2">
          <FormattedMessage
            id="LoginMethodConfigurationScreen.lockout.threshold.overall.description"
            values={overallDescValues}
          />
        </Text>
      </div>
    </SettingsSectionCard>
  );
}

const minDurationErrorParseRules: ErrorParseRule[] = [
  makeValidationErrorCustomMessageIDRule(
    "maximum",
    /\/authentication\/lockout\/minimum_duration/,
    "LoginMethodConfigurationScreen.lockout.errors.maxDurationMustBeGreaterThanMinDuration"
  ),
];

function LockoutDurationSection<T extends State>(props: {
  className?: string;
  state: T;
  onMinDurationChange: React.ChangeEventHandler<HTMLInputElement>;
  onBackoffFactorChange: React.ChangeEventHandler<HTMLInputElement>;
  onMaximumDurationMinsChange: React.ChangeEventHandler<HTMLInputElement>;
}) {
  const { renderToString } = useContext(MessageContext);
  const {
    className,
    state,
    onBackoffFactorChange,
    onMaximumDurationMinsChange,
    onMinDurationChange,
  } = props;

  const overallDescriptionValues = useMemo(() => {
    const durationMins = state.minimumDurationMins ?? 0;
    let backoffFactor = Number(state.backoffFactorRaw);
    if (!Number.isFinite(backoffFactor)) {
      backoffFactor = 1;
    }
    const durationMinsSecond = Number(
      (durationMins * backoffFactor).toFixed(2)
    );
    const durationMinsThird = Number(
      (durationMins * backoffFactor * backoffFactor).toFixed(2)
    );
    const maxDurationMins = state.maximumDurationMins ?? 0;
    const maxDurationHours = formatOptionalHour(maxDurationMins / 60);
    return {
      durationMins,
      backoffFactor,
      durationMinsSecond,
      durationMinsThird,
      maxDurationMins,
      maxDurationHours: maxDurationHours.value,
      maxDurationHoursDisplayed: maxDurationHours.isDisplayed,
    };
  }, [state]);

  return (
    <SettingsSectionCard
      className={className}
      contentClassName="gap-4"
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.title" />
      }
      description={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.description" />
      }
    >
      <TextField
        size="2"
        labelSize="2"
        fieldName="minimum_duration"
        parentJSONPointer="/authentication/lockout"
        type="text"
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.duration.duration.title"
        )}
        value={state.minimumDurationMins?.toFixed(0) ?? ""}
        onChange={onMinDurationChange}
        errorRules={minDurationErrorParseRules}
        hint={
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.duration.description" />
        }
      />
      <TextField
        size="2"
        labelSize="2"
        fieldName="backoff_factor"
        parentJSONPointer="/authentication/lockout"
        type="text"
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.duration.backoff.title"
        )}
        value={state.backoffFactorRaw ?? ""}
        onChange={onBackoffFactorChange}
        hint={
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.backoff.description" />
        }
      />
      <TextField
        size="2"
        labelSize="2"
        fieldName="maximum_duration"
        parentJSONPointer="/authentication/lockout"
        type="text"
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.duration.max.title"
        )}
        value={state.maximumDurationMins?.toFixed(0) ?? ""}
        onChange={onMaximumDurationMinsChange}
        hint={
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.max.description" />
        }
      />
      <div className={styles.descriptionBox}>
        <Text size="2">
          {Number(state.backoffFactorRaw) <= 1 ? (
            <FormattedMessage
              id="LoginMethodConfigurationScreen.lockout.duration.overall.description.noBackoff"
              values={{
                durationMins: state.minimumDurationMins ?? 0,
              }}
            />
          ) : (
            <FormattedMessage
              id="LoginMethodConfigurationScreen.lockout.duration.overall.description.withBackoff"
              values={overallDescriptionValues}
            />
          )}
        </Text>
      </div>
    </SettingsSectionCard>
  );
}

function LockoutTypeSection<T extends State>(props: {
  className?: string;
  state: T;
  onChangeLockoutType: (value: string) => void;
}) {
  const { className, state, onChangeLockoutType } = props;
  const { renderToString } = useContext(MessageContext);

  const lockoutTypeOptions = useMemo(() => {
    return [
      {
        key: "per_user",
        text: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUser"
        ),
        description: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUser.description"
        ),
      },
      {
        key: "per_user_per_ip",
        text: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUserPerIP"
        ),
        description: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUserPerIP.description"
        ),
      },
    ];
  }, [renderToString]);

  return (
    <SettingsSectionCard
      className={className}
      contentClassName="gap-4"
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.type.title" />
      }
      description={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.type.description" />
      }
    >
      <RadioGroup.Root
        value={state.lockoutType}
        onValueChange={onChangeLockoutType}
      >
        <Flex direction="column" gap="3">
          {lockoutTypeOptions.map((option) => (
            <Text
              key={option.key}
              as="label"
              size="2"
              className={styles.radioOption}
            >
              <Flex gap="2">
                <RadioGroup.Item value={option.key} />
                <Flex direction="column" gap="1">
                  <span>{option.text}</span>
                  <Text size="2" color="gray">
                    {option.description}
                  </Text>
                </Flex>
              </Flex>
            </Text>
          ))}
        </Flex>
      </RadioGroup.Root>
    </SettingsSectionCard>
  );
}

function LockoutAuthenticatorSection<T extends State>(props: {
  className?: string;
  state: T;
  onChangeIsEnabledForPassword: (checked: boolean) => void;
  onChangeIsEnabledForOOBOTP: (checked: boolean) => void;
  onChangeIsEnabledForTOTP: (checked: boolean) => void;
  onChangeIsEnabledForRecoveryCode: (checked: boolean) => void;
}) {
  const {
    className,
    state,
    onChangeIsEnabledForPassword,
    onChangeIsEnabledForOOBOTP,
    onChangeIsEnabledForTOTP,
    onChangeIsEnabledForRecoveryCode,
  } = props;
  const { renderToString } = useContext(MessageContext);

  return (
    <SettingsSectionCard
      className={className}
      contentClassName="gap-4"
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.authenticator.title" />
      }
      description={
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.authenticator.description" />
      }
    >
      <div className={styles.checkboxGroup}>
        <LabeledCheckbox
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.authenticator.password"
          )}
          checked={state.isEnabledForPassword}
          onCheckedChange={onChangeIsEnabledForPassword}
        />
        <LabeledCheckbox
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.authenticator.passwordless"
          )}
          checked={state.isEnabledForOOBOTP}
          onCheckedChange={onChangeIsEnabledForOOBOTP}
        />
        <LabeledCheckbox
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.authenticator.totp"
          )}
          checked={state.isEnabledForTOTP}
          onCheckedChange={onChangeIsEnabledForTOTP}
        />
        <LabeledCheckbox
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.authenticator.recoveryCode"
          )}
          checked={state.isEnabledForRecoveryCode}
          onCheckedChange={onChangeIsEnabledForRecoveryCode}
        />
      </div>
    </SettingsSectionCard>
  );
}

export default function LockoutSettings<T extends State>(
  props: LockoutSettingsProps<T>
): ReactElement {
  const { className, setState, ...state } = props;

  const { renderToString } = useContext(MessageContext);

  const onChangeIsEnabled = useBooleanOnChange(setState, "isEnabled");

  const onMaxAttemptsChange = useIntegerOnChange(setState, "maxAttempts");
  const onHistoryDurationMinsChange = useIntegerOnChange(
    setState,
    "historyDurationMins"
  );
  const onMinDurationChange = useIntegerOnChange(
    setState,
    "minimumDurationMins"
  );
  const onBackoffFactorChange = useDecimalOnChange(
    setState,
    "backoffFactorRaw"
  );
  const onMaximumDurationMinsChange = useIntegerOnChange(
    setState,
    "maximumDurationMins"
  );
  const onChangeLockoutType = useCallback(
    (value: string) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.lockoutType = value as AuthenticationLockoutType;
        })
      );
    },
    [setState]
  );
  const onChangeIsEnabledForPassword = useBooleanOnChange(
    setState,
    "isEnabledForPassword"
  );
  const onChangeIsEnabledForOOBOTP = useBooleanOnChange(
    setState,
    "isEnabledForOOBOTP"
  );
  const onChangeIsEnabledForTOTP = useBooleanOnChange(
    setState,
    "isEnabledForTOTP"
  );
  const onChangeIsEnabledForRecoveryCode = useBooleanOnChange(
    setState,
    "isEnabledForRecoveryCode"
  );

  return (
    <>
      <SettingsSectionCard
        className={cn(className, styles.settingsCardAlignCenter)}
        contentClassName="gap-4"
        title={<FormattedMessage id="AccountLockoutScreen.settings.label" />}
      >
        <Toggle
          textWeight="medium"
          checked={state.isEnabled}
          text={renderToString("AccountLockoutScreen.enable.label")}
          onCheckedChange={onChangeIsEnabled}
        />
      </SettingsSectionCard>
      {state.isEnabled ? (
        <>
          <LockoutThresholdSection
            className={className}
            state={state}
            onHistoryDurationMinsChange={onHistoryDurationMinsChange}
            onMaxAttemptsChange={onMaxAttemptsChange}
          />
          <LockoutDurationSection
            className={className}
            state={state}
            onBackoffFactorChange={onBackoffFactorChange}
            onMaximumDurationMinsChange={onMaximumDurationMinsChange}
            onMinDurationChange={onMinDurationChange}
          />
          <LockoutTypeSection
            className={className}
            state={state}
            onChangeLockoutType={onChangeLockoutType}
          />
          <LockoutAuthenticatorSection
            className={className}
            state={state}
            onChangeIsEnabledForPassword={onChangeIsEnabledForPassword}
            onChangeIsEnabledForOOBOTP={onChangeIsEnabledForOOBOTP}
            onChangeIsEnabledForTOTP={onChangeIsEnabledForTOTP}
            onChangeIsEnabledForRecoveryCode={onChangeIsEnabledForRecoveryCode}
          />
        </>
      ) : null}
    </>
  );
}
