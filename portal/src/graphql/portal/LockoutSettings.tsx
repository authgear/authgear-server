import React, { ReactElement, useMemo, useContext, useCallback } from "react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { Checkbox, Dropdown, IDropdownOption, Text } from "@fluentui/react";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import { WidgetSubsection } from "./LoginMethodConfigurationScreen";
import { AuthenticationLockoutType } from "../../types";
import produce from "immer";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./LockoutSettings.module.css";
import HorizontalDivider from "../../HorizontalDivider";
import FormTextField from "../../FormTextField";

export interface State {
  maxAttempts?: number;
  historyDurationMins?: number;
  minimumDurationMins?: number;
  maximumDurationMins?: number;
  backoffFactor?: number;
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

function SubsectionTitle(
  props: React.PropsWithChildren<Record<never, never>>
): ReactElement {
  const { children } = props;

  const styles = useMemo(
    () => ({
      root: {
        fontWeight: "600",
      },
    }),
    []
  );

  return (
    <Text as="h3" block={true} variant="mediumPlus" styles={styles}>
      {children}
    </Text>
  );
}

function useNumberOnChange<T extends State>(
  setState: LockoutSettingsProps<T>["setState"],
  key:
    | "maxAttempts"
    | "historyDurationMins"
    | "minimumDurationMins"
    | "backoffFactor"
    | "maximumDurationMins"
) {
  return useCallback(
    (_: unknown, value?: string | undefined) => {
      if (value == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev[key] = parseIntegerAllowLeadingZeros(value);
        })
      );
    },
    [setState, key]
  );
}

function useBooleanOnChange<T extends State>(
  setState: LockoutSettingsProps<T>["setState"],
  key:
    | "isEnabledForPassword"
    | "isEnabledForTOTP"
    | "isEnabledForOOBOTP"
    | "isEnabledForRecoveryCode"
) {
  return useCallback(
    (_: unknown, value?: boolean | undefined) => {
      if (value == null) {
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

function formatOptionalHour(hour: number): string {
  if (hour < 1) {
    return "none";
  }
  return Number(hour.toFixed(2)).toString();
}

function LockoutThresholdSection<T extends State>(props: {
  state: T;
  onMaxAttemptsChange: (_: unknown, value?: string | undefined) => void;
  onHistoryDurationMinsChange: (_: unknown, value?: string | undefined) => void;
}) {
  const { renderToString } = useContext(MessageContext);
  const { state, onHistoryDurationMinsChange, onMaxAttemptsChange } = props;

  return (
    <WidgetSubsection>
      <SubsectionTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.title" />
      </SubsectionTitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.description" />
      </WidgetDescription>
      <div>
        <FormTextField
          fieldName="max_attempts"
          parentJSONPointer="/authentication/lockout"
          type="text"
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.threshold.failedAttempts.title"
          )}
          value={state.maxAttempts?.toFixed(0) ?? ""}
          onChange={onMaxAttemptsChange}
        />
        <WidgetDescription className="mt-2">
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.failedAttempts.description" />
        </WidgetDescription>
      </div>
      <div>
        <FormTextField
          fieldName="history_duration"
          parentJSONPointer="/authentication/lockout"
          type="text"
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.threshold.resetAfter.title"
          )}
          value={state.historyDurationMins?.toFixed(0) ?? ""}
          onChange={onHistoryDurationMinsChange}
        />
        <WidgetDescription className="mt-2">
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.resetAfter.description" />
        </WidgetDescription>
      </div>
      <div className={styles.descriptionBox}>
        <Text variant="medium">
          <FormattedMessage
            id="LoginMethodConfigurationScreen.lockout.threshold.overall.description"
            values={{
              attempts: state.maxAttempts ?? 0,
              resetIntervalMins: state.historyDurationMins ?? 0,
              resetIntervalHours: formatOptionalHour(
                (state.historyDurationMins ?? 0) / 60
              ),
            }}
          />
        </Text>
      </div>
    </WidgetSubsection>
  );
}

