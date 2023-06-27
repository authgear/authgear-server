import React, { ReactElement, useMemo, useContext, useCallback } from "react";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { Text } from "@fluentui/react";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import { WidgetSubsection } from "./LoginMethodConfigurationScreen";
import TextField from "../../TextField";
import { AuthenticationLockoutType } from "../../types";
import produce from "immer";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./LockoutSettings.module.css";

export interface State {
  maxAttempts?: number;
  historyDurationMins?: number;
  minimumDurationMins?: number;
  maximumDuration?: number;
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
  key: "maxAttempts" | "historyDurationMins"
) {
  return useCallback(
    (_, value) => {
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

export default function LockoutSettings<T extends State>(
  props: LockoutSettingsProps<T>
): ReactElement {
  const { className, setState, ...state } = props;
  const { renderToString } = useContext(MessageContext);

  const onMaxAttemptsChange = useNumberOnChange(setState, "maxAttempts");
  const onHistoryDurationMinsChange = useNumberOnChange(
    setState,
    "historyDurationMins"
  );

  return (
    <Widget className={className}>
      <WidgetTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.lockout.title" />
      </WidgetTitle>
      <WidgetSubsection>
        <SubsectionTitle>
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.title" />
        </SubsectionTitle>
        <WidgetDescription>
          <FormattedMessage id="LoginMethodConfigurationScreen.lockout.threshold.description" />
        </WidgetDescription>
        <div>
          <TextField
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
          <TextField
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
                resetIntervalHours: Number(
                  ((state.historyDurationMins ?? 0) / 60).toFixed(2)
                ),
              }}
            />
          </Text>
        </div>
      </WidgetSubsection>
    </Widget>
  );
}