function LockoutDurationSection<T extends State>(props: {
  state: T;
  onMinDurationChange: (_: unknown, value?: string | undefined) => void;
  onBackoffFactorChange: (_: unknown, value?: string | undefined) => void;
  onMaximumDurationMinsChange: (_: unknown, value?: string | undefined) => void;
}) {
  const { renderToString } = useContext(MessageContext);
  const {
    state,
    onBackoffFactorChange,
    onMaximumDurationMinsChange,
    onMinDurationChange,
  } = props;

  const overallDescriptionValues = useMemo(() => {
    const durationMins = state.minimumDurationMins ?? 0;
    const backoffFactor = state.backoffFactor ?? 1;
    const durationMinsSecond = durationMins * backoffFactor;
    const durationMinsThird = durationMins * backoffFactor * backoffFactor;
    const maxDurationMins = state.maximumDurationMins ?? 0;
    const maxDurationHours = formatOptionalHour(maxDurationMins / 60);
    return {
      durationMins,
      backoffFactor,
      durationMinsSecond,
      durationMinsThird,
      maxDurationMins,
      maxDurationHours,
    };
  }, [state]);

  return (
    <WidgetSubsection>
      <SubsectionTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.title" />
      </SubsectionTitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.description" />
      </WidgetDescription>
      <div>
        <FormTextField
          fieldName="minimum_duration"
          parentJSONPointer="/authentication/lockout"
          type="text"
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.duration.duration.title"
          )}
          value={state.minimumDurationMins?.toFixed(0) ?? ""}
          onChange={onMinDurationChange}
        />
        <WidgetDescription className="mt-2">
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.duration.description" />
        </WidgetDescription>
      </div>
      <div>
        <FormTextField
          fieldName="backoff_factor"
          parentJSONPointer="/authentication/lockout"
          type="text"
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.duration.backoff.title"
          )}
          value={state.backoffFactor?.toString() ?? ""}
          onChange={onBackoffFactorChange}
        />
        <WidgetDescription className="mt-2">
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.backoff.description" />
        </WidgetDescription>
      </div>
      <div>
        <FormTextField
          fieldName="maximum_duration"
          parentJSONPointer="/authentication/lockout"
          type="text"
          label={renderToString(
            "LoginMethodConfigurationScreen.lockout.duration.max.title"
          )}
          value={state.maximumDurationMins?.toFixed(0) ?? ""}
          onChange={onMaximumDurationMinsChange}
        />
        <WidgetDescription className="mt-2">
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.duration.max.description" />
        </WidgetDescription>
      </div>

      <div className={styles.descriptionBox}>
        <Text variant="medium">
          {(state.backoffFactor ?? 1) <= 1 ? (
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
    </WidgetSubsection>
  );
}

function LockoutTypeSection<T extends State>(props: {
  state: T;
  onChangeLockoutType: (
    _: unknown,
    option?: IDropdownOption<AuthenticationLockoutType>
  ) => void;
}) {
  const { state, onChangeLockoutType } = props;
  const { renderToString } = useContext(MessageContext);

  const lockoutTypeOptions = useMemo<
    IDropdownOption<AuthenticationLockoutType>[]
  >(() => {
    return [
      {
        key: "per_user",
        data: "per_user",
        text: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUser"
        ),
      },
      {
        key: "per_user_per_ip",
        data: "per_user_per_ip",
        text: renderToString(
          "LoginMethodConfigurationScreen.lockout.type.perUserPerIP"
        ),
      },
    ];
  }, [renderToString]);

  return (
    <WidgetSubsection>
      <SubsectionTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.type.title" />
      </SubsectionTitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.type.description" />
      </WidgetDescription>
      <Dropdown
        options={lockoutTypeOptions}
        selectedKey={state.lockoutType}
        onChange={onChangeLockoutType}
      />
    </WidgetSubsection>
  );
}

function LockoutAuthenticatorSection<T extends State>(props: {
  state: T;
  onChangeIsEnabledForPassword: (_: unknown, checked?: boolean) => void;
  onChangeIsEnabledForOOBOTP: (_: unknown, checked?: boolean) => void;
  onChangeIsEnabledForTOTP: (_: unknown, checked?: boolean) => void;
  onChangeIsEnabledForRecoveryCode: (_: unknown, checked?: boolean) => void;
}) {
  const {
    state,
    onChangeIsEnabledForPassword,
    onChangeIsEnabledForOOBOTP,
    onChangeIsEnabledForTOTP,
    onChangeIsEnabledForRecoveryCode,
  } = props;
  const { renderToString } = useContext(MessageContext);

  return (
    <WidgetSubsection>
      <SubsectionTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.authenticator.title" />
      </SubsectionTitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.authenticator.description" />
      </WidgetDescription>
      <Checkbox
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.authenticator.password"
        )}
        checked={state.isEnabledForPassword}
        onChange={onChangeIsEnabledForPassword}
      />
      <Checkbox
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.authenticator.passwordless"
        )}
        checked={state.isEnabledForOOBOTP}
        onChange={onChangeIsEnabledForOOBOTP}
      />
      <Checkbox
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.authenticator.totp"
        )}
        checked={state.isEnabledForTOTP}
        onChange={onChangeIsEnabledForTOTP}
      />
      <Checkbox
        label={renderToString(
          "LoginMethodConfigurationScreen.lockout.authenticator.recoveryCode"
        )}
        checked={state.isEnabledForRecoveryCode}
        onChange={onChangeIsEnabledForRecoveryCode}
      />
    </WidgetSubsection>
  );
}

export default function LockoutSettings<T extends State>(
  props: LockoutSettingsProps<T>
): ReactElement {
  const { className, setState, ...state } = props;

  const onMaxAttemptsChange = useNumberOnChange(setState, "maxAttempts");
  const onHistoryDurationMinsChange = useNumberOnChange(
    setState,
    "historyDurationMins"
  );
  const onMinDurationChange = useNumberOnChange(
    setState,
    "minimumDurationMins"
  );
  const onBackoffFactorChange = useNumberOnChange(setState, "backoffFactor");
  const onMaximumDurationMinsChange = useNumberOnChange(
    setState,
    "maximumDurationMins"
  );
  const onChangeLockoutType = useCallback(
    (_: unknown, option?: IDropdownOption<AuthenticationLockoutType>) => {
      if (option == null) {
        return;
      }
      const { data: newType } = option;
      if (newType == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.lockoutType = newType;
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
    <Widget className={className}>
      <WidgetTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.title" />
      </WidgetTitle>
      <LockoutThresholdSection
        state={state}
        onHistoryDurationMinsChange={onHistoryDurationMinsChange}
        onMaxAttemptsChange={onMaxAttemptsChange}
      />
      <HorizontalDivider />
      <LockoutDurationSection
        state={state}
        onBackoffFactorChange={onBackoffFactorChange}
        onMaximumDurationMinsChange={onMaximumDurationMinsChange}
        onMinDurationChange={onMinDurationChange}
      />
      <HorizontalDivider />
      <LockoutTypeSection
        state={state}
        onChangeLockoutType={onChangeLockoutType}
      />
      <HorizontalDivider />
      <LockoutAuthenticatorSection
        state={state}
        onChangeIsEnabledForPassword={onChangeIsEnabledForPassword}
        onChangeIsEnabledForOOBOTP={onChangeIsEnabledForOOBOTP}
        onChangeIsEnabledForTOTP={onChangeIsEnabledForTOTP}
        onChangeIsEnabledForRecoveryCode={onChangeIsEnabledForRecoveryCode}
      />
    </Widget>
  );
}
